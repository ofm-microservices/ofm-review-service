# OFM Review Service

## Purpose

The Review Service owns buyer reviews and their read-model projections. It validates review creation against completed order data and does not own order completion, user profiles, or payment state. Status: active.

## Interfaces and flow

The gateway or order flow invokes the review write contract after an eligible completion. The service validates the buyer and order snapshot, persists the review and outbox event in PostgreSQL, and publishes changes for Redis and monolith projections. Duplicate event processing is rejected or ignored by event identity.

## Configuration

.env.example groups are DB_*, MIGRATIONS_*, REDIS_*, gRPC/NATS/Kafka, retry settings, and observability. PostgreSQL is authoritative; Redis is derived; broker values select commands, events, and consumers.

## Local development

    cp .env.example .env
    just run
    go test ./...

## Build and operations

Dockerfile builds ofm/review-service:<tag>. Helm deploys the service. Diagnose order eligibility, review rows, outbox/CDC, Kafka lag, Redis projection, and migration audit state.

