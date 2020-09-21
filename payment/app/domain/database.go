package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

type DB interface {
	SaveTransaction(t Transaction) (*string, error)
	GetBalance(userID string) (*Balance, error)
	Lock(t Transaction, conn *sql.Conn) (int, error)
	Unlock(keyStr int, conn *sql.Conn) error
	GetAllTransactions(request GetTransactions) ([]Transaction, error)
}

type Transaction struct {
	RequestID     string    `json:"request_id"`
	TransactionID string    `json:"transaction_id"`
	SenderID      string    `json:"sender_id"`
	RecipientID   string    `json:"recipient_id"`
	Message       string    `json:"message"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

type GetTransactions struct {
	UserID string `json:"user_id"`
}

func (t Transaction) String() string {
	return fmt.Sprintf("%s-%s-%s-%f-%s-%s", t.RequestID, t.SenderID, t.RecipientID, t.Amount, t.Currency, t.Message)
}

type Balance struct {
	Amount          float64 `json:"amount"`
	LastTransaction *string `json:"last_transaction"`
}

type SQLDatabase struct {
	db *sql.DB
}

func NewSQLDatabase(db *sql.DB) *SQLDatabase {
	return &SQLDatabase{db: db}
}

func (s *SQLDatabase) SaveTransaction(t Transaction) (*string, error) {
	conn, err := s.db.Conn(context.Background())
	if err != nil {
		return nil, err
	}

	// mutex lock
	keyMutex, err := s.Lock(t, conn)
	if err != nil {
		return nil, err
	}

	// defer unlock
	defer func() {
		err := s.Unlock(keyMutex, conn)
		log.Error().Err(err).Msg("error releasing the lock")
	}()

	senderBalance, err := s.GetBalance(t.SenderID)
	if err != nil {
		return nil, err
	}

	err = checkTransaction(senderBalance.Amount, t.Amount)
	if err != nil {
		return nil, err
	}

	recipientBalance, err := s.GetBalance(t.RecipientID)
	if err != nil {
		return nil, err
	}

	// generate uuid
	txID := uuid.NewV4().String()

	// begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	// save transaction record
	err = saveTransaction(tx, t, txID)
	if err != nil {
		return nil, err
	}

	// update sender balance
	err = updateBalance(tx, senderBalance.Amount-t.Amount, t.SenderID, txID)
	if err != nil {
		return nil, err
	}

	// update receiver balance
	err = updateBalance(tx, recipientBalance.Amount+t.Amount, t.RecipientID, txID)
	if err != nil {
		return nil, err
	}

	// commit
	return &txID, tx.Commit()
}

func (s *SQLDatabase) GetBalance(userID string) (*Balance, error) {
	balance := Balance{}
	query := "SELECT amount, lastTransactionId FROM balance WHERE userid = $1"

	err := s.db.QueryRow(query, userID).Scan(&balance.Amount, &balance.LastTransaction)
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}

	return &balance, nil
}

func checkTransaction(balance, withdrawal float64) error {
	if withdrawal < 0 {
		return errors.New("amount is negative")
	}

	if balance-withdrawal < 0 {
		return errors.New("insufficient balance")
	}

	return nil
}

func (s *SQLDatabase) Lock(t Transaction, conn *sql.Conn) (int, error) {
	key := []string{t.RecipientID, t.SenderID}

	sort.Strings(key)

	keyStr := fmt.Sprint(key)
	h := fnv.New32a()
	_, err := h.Write([]byte(keyStr))

	if err != nil {
		return 0, err
	}
	hash := h.Sum32()

	_, err = conn.ExecContext(context.Background(), `SELECT pg_advisory_lock($1)`, hash)
	if err != nil {
		return 0, err
	}

	return int(hash), err
}

func (s *SQLDatabase) Unlock(keyStr int, conn *sql.Conn) error {
	_, err := conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, keyStr)
	return err
}

func saveTransaction(tx *sql.Tx, t Transaction, txID string) error {
	query := `INSERT into transactions  (requestid, transactionid, senderid, receiverid, amount, currency, message )
 			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tx.Exec(query, t.RequestID, txID, t.SenderID, t.RecipientID, t.Amount, t.Currency, t.Message)
	if err != nil {
		log.Error().Err(err)
		tx.Rollback()
	}
	return err
}

func updateBalance(tx *sql.Tx, amount float64, userID string, txID string) error {
	_, err := tx.Exec("UPDATE balance SET Amount = $1, lastTransactionId = $2 WHERE userId = $3 ", amount, txID, userID)
	if err != nil {
		log.Error().Err(err)
		tx.Rollback()
	}
	return err
}

func (s *SQLDatabase) GetAllTransactions(request GetTransactions) ([]Transaction, error) {
	userSQL := "SELECT requestid, transactionid, senderid, receiverid, amount, currency, createdat FROM transactions WHERE senderid = $1 OR receiverid = $1"

	rows, err := s.db.Query(userSQL, request.UserID)
	if err != nil {
		log.Error().Msgf("Failed to execute query: %s", err)
	}

	if err != nil {
		log.Error().Err(err)
		return nil, err
	}

	defer rows.Close()

	var t []Transaction
	var requestID, transactionID, currency, msg string
	var senderID, recipientID string
	var amount float64
	var createdAt time.Time

	for rows.Next() {
		err := rows.Scan(&requestID, &transactionID, &senderID, &recipientID, &amount, &currency, &msg, &createdAt)
		if err != nil {
			log.Error().Err(err)
			return nil, err
		}

		t = append(t, Transaction{
			RequestID:     requestID,
			TransactionID: transactionID,
			SenderID:      senderID,
			RecipientID:   recipientID,
			Amount:        amount,
			Message:       msg,
			Currency:      currency,
			CreatedAt:     createdAt,
		})
	}

	err = rows.Err()
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}
	return t, nil
}
