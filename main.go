package main

import (
	"log"
	"net/http"

	"github.com/charlinchui/galliard/server"
	"github.com/charlinchui/galliard/transport"
)

func main() {
	srv := server.NewServer()
	handler := transport.NewHTTPHandler(srv)
	http.Handle("/bayeux", handler)
	log.Println("Galliard Bayeux server running on :8080/bayeux")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
