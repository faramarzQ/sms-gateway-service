# ---------- Builder ----------
FROM golang:1.26 AS builder

WORKDIR /app

# Download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build Message Consumer
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o sms-message-consumer \
    ./cmd/message_consumer

# ---------- Runtime ----------
FROM alpine:3.22

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/sms-message-consumer .

ENTRYPOINT ["./sms-message-consumer"]