# event-driven-transactions

Small publisher service that validates transaction requests and publishes events to RabbitMQ.

Features
- Clean architecture: domain, application, transport, infrastructure
- HTTP API: POST /transactions
- Publishes JSON events to RabbitMQ
- Table-driven tests for request validation


## Setup

1. Build and start services:
	 ```sh
	 docker-compose up --build
	 ```

2. Open RabbitMQ management UI:
	 - [http://localhost:15672](http://localhost:15672)
	 - Username: guest
	 - Password: guest

3. Try the API endpoint:
	 ```sh
	 curl -X POST http://localhost:8080/api/v1/transactions \
		 -d '{"account_id":"acc1","amount":10,"currency":"USD","type":"credit"}' \
		 -H "Content-Type: application/json"
	 ```

## API

- `POST /api/v1/transactions` — create and publish a transaction event
	- Request JSON:
		```json
		{
			"account_id": "acc1",
			"amount": 10,
			"currency": "USD",
			"type": "credit"
		}
		```
	- Response:
		- 200 OK: `{ ...event... }`
		- 400 Bad Request: `{ "error": "..." }`

## Configuration

- `PORT` — HTTP server port (default: 8080)
- `RABBITMQ_URL` — RabbitMQ connection URL (default: amqp://guest:guest@rabbitmq:5672/)

## Testing

Run unit tests:
```sh
go test ./...
```
