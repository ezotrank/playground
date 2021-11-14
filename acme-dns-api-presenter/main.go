package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-acme/lego/v4/providers/dns"
)

var (
	addr     = flag.String("a", ":8888", "Server listen address")
	provider = flag.String("p", "cloudflare", "DNS provider name")
)

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	prov, err := dns.NewDNSChallengeProviderByName(*provider)
	if err != nil {
		log.Fatalln("Failed recognize DNS provider ", err)
	}

	srv := http.Server{
		Addr: *addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
				return
			}

			if err := r.ParseForm(); err != nil {
				log.Println("Failed parse form", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			domain := r.Form.Get("domain")
			key := r.Form.Get("key")

			switch r.RequestURI {
			case "/set":
				if err := prov.Present(domain, key, key); err != nil {
					log.Println("Failed set DNS record", err)
					http.Error(w, "Failed set DNS record", http.StatusInternalServerError)
					return
				}
			case "/del":
				if err := prov.CleanUp(domain, key, key); err != nil {
					log.Println("Failed del DNS record", err)
					http.Error(w, "Failed del DNS record", http.StatusInternalServerError)
					return
				}
			default:
				http.NotFound(w, r)
			}
		}),
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Failed ListenAndServe HTTP ", err)
	}
}
