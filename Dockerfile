# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for Go modules
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o carbon-slots ./cmd/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/carbon-slots .

EXPOSE 3000

ENTRYPOINT ["./carbon-slots"]
