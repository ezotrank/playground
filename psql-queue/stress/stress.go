package main

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	threads = 10
	addr    = "http://localhost:23123"
)

func main() {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	pool := make(chan struct{}, threads)
	for {
		pool <- struct{}{}
		go func() {
			res, err := client.Get(addr + "/set?key=" + uuid.New().String())
			if err != nil || res.StatusCode != http.StatusOK {
				log.Fatalln("failed set key")
			}

			<-pool
		}()
	}
}
