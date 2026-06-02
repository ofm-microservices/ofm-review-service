# OFM Review Service

## Purpose

`ofm-review-service` owns review data.
It is the source of truth for the review write model and the owner of its
read-model projection. It validates buyer-authenticated writes against the
completed order snapshot.

Current responsibilities:

- persist review data in YugabyteDB
- project read-model data to Redis
- expose the buyer review write RPC

## Run

Local process:

```bash
cp .env.example .env
just run
```

Direct Go command:

```bash
set -a && source .env && set +a && go run ./cmd/review-service
```

Docker stack from the shared infra repo:

```bash
cd ../ofm-infra
just infra-up
```

## Environment

```env
APP_ENV=local
LOG_LEVEL=info

DB_HOST=127.0.0.1
DB_PORT=5433
DB_USER=yugabyte
DB_PASSWORD=yugabyte
DB_NAME=review_service
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=20
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m

MIGRATIONS_PATH=file://migration/yugabyte
MIGRATIONS_TABLE=schema_migrations_review_service

REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

NATS_URL=nats://127.0.0.1:4222
NATS_USER=
NATS_PASSWORD=
NATS_STREAM_REVIEW_EVENTS=REVIEW_EVENTS
NATS_STREAM_SAGA_COMMANDS=SAGA_REVIEW_COMMANDS
NATS_SUBJECT_REVIEW_GIG_PROJECTION_REQUESTED=review.projection.gig
NATS_SUBJECT_REVIEW_USER_PROJECTION_REQUESTED=review.projection.user
NATS_SUBJECT_REVIEW_GIG_RATING_REQUESTED=review.rating.gig
NATS_SUBJECT_REVIEW_SELLER_RATING_REQUESTED=review.rating.seller
NATS_SUBJECT_REVIEW_AUTHOR_RETRY_REQUESTED=review.author.retry.requested
NATS_SUBJECT_REVIEW_AVATAR_RETRY_REQUESTED=review.avatar.retry.requested
NATS_DURABLE_REVIEW_GIG_RATING=review_service_gig_rating
NATS_DURABLE_REVIEW_SELLER_RATING=review_service_seller_rating
NATS_SUBJECT_SAGA_CREATE_USER=saga.review.create
NATS_SUBJECT_SAGA_DELETE_USER=saga.review.delete
NATS_SUBJECT_SAGA_CREATE_USER_RESULT=saga.review.create.result
NATS_SUBJECT_SAGA_DELETE_USER_RESULT=saga.review.delete.result
NATS_DURABLE_SAGA_CREATE_USER=review_service_saga_create
NATS_DURABLE_SAGA_DELETE_USER=review_service_saga_delete
NATS_SAGA_BATCH_SIZE=32
NATS_SAGA_MAX_WAIT=10ms
NATS_SAGA_WORKERS=8
NATS_SAGA_QUEUE_SIZE=500
NATS_SAGA_ACK_WAIT=30s
NATS_SAGA_MAX_DELIVER=5
NATS_SAGA_ADAPTIVE_ENABLED=false
NATS_SAGA_ADAPTIVE_CHECK_INTERVAL=2s
NATS_SAGA_ADAPTIVE_MEDIUM_PENDING=200
NATS_SAGA_ADAPTIVE_HIGH_PENDING=1000
NATS_SAGA_ADAPTIVE_LOW_BATCH_SIZE=8
NATS_SAGA_ADAPTIVE_LOW_MAX_WAIT=25ms
NATS_SAGA_ADAPTIVE_MEDIUM_BATCH_SIZE=32
NATS_SAGA_ADAPTIVE_MEDIUM_MAX_WAIT=10ms
NATS_SAGA_ADAPTIVE_HIGH_BATCH_SIZE=128
NATS_SAGA_ADAPTIVE_HIGH_MAX_WAIT=2ms
```

## Technologies

Core runtime:

- Go
- YugabyteDB for the write model
- Redis for the review read model
- NATS JetStream for saga command/result transport
- Uber Fx for wiring
- Zap for structured logging
- SQL migrations via `golang-migrate`

Main libraries from `go.mod`:

- `github.com/jmoiron/sqlx`
- `github.com/redis/go-redis/v9`
- `github.com/jackc/pgx/v5`
- `github.com/golang-migrate/migrate/v4`
- `github.com/nats-io/nats.go`
- `go.uber.org/fx`
- `go.uber.org/zap`

## Architecture Notes

- `internal/domain` defines review entities and domain errors
- `internal/application` owns review use cases
- `internal/infra/write/yugabyte` owns write-side persistence
- `internal/infra/read/redis` owns the read model
- `internal/presentation/grpc` owns the review write RPC
- `migration/yugabyte` contains schema migrations

This service should not own auth fields such as password hashes or verification
codes. It owns review data only.
