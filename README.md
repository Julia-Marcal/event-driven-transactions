# event-driven-transactions

A Go service that exposes an HTTP API to create financial transactions, persists them to MongoDB, and publishes events to RabbitMQ. A built-in consumer listens to the same queue and processes incoming messages — enabling an event-driven flow end-to-end.

## Architecture

```
HTTP Request
    │
    ▼
[Gin Handler]
    │  validates & parses DTO
    ▼
[TransactionService]
    │  checks idempotency key
    │  persists to MongoDB
    ▼
[RabbitMQ Publisher]
    │  publishes JSON event
    ▼
[RabbitMQ Consumer]
    │  processes message (manual ack/nack + retry)
    ▼
[MongoDB]  ←── transaction + idempotency_key stored
```

**Packages:**

| Path | Responsibility |
|------|---------------|
| `internal/core/model` | Domain model (`Transaction`, `TransactionEvent`) |
| `internal/core/service` | Application logic (`CreateAndPublish`) |
| `internal/dto` | Request DTO and validation |
| `internal/handler` | Gin HTTP handler |
| `internal/infrastructure/rabbitmq` | Publisher and Consumer |
| `internal/infrastructure/mongodb` | Connection and transaction persistence |
| `internal/infrastructure/config` | Env-based config loading |

## Features

- HTTP API (Gin): `POST /api/v1/transactions`
- Publishes JSON transaction events to RabbitMQ
- Persists transactions to MongoDB (`ledger` database)
- Idempotency key support — duplicate requests return `200 already processed`
- RabbitMQ consumer with manual ack, nack, and requeue on error
- Graceful consumer shutdown
- Input validation (account_id, amount > 0, type must be `credit` or `debit`)

## Setup

1. Start RabbitMQ:
   ```sh
   docker-compose up -d
   ```

2. Start a MongoDB instance (e.g. via Docker):
   ```sh
   docker run -d -p 27017:27017 --name mongodb mongo:7
   ```

3. Run the service:
   ```sh
   go run ./...
   ```

4. Open the RabbitMQ management UI:
   - [http://localhost:15672](http://localhost:15672)
   - Username: `guest` / Password: `guest`

## API

### `POST /api/v1/transactions`

Creates and publishes a transaction event.

**Request:**
```json
{
  "account_id": "acc1",
  "amount": 100,
  "type": "credit",
  "idempotency_key": "unique-key-123"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `account_id` | yes | Non-empty string |
| `amount` | yes | Positive number |
| `type` | yes | `"credit"` or `"debit"` |
| `idempotency_key` | no | Unique key to prevent duplicate processing |

**Responses:**

| Status | Body |
|--------|------|
| `200 OK` | Transaction event JSON |
| `200 OK` | `{"message": "already processed"}` (duplicate idempotency key) |
| `400 Bad Request` | `{"error": "..."}` |
| `500 Internal Server Error` | `{"error": "internal error"}` |

**Example:**
```sh
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{"account_id":"acc1","amount":100,"type":"credit","idempotency_key":"key-001"}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ connection URL |
| `MONGO_URI` | — | Full MongoDB connection URI (takes precedence) |
| `MONGO_HOST` | `mongodb:27017` | MongoDB host (used if `MONGO_URI` is not set) |
| `MONGO_INITDB_ROOT_USERNAME` | — | MongoDB username |
| `MONGO_INITDB_ROOT_PASSWORD` | — | MongoDB password |
| `MONGO_INITDB_DATABASE` | `admin` | Auth database |

## Testing

```sh
go test ./...
```
