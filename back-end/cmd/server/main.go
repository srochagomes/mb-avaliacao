package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/srochagomes/mb-avaliacao/internal/app"
	"github.com/srochagomes/mb-avaliacao/internal/httpapi"
	"github.com/srochagomes/mb-avaliacao/internal/memory"
)

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}
	if addr[0] != ':' {
		addr = ":" + addr
	}
	svc := app.New(memory.New())
	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.Handler(svc),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("api listening on", server.Addr)
	log.Fatal(server.ListenAndServe())
}
