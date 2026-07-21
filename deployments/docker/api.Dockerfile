# ---------- Builder ----------
FROM golang:1.26 AS builder

WORKDIR /app

# Download dependencies first (better Docker layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o sms-api \
    ./cmd/api

# ---------- Runtime ----------
FROM alpine:3.22

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/sms-api .

EXPOSE 8080

ENTRYPOINT ["./sms-api"]