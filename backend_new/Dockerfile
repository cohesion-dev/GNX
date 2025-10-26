# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the entire project (needed for local module replacement)
COPY . .

# Change to backend_new directory
WORKDIR /app/backend_new

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/backend_new ./cmd/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/bin/backend_new .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./backend_new"]
