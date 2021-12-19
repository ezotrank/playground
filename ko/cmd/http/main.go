package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "World!\n")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
