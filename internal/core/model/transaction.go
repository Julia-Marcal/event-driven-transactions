package model

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeCredit TransactionType = "credit"
	TransactionTypeDebit  TransactionType = "debit"
)

// Transaction is the domain model for a financial transaction request.
type Transaction struct {
	ID        uuid.UUID       `json:"transaction_id"`
	AccountID string          `json:"account_id"`
	Amount    float64         `json:"amount"`
	Currency  string          `json:"currency"`
	Type      TransactionType `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
}

// TransactionEvent is the event published to the message broker.
type TransactionEvent = Transaction
