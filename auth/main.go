package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/heetch/MehdiSouilhed-technical-test/auth/auth/handlers"
)

func main() {
	r := mux.NewRouter()

	//_, err := domain.NewConfig("config.yaml")
	//if err != nil {
	//	panic(err)
	//}

	handler := handlers.NewRequestHandler(map[string]string{"1": "h56Zf2gRZBGTxi5iortR"})

	r.HandleFunc("/authenticate", handler.Authenticate).Methods(http.MethodPost)

	log.Print("Listening on port 80")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 80), r))
}
