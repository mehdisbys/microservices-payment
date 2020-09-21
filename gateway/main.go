package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/heetch/MehdiSouilhed-technical-test/gateway/app/domain"
)

func main() {
	client := &http.Client{Timeout: 5 * time.Second}
	auth := domain.NewAuth(client)
	handler, err := domain.NewRequestHandler(client, mux.NewRouter(), auth)
	if err != nil {
		panic(err)
	}

	configFile, err := domain.ParseFileConfig("config.yaml")

	if err != nil {
		log.Print(err)
		os.Exit(2)
	}

	handler.Gateway(configFile)
	log.Println("Listening on port 80")
	log.Fatal(http.ListenAndServe(":80", handler.GetRouter()))
}
