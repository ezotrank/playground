package main

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	cert := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("./cache"),
	}

	srv := http.Server{
		Addr: ":8443",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintln(w, "Hello HTTPS")
		}),
	}
	srv.TLSConfig = cert.TLSConfig()

	httpSrv := http.Server{
		Addr:    ":8080",
		Handler: cert.HTTPHandler(nil),
	}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	if err := srv.ListenAndServeTLS("", ""); err != nil {
		panic(err)
	}
}
