package main

import (
	"log"
	"net/http"	
	"encoding/json"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	addr := ":8080"
	log.Println("Starting server on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))

	
}