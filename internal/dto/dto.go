package dto

import (
	"errors"
	"strings"
)

// CreateTransactionRequest is the DTO received from the HTTP API.
type CreateTransactionRequest struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Type      string  `json:"type"` // expected: "credit" or "debit"
}

// Validate performs simple, deterministic validation on the DTO.
func (r CreateTransactionRequest) Validate() error {
	if strings.TrimSpace(r.AccountID) == "" {
		return errors.New("account_id is required")
	}
	if r.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if len(strings.TrimSpace(r.Currency)) != 3 {
		return errors.New("currency must be a 3-letter code")
	}
	t := strings.ToLower(strings.TrimSpace(r.Type))
	if t != "credit" && t != "debit" {
		return errors.New("type must be 'credit' or 'debit'")
	}
	return nil
}
