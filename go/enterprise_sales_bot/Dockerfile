# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/sales_bot ./cmd/sales_bot

# Final stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bin/sales_bot .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./sales_bot"]
