# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install CA certificates (required to fetch Go modules via HTTPS)
RUN apk add --no-cache ca-certificates

# Copy and build the Go app
COPY . .
RUN go build -o gogunzy .

# Final stage
FROM alpine:3.18

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/gogunzy .

# Expose the port
EXPOSE 8000

# Run the app
CMD ["./gogunzy"]