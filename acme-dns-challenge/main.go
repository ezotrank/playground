package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/net/publicsuffix"
)

const (
	LEDirectoryProduction = "https://acme-v02.api.letsencrypt.org/directory"
	LEDirectoryStaging    = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

var (
	domains = flag.String("d", "", "Domain identifiers, separated by the comma.")
	api     = flag.String("u", "", "DNS api service URL.")
	addr    = flag.String("a", ":8443", "HTTPS server listen address.")
	email   = flag.String("e", "", "Email address for registration.")
	staging = flag.Bool("s", false, "Use staging letsencrypt api.")
)

func fillCert(ctx context.Context, domains []string, email, durl, leurl string, cert *tls.Certificate) error {
	accountKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed generate account key: %w", err)
	}

	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: leurl,
	}

	account := &acme.Account{}
	if len(email) > 0 {
		account.Contact = []string{"mailto:" + email}
	}
	if _, err := client.Register(ctx, account, acme.AcceptTOS); err != nil {
		return fmt.Errorf("failed register ACME account: %w", err)
	}

	order, err := client.AuthorizeOrder(ctx, acme.DomainIDs(domains...))
	if err != nil {
		return fmt.Errorf("failed authorize order: %w", err)
	}

	authzs := make([]*acme.Authorization, 0)
	afterAuthHooks := make([]func() error, 0)
	defer func() {
		for _, hook := range afterAuthHooks {
			if err := hook(); err != nil {
				log.Println("failed after auth hook ", err)
			}
		}
	}()

	for _, u := range order.AuthzURLs {
		authz, err := client.GetAuthorization(ctx, u)
		if err != nil {
			return fmt.Errorf("failed get authorization: %w", err)
		}
		authzs = append(authzs, authz)

		var chal *acme.Challenge
		for _, c := range authz.Challenges {
			if c.Type == "dns-01" {
				chal = c
				break
			}
		}
		if chal == nil {
			return fmt.Errorf("no DNS-01 challenge found")
		}

		txtLabel := "_acme-challenge." + authz.Identifier.Value
		txtValue, err := client.DNS01ChallengeRecord(chal.Token)
		if err != nil {
			return fmt.Errorf("failed DNS-01 challenge record: %w", err)
		}

		ka, err := keyAuth(client.Key.Public(), chal.Token)
		if err != nil {
			return fmt.Errorf("failed key auth: %w", err)
		}

		if err := setDNSRecord(durl, authz.Identifier.Value, ka); err != nil {
			return fmt.Errorf("failed set DNS record: %w", err)
		}

		afterAuthHooks = append(afterAuthHooks, func() error {
			return delDNSRecord(durl, authz.Identifier.Value, ka)
		})

		if err := waitPropagation(txtLabel, txtValue); err != nil {
			return fmt.Errorf("failed wait propagation: %w", err)
		}

		if _, err = client.Accept(ctx, chal); err != nil {
			return fmt.Errorf("failed accept challenge: %w", err)
		}

		if _, err = client.WaitAuthorization(ctx, authz.URI); err != nil {
			return fmt.Errorf("failed wait authorization: %w", err)
		}
	}

	if _, err := client.WaitOrder(ctx, order.URI); err != nil {
		return fmt.Errorf("failed wait order: %w", err)
	}

	authzIDs := make([]acme.AuthzID, len(authzs))
	for i := range authzs {
		authzIDs[i] = authzs[i].Identifier
	}

	csr, certkey := newCSR(domains)
	der, _, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return fmt.Errorf("failed create order cert: %w", err)
	}

	leaf, err := x509.ParseCertificate(der[0])
	if err != nil {
		return fmt.Errorf("failed parse certificate: %w", err)
	}

	// There's a race condition, but I can live with that.
	*cert = tls.Certificate{
		PrivateKey:  certkey,
		Certificate: der,
		Leaf:        leaf,
	}
	go time.AfterFunc(next(cert.Leaf.NotAfter), func() {
		for {
			log.Println("Renewing certificate...")
			err := fillCert(ctx, domains, email, durl, leurl, cert)
			if err == nil {
				break
			}
			log.Println("Failed create certificate ", err)
			time.Sleep(time.Minute)
		}
	})
	return nil
}

func next(expiry time.Time) time.Duration {
	d := expiry.Sub(time.Now()) - 720*time.Hour // 30 days
	if d < 0 {
		return 0
	}
	return d
}

func GetGetCertificate(ctx context.Context, domains []string, email, durl, leurl string) func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var cert tls.Certificate
	if err := fillCert(ctx, domains, email, durl, leurl, &cert); err != nil {
		log.Fatal("Failed fill certificate ", err)
	}

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return &cert, nil
	}
}

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	domains := strings.Split(*domains, ",")
	u := LEDirectoryProduction
	if *staging {
		u = LEDirectoryStaging
	}
	srv := http.Server{
		Addr: *addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintln(w, "Hello HTTPS")
		}),
		TLSConfig: &tls.Config{
			GetCertificate: GetGetCertificate(context.Background(), domains, *email, *api, u),
		},
	}

	log.Println("Start listening on ", *addr)
	if err := srv.ListenAndServeTLS("", ""); err != nil {
		log.Fatal("Failed ListenAndServer HTTPS ", err)
	}
}

func waitPropagation(label, value string) interface{} {
	etldplus, err := publicsuffix.EffectiveTLDPlusOne(label)
	if err != nil {
		log.Fatal("Failed get effective TLD+1: ", err)
	}

	nameservers, err := net.LookupNS(etldplus)
	if err != nil {
		log.Fatal("Failed lookup nameservers: ", err)
	}

	resolver := net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			var err error
			var conn net.Conn
			for _, srv := range nameservers {
				if conn, err = d.DialContext(ctx, network, srv.Host+":53"); err == nil {
					return conn, nil
				}
			}
			return nil, fmt.Errorf("failed to connect to any nameserver: %w", err)
		},
	}

	// Two minutes ought to be enough for anybody.
	for i := 0; i < 120; i++ {
		time.Sleep(time.Second)
		if records, err := resolver.LookupTXT(context.Background(), label); err == nil {
			for _, txt := range records {
				if txt == value {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("failed propoganation domain")
}

func delDNSRecord(u string, domain string, value string) error {
	values := url.Values{
		"domain": {domain},
		"key":    {value},
	}

	resp, err := http.PostForm(u+"/del", values)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed delete DNS record: %s", resp.Status)
	}

	return nil
}

func setDNSRecord(u string, domain string, value string) error {
	values := url.Values{
		"domain": {domain},
		"key":    {value},
	}

	resp, err := http.PostForm(u+"/set", values)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed set DNS record: %s", resp.Status)
	}

	return nil
}

func newCSR(names []string) ([]byte, crypto.Signer) {
	var csr x509.CertificateRequest
	for _, name := range names {
		csr.DNSNames = append(csr.DNSNames, name)
	}
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("newCSR: ecdsa.GenerateKey for a cert: %v", err))
	}
	b, err := x509.CreateCertificateRequest(rand.Reader, &csr, k)
	if err != nil {
		panic(fmt.Sprintf("newCSR: x509.CreateCertificateRequest: %v", err))
	}
	return b, k
}

func keyAuth(pub crypto.PublicKey, token string) (string, error) {
	th, err := acme.JWKThumbprint(pub)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", token, th), nil
}
