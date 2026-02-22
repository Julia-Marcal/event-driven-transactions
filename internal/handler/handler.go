// Package handler implements HTTP handlers for transaction APIs.
package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/service"
	"github.com/Julia-Marcal/event-driven-transactions/internal/dto"
	"github.com/gin-gonic/gin"
)

var svc *service.TransactionService

func Init(s *service.TransactionService) {
	svc = s
}

// CreateTransaction handles POST /transactions
func CreateTransaction(c *gin.Context) {
	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	defer c.Request.Body.Close()

	var req dto.CreateTransactionRequest
	if err := json.Unmarshal(b, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	ctx := c.Request.Context()
	event, err := svc.CreateAndPublish(ctx, req)
	if err != nil {
		switch err.Error() {
		case "amount must be greater than zero", "currency must be a 3-letter code", "account_id is required", "type must be 'credit' or 'debit'":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		svc.Log("failed to create/publish transaction: %v", err)
		return
	}

	c.JSON(http.StatusOK, event)

	c.Status(http.StatusAccepted)
}
