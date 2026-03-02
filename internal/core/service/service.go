package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
	"github.com/Julia-Marcal/event-driven-transactions/internal/dto"
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
		Amount:    req.Amount,
		Type:      model.TransactionType(req.Type),
		CreatedAt: time.Now(),
	}

	if s.Logger != nil {
		s.Logger.Printf("Created transaction: %+v", e)
	}

	return e, nil
}
