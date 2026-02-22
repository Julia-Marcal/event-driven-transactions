package config

import (
	"os"
)

type AppConfig struct {
	Port        string
	RabbitMQURL string
}

func Load() AppConfig {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	return AppConfig{
		Port:        port,
		RabbitMQURL: url,
	}
}
