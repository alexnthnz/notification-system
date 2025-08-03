# Notification Service (Go)

## Overview
The Notification Service is a scalable, high-performance system built in Go to deliver notifications across multiple channels, such as email, SMS, and push notifications. Inspired by Alex Xu's system design principles, it emphasizes modularity, scalability, and fault tolerance.

## Features

- **Multi-Channel Support**: Sends notifications via email, SMS, and mobile push notifications.
- **Scalability**: Handles high throughput using Go's concurrency model and distributed architecture.
- **Reliability**: Ensures delivery with retry mechanisms and fault-tolerant design.
- **Extensibility**: Easily integrates new notification channels or providers.
- **Monitoring and Analytics**: Tracks delivery status and system metrics.
- **Dual API Support**: Both REST and gRPC APIs for maximum flexibility.

## Architecture
The system follows a microservices-based architecture with the following components:

- **API Gateway**: Routes incoming HTTP/gRPC requests to the Notification Service.
- **Notification Service**: Core service written in Go to process and queue notifications.
- **Message Queue**: Uses Kafka for decoupling notification processing and delivery.
- **Channel Services**: Go-based services for each channel (Email, SMS, Push).
- **Database**: Stores notification metadata and user preferences.
- **Monitoring Service**: Tracks system health and delivery metrics.

## Data Flow

1. Clients send notification requests to the API Gateway via REST or gRPC.
2. The Notification Service validates requests, stores metadata, and publishes to a message queue.
3. Channel Services consume queue messages and send notifications via third-party providers (e.g., SendGrid for email, Twilio for SMS, Firebase for push).
4. Delivery status is updated in the database, and metrics are sent to the Monitoring Service.

## Technologies Used

- **Language**: Go (1.23+)
- **APIs**: REST with Gorilla Mux + gRPC with Protocol Buffers
- **Message Queue**: Apache Kafka
- **Database**: PostgreSQL for metadata, Redis for caching
- **Channel Services**: Go with HTTP clients for third-party API calls
- **Monitoring**: Prometheus and Grafana
- **Third-Party Providers**: SendGrid (Email), Twilio (SMS), Firebase (Push)

## Project Structure
```
notification-system/
├── cmd/
│   ├── api/               # Main API server (REST + gRPC)
│   ├── email-service/     # Email channel service
│   ├── sms-service/       # SMS channel service
│   ├── push-service/      # Push notification service
├── internal/
│   ├── config/            # Configuration loading (Viper)
│   ├── queue/             # Kafka client
│   ├── database/          # PostgreSQL/Redis clients
│   ├── notification/      # Core notification logic
│   ├── channels/          # Channel-specific logic
│   ├── monitoring/        # Prometheus metrics
├── api/
│   ├── proto/             # gRPC protobuf definitions
│   │   └── gen/           # Generated protobuf code
│   ├── grpc/              # gRPC service implementation
│   └── rest/              # REST API handlers
├── examples/
│   └── grpc-client/       # Example gRPC client
├── scripts/
│   └── generate-proto.sh  # Protobuf code generation
├── Dockerfile             # Docker configuration
├── docker-compose.yml     # Multi-service orchestration
├── Makefile              # Development commands
├── prometheus.yml         # Prometheus configuration
├── config.env.example     # Environment configuration example
├── go.mod                 # Go module dependencies
├── README.md              # This file
```

## Setup Instructions

### Prerequisites:
- Go 1.23+
- Docker and Docker Compose
- Protocol Buffers compiler (`protoc`)
- Accounts for third-party providers (SendGrid, Twilio, Firebase)

### Installation:
```bash
git clone https://github.com/your-repo/notification-system.git
cd notification-system

# Install development dependencies
make dev-deps

# Generate protobuf code
make proto

# Build all services
make build
```

### Configuration:
1. Copy the example configuration:
```bash
cp config.env.example .env
```

2. Update the `.env` file with your configuration:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
REDIS_ADDR=localhost:6379
KAFKA_BROKERS=localhost:9092
SENDGRID_API_KEY=your-sendgrid-key
TWILIO_ACCOUNT_SID=your-twilio-sid
TWILIO_AUTH_TOKEN=your-twilio-token
FIREBASE_CREDENTIALS_PATH=/path/to/firebase-credentials.json
```

### Running Locally:

#### Using Docker Compose (Recommended):
```bash
# Start all services including dependencies
make docker-up

# View logs
make docker-logs

# Stop all services
make docker-down
```

#### Manual Setup:
```bash
# Start dependencies
docker-compose up -d postgres redis kafka

# Start services manually
make run-api          # Start API server (REST + gRPC)
make run-email        # Start email service
make run-sms          # Start SMS service
make run-push         # Start push service
```

## API Endpoints

### REST API (Port 8080)

#### POST /api/v1/notifications
Create a new notification
```json
{
  "user_id": "123",
  "channel": "email",
  "recipient": "user@example.com",
  "subject": "Welcome!",
  "body": "Thank you for joining.",
  "priority": 1
}
```

#### GET /api/v1/notifications/{id}
Retrieve notification status
```json
{
  "id": "uuid",
  "user_id": "123",
  "channel": "email",
  "recipient": "user@example.com",
  "subject": "Welcome!",
  "body": "Thank you for joining.",
  "status": "sent",
  "created_at": "2023-01-01T00:00:00Z"
}
```

#### GET /health
Health check endpoint

#### GET /metrics
Prometheus metrics endpoint

### gRPC API (Port 9090)

The service exposes a full gRPC API defined in `api/proto/notification.proto`. Key methods include:

- `CreateNotification` - Create a new notification
- `GetNotification` - Retrieve notification by ID
- `ListNotifications` - List notifications with filtering
- `UpdateNotificationStatus` - Update notification status
- `GetUserPreferences` - Get user notification preferences
- `UpdateUserPreferences` - Update user preferences

#### Example gRPC Usage:
```bash
# Test gRPC API with example client
make grpc-test

# List available gRPC services
make grpc-list

# Manual gRPC call with grpcurl
grpcurl -plaintext -d '{"user_id":"123","channel":"CHANNEL_EMAIL","recipient":"test@example.com","subject":"Test","body":"Hello gRPC!"}' localhost:9090 notification.v1.NotificationService/CreateNotification
```

## Development Commands

The project includes a comprehensive Makefile for development:

```bash
make help           # Show all available commands
make build          # Build all services
make proto          # Generate protobuf code
make test           # Run tests
make lint           # Run linter
make docker-up      # Start with Docker Compose
make docker-down    # Stop all services
make grpc-test      # Test gRPC API
make example-rest   # Test REST API
```

## Scaling Considerations

- **Concurrency**: Leverages Go's goroutines and channels for high concurrency.
- **Horizontal Scaling**: Deploy multiple instances of services with a load balancer.
- **Queue Partitioning**: Uses Kafka partitions for load distribution.
- **Database Optimization**: PostgreSQL with proper indexing and connection pooling.
- **Caching**: Redis for user preferences and rate limiting.
- **API Gateway**: Support for both REST and gRPC protocols.

## Monitoring and Logging

- **Metrics**: Prometheus endpoints expose delivery rates, latency, and error metrics.
- **Dashboards**: Grafana visualizes system performance.
- **Logging**: Structured logging with Zap to stdout.
- **Health Checks**: Each service exposes health endpoints.
- **gRPC Reflection**: Enabled for development tools.

## Service Architecture

### API Service (Ports 8080 + 9090)
- **REST API** (8080): HTTP endpoints for web clients
- **gRPC API** (9090): High-performance binary protocol for service-to-service communication
- Validates requests and publishes to Kafka queue
- Manages user preferences and rate limiting

### Email Service
- Consumes email notifications from Kafka
- Integrates with SendGrid for email delivery
- Updates notification status in database

### SMS Service
- Consumes SMS notifications from Kafka
- Integrates with Twilio for SMS delivery
- Handles delivery reports and status updates

### Push Service
- Consumes push notifications from Kafka
- Integrates with Firebase Cloud Messaging
- Supports both Android and iOS devices

## gRPC Protocol Buffer Schema

The gRPC API is defined using Protocol Buffers with the following key message types:

```protobuf
// Create notification request
message CreateNotificationRequest {
  string user_id = 1;
  Channel channel = 2;      // EMAIL, SMS, PUSH
  string recipient = 3;
  string subject = 4;
  string body = 5;
  Priority priority = 6;    // LOW, MEDIUM, HIGH
  google.protobuf.Timestamp scheduled_at = 7;
  map<string, string> metadata = 10;
}

// Notification entity
message Notification {
  string id = 1;
  string user_id = 2;
  Channel channel = 3;
  string recipient = 4;
  string subject = 5;
  string body = 6;
  NotificationStatus status = 7;  // PENDING, SENT, DELIVERED, FAILED
  google.protobuf.Timestamp created_at = 14;
  // ... additional fields
}
```

## Database Schema

### Users Table
- id (UUID, Primary Key)
- email (VARCHAR, Unique)
- phone (VARCHAR)
- push_token (VARCHAR)

### Notifications Table
- id (UUID, Primary Key)
- user_id (UUID, Foreign Key)
- channel (VARCHAR)
- recipient (VARCHAR)
- subject (VARCHAR)
- body (TEXT)
- status (VARCHAR)
- external_id (VARCHAR)
- retry_count (INTEGER)
- created_at (TIMESTAMP)

### User Preferences Table
- id (UUID, Primary Key)
- user_id (UUID, Foreign Key)
- channel (VARCHAR)
- enabled (BOOLEAN)
- frequency (VARCHAR)

## Future Improvements

- Add support for in-app notifications
- Implement rate limiting using golang.org/x/time/rate
- Introduce A/B testing for notification content
- Enhance retry logic with exponential backoff
- Add notification templates with variables
- Implement webhook delivery confirmations
- Add notification scheduling and batching
- Add authentication and authorization to gRPC API
- Implement streaming APIs for real-time updates

## Contributing
Contributions are welcome! Follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature-name`)
3. Commit changes (`git commit -m 'Add feature'`)
4. Push to the branch (`git push origin feature-name`)
5. Open a pull request

## License
This project is licensed under the MIT License. See the LICENSE file for details.