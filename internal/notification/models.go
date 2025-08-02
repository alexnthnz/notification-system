package notification

import (
	"time"
)

// Notification represents a notification entity
type Notification struct {
	ID          string            `json:"id" db:"id"`
	UserID      string            `json:"user_id" db:"user_id"`
	Channel     string            `json:"channel" db:"channel"`
	Recipient   string            `json:"recipient" db:"recipient"`
	Subject     string            `json:"subject,omitempty" db:"subject"`
	Body        string            `json:"body" db:"body"`
	Status      NotificationStatus `json:"status" db:"status"`
	ExternalID  string            `json:"external_id,omitempty" db:"external_id"`
	ErrorMessage string           `json:"error_message,omitempty" db:"error_message"`
	RetryCount  int               `json:"retry_count" db:"retry_count"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty" db:"scheduled_at"`
	SentAt      *time.Time        `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt *time.Time        `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
	StatusCancelled NotificationStatus = "cancelled"
)

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	UserID    string            `json:"user_id" validate:"required"`
	Channel   string            `json:"channel" validate:"required,oneof=email sms push"`
	Recipient string            `json:"recipient" validate:"required"`
	Subject   string            `json:"subject,omitempty"`
	Body      string            `json:"body" validate:"required"`
	Priority  int               `json:"priority,omitempty"` // 1 = high, 2 = medium, 3 = low
	ScheduledAt *time.Time      `json:"scheduled_at,omitempty"`
	Template  string            `json:"template,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// User represents a user entity
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone,omitempty" db:"phone"`
	PushToken string    `json:"push_token,omitempty" db:"push_token"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserPreference represents user notification preferences
type UserPreference struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Channel   string    `json:"channel" db:"channel"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	Frequency string    `json:"frequency" db:"frequency"` // immediate, hourly, daily
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID              string            `json:"id" db:"id"`
	Name            string            `json:"name" db:"name"`
	Channel         string            `json:"channel" db:"channel"`
	SubjectTemplate string            `json:"subject_template,omitempty" db:"subject_template"`
	BodyTemplate    string            `json:"body_template" db:"body_template"`
	Variables       map[string]string `json:"variables,omitempty" db:"variables"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// DeliveryReport represents a delivery report from a channel provider
type DeliveryReport struct {
	NotificationID string             `json:"notification_id"`
	ExternalID     string             `json:"external_id"`
	Status         NotificationStatus `json:"status"`
	ErrorMessage   string             `json:"error_message,omitempty"`
	DeliveredAt    *time.Time         `json:"delivered_at,omitempty"`
	Metadata       map[string]string  `json:"metadata,omitempty"`
}