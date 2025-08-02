package channels

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/alexnthnz/notification-system/internal/config"
	"github.com/alexnthnz/notification-system/internal/notification"
	"google.golang.org/api/option"
)

// PushChannel handles push notifications using Firebase Cloud Messaging
type PushChannel struct {
	client *messaging.Client
	config config.FirebaseConfig
}

// NewPushChannel creates a new push notification channel
func NewPushChannel(ctx context.Context, cfg config.FirebaseConfig) (*PushChannel, error) {
	// Check if credentials file exists
	if _, err := os.Stat(cfg.CredentialsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Firebase credentials file not found at %s", cfg.CredentialsPath)
	}

	// Initialize Firebase app
	opt := option.WithCredentialsFile(cfg.CredentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	// Get messaging client
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase messaging client: %w", err)
	}

	return &PushChannel{
		client: client,
		config: cfg,
	}, nil
}

// SendNotification sends a push notification
func (p *PushChannel) SendNotification(ctx context.Context, notif notification.Notification) (*notification.DeliveryReport, error) {
	log.Printf("Sending push notification %s to %s", notif.ID, notif.Recipient)

	// Parse additional data from metadata
	data := make(map[string]string)
	if notif.Metadata != nil {
		data = notif.Metadata
	}
	data["notification_id"] = notif.ID
	data["user_id"] = notif.UserID

	// Create the FCM message
	message := &messaging.Message{
		Token: notif.Recipient, // The recipient should be the FCM token
		Notification: &messaging.Notification{
			Title: notif.Subject,
			Body:  notif.Body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Priority: messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: notif.Subject,
						Body:  notif.Body,
					},
					Sound: "default",
				},
			},
		},
	}

	// Send the message
	response, err := p.client.Send(ctx, message)
	if err != nil {
		log.Printf("Failed to send push notification %s: %v", notif.ID, err)
		return &notification.DeliveryReport{
			NotificationID: notif.ID,
			Status:         notification.StatusFailed,
			ErrorMessage:   err.Error(),
		}, err
	}

	log.Printf("Successfully sent push notification %s (FCM response: %s)", notif.ID, response)
	return &notification.DeliveryReport{
		NotificationID: notif.ID,
		ExternalID:     response,
		Status:         notification.StatusSent,
	}, nil
}

// SendBulkNotification sends push notifications to multiple tokens
func (p *PushChannel) SendBulkNotification(ctx context.Context, tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens provided")
	}

	// Create the multicast message
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
		},
	}

	// Send to multiple devices
	response, err := p.client.SendMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send bulk push notification: %w", err)
	}

	log.Printf("Sent bulk push notification to %d tokens, %d successful, %d failed",
		len(tokens), response.SuccessCount, response.FailureCount)

	return response, nil
}

// GetChannelType returns the channel type
func (p *PushChannel) GetChannelType() string {
	return "push"
}
