package notification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/alexnthnz/notification-system/internal/database"
	"github.com/alexnthnz/notification-system/internal/queue"
	"github.com/google/uuid"
)

// Service handles notification business logic
type Service struct {
	db       *database.PostgresDB
	redis    *database.RedisClient
	producer *queue.Producer
}

// NewService creates a new notification service
func NewService(db *database.PostgresDB, redis *database.RedisClient, producer *queue.Producer) *Service {
	return &Service{
		db:       db,
		redis:    redis,
		producer: producer,
	}
}

// CreateNotification creates a new notification request
func (s *Service) CreateNotification(ctx context.Context, req NotificationRequest) (*Notification, error) {
	// Generate unique ID
	id := uuid.New().String()

	// Validate user preferences
	preferences, err := s.getUserPreferences(ctx, req.UserID, req.Channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	if !preferences.Enabled {
		return nil, fmt.Errorf("notifications disabled for user %s on channel %s", req.UserID, req.Channel)
	}

	// Set default priority if not specified
	priority := req.Priority
	if priority == 0 {
		priority = 2 // Medium priority
	}

	// Create notification record
	notification := &Notification{
		ID:          id,
		UserID:      req.UserID,
		Channel:     req.Channel,
		Recipient:   req.Recipient,
		Subject:     req.Subject,
		Body:        req.Body,
		Status:      StatusPending,
		RetryCount:  0,
		ScheduledAt: req.ScheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}

	// Insert into database
	query := `
		INSERT INTO notifications (id, user_id, channel, recipient, subject, body, status, scheduled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = s.db.ExecContext(ctx, query,
		notification.ID, notification.UserID, notification.Channel, notification.Recipient,
		notification.Subject, notification.Body, notification.Status, notification.ScheduledAt,
		notification.CreatedAt, notification.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert notification: %w", err)
	}

	// Check if notification should be sent immediately or scheduled
	if req.ScheduledAt == nil || req.ScheduledAt.Before(time.Now()) {
		// Publish to queue for immediate processing
		queueMsg := queue.NotificationMessage{
			ID:        notification.ID,
			UserID:    notification.UserID,
			Channel:   notification.Channel,
			Recipient: notification.Recipient,
			Subject:   notification.Subject,
			Body:      notification.Body,
			Metadata:  notification.Metadata,
			Priority:  priority,
			CreatedAt: notification.CreatedAt,
		}

		if err := s.producer.PublishNotification(ctx, queueMsg); err != nil {
			log.Printf("Failed to publish notification %s to queue: %v", id, err)
			// Don't return error here, notification is still created and can be retried
		}
	}

	log.Printf("Created notification %s for user %s via %s", id, req.UserID, req.Channel)
	return notification, nil
}

// GetNotification retrieves a notification by ID
func (s *Service) GetNotification(ctx context.Context, id string) (*Notification, error) {
	query := `
		SELECT id, user_id, channel, recipient, subject, body, status, external_id, 
		       error_message, retry_count, scheduled_at, sent_at, delivered_at, created_at, updated_at
		FROM notifications WHERE id = $1
	`

	var notification Notification
	var scheduledAt, sentAt, deliveredAt sql.NullTime
	var externalID, errorMessage sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID, &notification.UserID, &notification.Channel, &notification.Recipient,
		&notification.Subject, &notification.Body, &notification.Status, &externalID,
		&errorMessage, &notification.RetryCount, &scheduledAt, &sentAt, &deliveredAt,
		&notification.CreatedAt, &notification.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Handle nullable fields
	if externalID.Valid {
		notification.ExternalID = externalID.String
	}
	if errorMessage.Valid {
		notification.ErrorMessage = errorMessage.String
	}
	if scheduledAt.Valid {
		notification.ScheduledAt = &scheduledAt.Time
	}
	if sentAt.Valid {
		notification.SentAt = &sentAt.Time
	}
	if deliveredAt.Valid {
		notification.DeliveredAt = &deliveredAt.Time
	}

	return &notification, nil
}

// UpdateNotificationStatus updates the status of a notification
func (s *Service) UpdateNotificationStatus(ctx context.Context, id string, status NotificationStatus, externalID, errorMessage string) error {
	now := time.Now()

	query := `
		UPDATE notifications 
		SET status = $1, external_id = $2, error_message = $3, updated_at = $4
	`
	args := []interface{}{status, externalID, errorMessage, now}

	// Set sent_at timestamp for sent status
	if status == StatusSent {
		query += ", sent_at = $5"
		args = append(args, now)
	}

	// Set delivered_at timestamp for delivered status
	if status == StatusDelivered {
		query += ", delivered_at = $5"
		args = append(args, now)
	}

	query += " WHERE id = $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, id)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	log.Printf("Updated notification %s status to %s", id, status)
	return nil
}

// getUserPreferences retrieves user preferences for a specific channel
func (s *Service) getUserPreferences(ctx context.Context, userID, channel string) (*UserPreference, error) {
	// Try to get from cache first
	if s.redis != nil {
		cacheKey := fmt.Sprintf("user_preferences:%s:%s", userID, channel)
		if _, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
			log.Printf("Retrieved user preferences from cache for user %s", userID)
			// In a real implementation, you'd unmarshal the cached data
			// For now, we'll fall through to database
		}
	}

	// Get from database
	query := `
		SELECT id, user_id, channel, enabled, frequency, created_at, updated_at
		FROM user_preferences 
		WHERE user_id = $1 AND channel = $2
	`

	var pref UserPreference
	err := s.db.QueryRowContext(ctx, query, userID, channel).Scan(
		&pref.ID, &pref.UserID, &pref.Channel, &pref.Enabled,
		&pref.Frequency, &pref.CreatedAt, &pref.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default preferences if not found
			return &UserPreference{
				UserID:    userID,
				Channel:   channel,
				Enabled:   true,
				Frequency: "immediate",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Cache the result
	if s.redis != nil {
		s.redis.CacheUserPreferences(ctx, userID, pref)
	}

	return &pref, nil
}
