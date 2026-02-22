package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
	"github.com/Julia-Marcal/event-driven-transactions/internal/core/repository"
	"github.com/Julia-Marcal/event-driven-transactions/internal/dto"
	"github.com/google/uuid"
)

type TransactionService struct {
	pub    repository.Publisher
	logger *log.Logger
}

func (s *TransactionService) Log(format string, v ...interface{}) {
	if s != nil && s.logger != nil {
		s.logger.Printf(format, v...)
	}
}

func NewTransactionService(p repository.Publisher, logger *log.Logger) *TransactionService {
	return &TransactionService{pub: p, logger: logger}
}

func (s *TransactionService) CreateAndPublish(ctx context.Context, req dto.CreateTransactionRequest) (model.TransactionEvent, error) {
	if req.Amount <= 0 {
		return model.TransactionEvent{}, errors.New("amount must be greater than zero")
	}

	e := model.TransactionEvent{
		ID:        uuid.New(),
		AccountID: req.AccountID,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Type:      model.TransactionType(req.Type),
		CreatedAt: time.Now(),
	}

	if err := s.pub.Publish(ctx, e); err != nil {
		return model.TransactionEvent{}, err
	}
	return e, nil
}
