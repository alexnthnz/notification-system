# Build stage
FROM golang:1.23.10-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o email-service ./cmd/email-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sms-service ./cmd/sms-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o push-service ./cmd/push-service

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /root/

# Copy binaries from builder stage
COPY --from=builder /app/api .
COPY --from=builder /app/email-service .
COPY --from=builder /app/sms-service .
COPY --from=builder /app/push-service .

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of the binaries
RUN chown -R appuser:appgroup /root/

# Switch to non-root user
USER appuser

# Default command (can be overridden)
CMD ["./api"]