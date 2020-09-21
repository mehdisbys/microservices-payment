//!build integration

package domain

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432 // custom port to avoid clashing with non-test container
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

// Make sure a postgres instance is running or these tests will fail

var db *sql.DB

func TestMain(m *testing.M) {
	setup()
	m.Run()
	teardown()
}

func setup() {
	var err error

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)

	err = db.Ping()
	if err != nil {
		panic(err)
	}
}

func teardown() {
	db.Close()
}

func cleanDB(db *sql.DB) {
	query := `DELETE from balance WHERE id > 0`

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	query = `DELETE from transactions WHERE id > 0`

	_, err = db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func initBalance(userID string, amount float64) error {
	query := `INSERT into balance (userid, Amount ) VALUES ($1, $2) 
			  ON CONFLICT (userid) DO UPDATE SET Amount=$2 WHERE balance.userid = $1`

	_, err := db.Exec(query, userID, amount)
	return err
}

func TestDatabase(t *testing.T) {
	pay := NewSQLDatabase(db)
	cleanDB(db)

	tests := []struct {
		name                     string
		t                        Transaction
		initialBalance           float64
		senderExpectedBalance    float64
		recipientExpectedBalance float64
		expectErr                bool
	}{
		{
			name: "Happy path send",
			t: Transaction{
				RequestID:   "123",
				SenderID:    "1",
				RecipientID: "2",
				Amount:      50,
				Currency:    "SGD",
				Message:     "Ref: abc",
			},
			initialBalance:           51,
			senderExpectedBalance:    1,
			recipientExpectedBalance: 50,
		},
		{
			name: "Insufficient funds",
			t: Transaction{
				RequestID:     "123",
				TransactionID: "1",
				SenderID:      "1",
				RecipientID:   "2",
				Amount:        50,
			},
			initialBalance:           49,
			expectErr:                true,
			senderExpectedBalance:    49,
			recipientExpectedBalance: 0,
		},
		{
			name: "Sending negative Amount",
			t: Transaction{
				RequestID:     "123",
				TransactionID: "1",
				SenderID:      "1",
				RecipientID:   "2",
				Amount:        -10,
			},
			initialBalance:           49,
			expectErr:                true,
			senderExpectedBalance:    49,
			recipientExpectedBalance: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			err := initBalance(test.t.SenderID, test.initialBalance)
			if err != nil {
				t.Fatal(err)
			}

			err = initBalance(test.t.RecipientID, 0)
			if err != nil {
				t.Fatal(err)
			}

			_, err = pay.SaveTransaction(test.t)

			if (err == nil) && test.expectErr {
				t.Fatalf("expected error to occur but got none")
			}

			if (err != nil) && !test.expectErr {
				t.Fatalf("expected to have no errors but got : %s", err)
			}

			senderBalance, err := pay.GetBalance(test.t.SenderID)
			if err != nil {
				t.Fatal(err)
			}

			if senderBalance.Amount != test.senderExpectedBalance {
				t.Fatalf("expected sender balance to be %f got %f", test.senderExpectedBalance, senderBalance.Amount)
			}

			recipientBalance, err := pay.GetBalance(test.t.RecipientID)
			if err != nil {
				t.Fatal(err)
			}

			if recipientBalance.Amount != test.recipientExpectedBalance {
				t.Fatalf("expected recipient balance to be %f got %f", test.recipientExpectedBalance, recipientBalance.Amount)
			}
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	cleanDB(db)

	n := 50 // goroutines

	wg := sync.WaitGroup{}
	pay := NewSQLDatabase(db)

	err := initBalance("1", 100)
	if err != nil {
		t.Fatal(err)
	}

	err = initBalance("2", 0)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < n; i++ {
		i := i
		go func() {
			wg.Add(1)
			defer wg.Done()

			_, err := pay.SaveTransaction(Transaction{
				RequestID:     fmt.Sprint(i),
				TransactionID: fmt.Sprint(i),
				SenderID:      "1",
				RecipientID:   "2",
				Amount:        2,
				Currency:      "SGD",
			})
			if err != nil {
				panic(err)
			}
		}()
	}

	wg.Wait()

	senderBalance, err := pay.GetBalance("1")
	if err != nil {
		t.Fatal(err)
	}

	if senderBalance.Amount != 0 {
		t.Fatalf("expected sender balance to be %f got %f", 0.0, senderBalance.Amount)
	}

	recipientBalance, err := pay.GetBalance("2")
	if err != nil {
		t.Fatal(err)
	}

	if recipientBalance.Amount != 100 {
		t.Fatalf("expected recipient balance to be %f got %f", 100.0, recipientBalance.Amount)
	}
}

func TestSQLDatabase_GetAllTransactions(t *testing.T) {
	senderID := "1"
	recipientID := "2"

	pay := NewSQLDatabase(db)
	cleanDB(db)

	err := initBalance(senderID, 100)
	if err != nil {
		t.Fatal(err)
	}

	err = initBalance(recipientID, 100)
	if err != nil {
		t.Fatal(err)
	}

	txs := []Transaction{
		{
			RequestID:   "456",
			SenderID:    "2",
			RecipientID: "1",
			Amount:      50,
			Currency:    "SGD",
			Message:     "Ref: abc",
		},
		{
			RequestID:   "123",
			SenderID:    "1",
			RecipientID: "2",
			Amount:      30,
			Currency:    "SGD",
			Message:     "Ref: xyz",
		},
	}

	_, err = pay.SaveTransaction(txs[0])
	if err != nil {
		t.Fatal(err)
	}

	_, err = pay.SaveTransaction(txs[1])
	if err != nil {
		t.Fatal(err)
	}

	results, err := pay.GetAllTransactions(GetTransactions{UserID: "1"})
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range results {
		for _, tx := range txs {
			if tx.String() == r.String() {
				found = true
			}
		}
		if !found {
			t.Fatalf("%v not found in returned results", r)
		}
	}

}
