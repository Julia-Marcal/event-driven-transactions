package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
	"github.com/Julia-Marcal/event-driven-transactions/internal/dto"
	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/mongodb"
	"github.com/google/uuid"
)

type TransactionService struct {
	Logger *log.Logger
}

func (s *TransactionService) Log(format string, v ...interface{}) {
	if s != nil && s.Logger != nil {
		s.Logger.Printf(format, v...)
	}
}

func (s *TransactionService) CreateAndPublish(ctx context.Context, req dto.CreateTransactionRequest) (model.TransactionEvent, error) {
	if req.Amount <= 0 {
		return model.TransactionEvent{}, errors.New("amount must be greater than zero")
	}

	e := model.TransactionEvent{
		ID:        uuid.New(),
		AccountID: req.AccountID,
		Amount:    float64(int64(req.Amount)),
		Type:      model.TransactionType(req.Type),
		CreatedAt: time.Now(),
	}

	if err := mongodb.ProcessTransaction(ctx, e, req.IdempotencyKey); err != nil {
		switch {
		case errors.Is(err, mongodb.ErrDuplicateIdempotencyKey):
			s.Log("Duplicate idempotency key: %s", req.IdempotencyKey)
		case errors.Is(err, mongodb.ErrAccountNotFound):
			s.Log("Account not found: %s", req.AccountID)
		case errors.Is(err, mongodb.ErrInsufficientFunds):
			s.Log("Insufficient funds for account %s: balance too low for amount %.2f", req.AccountID, req.Amount)
		default:
			s.Log("Failed to process transaction: %v", err)
		}
		return model.TransactionEvent{}, err
	}

	s.Log("Transaction processed: id=%s account=%s type=%s amount=%.2f", e.ID, e.AccountID, e.Type, e.Amount)
	return e, nil
}
