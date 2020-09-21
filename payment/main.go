package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"github.com/heetch/MehdiSouilhed-technical-test/payment/app/domain"
	"github.com/heetch/MehdiSouilhed-technical-test/payment/app/handlers"
)

const (
	host     = "payment-db"
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

// Make sure a postgres instance is running or these tests will fail

func main() {
	r := mux.NewRouter()

	var err error

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	sqlDB := domain.NewSQLDatabase(db)

	handler := handlers.NewRequestHandler(sqlDB)

	r.HandleFunc("/pay_user", handler.PayUser).Methods(http.MethodPost)
	r.HandleFunc("/get_transactions", handler.GetTransactions).Methods(http.MethodPost)

	log.Print("Listening on port 80")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 80), r))

}
