// Package router configures HTTP routes and server settings.
package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Julia-Marcal/event-driven-transactions/internal/handler"
	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func Router() *http.Server {
	router := gin.Default()

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/transactions", handler.CreateTransaction)
		}
	}

	cfg := config.Load()

	return &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
