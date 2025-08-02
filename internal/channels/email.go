package channels

import (
	"context"
	"fmt"
	"log"

	"github.com/alexnthnz/notification-system/internal/config"
	"github.com/alexnthnz/notification-system/internal/notification"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailChannel handles email notifications using SendGrid
type EmailChannel struct {
	client *sendgrid.Client
	config config.SendGridConfig
}

// NewEmailChannel creates a new email channel
func NewEmailChannel(cfg config.SendGridConfig) *EmailChannel {
	client := sendgrid.NewSendClient(cfg.APIKey)
	return &EmailChannel{
		client: client,
		config: cfg,
	}
}

// SendNotification sends an email notification
func (e *EmailChannel) SendNotification(ctx context.Context, notif notification.Notification) (*notification.DeliveryReport, error) {
	log.Printf("Sending email notification %s to %s", notif.ID, notif.Recipient)

	// Create the email message
	from := mail.NewEmail("Notification Service", "noreply@yourcompany.com")
	to := mail.NewEmail("", notif.Recipient)

	message := mail.NewSingleEmail(from, notif.Subject, to, notif.Body, notif.Body)

	// Add custom headers for tracking
	message.SetHeader("X-Notification-ID", notif.ID)
	if notif.UserID != "" {
		message.SetHeader("X-User-ID", notif.UserID)
	}

	// Send the email
	response, err := e.client.Send(message)
	if err != nil {
		log.Printf("Failed to send email notification %s: %v", notif.ID, err)
		return &notification.DeliveryReport{
			NotificationID: notif.ID,
			Status:         notification.StatusFailed,
			ErrorMessage:   err.Error(),
		}, err
	}

	// Check response status
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		var messageID string
		if msgIDs, ok := response.Headers["X-Message-Id"]; ok && len(msgIDs) > 0 {
			messageID = msgIDs[0]
		}
		log.Printf("Successfully sent email notification %s (SendGrid ID: %s)", notif.ID, messageID)
		return &notification.DeliveryReport{
			NotificationID: notif.ID,
			ExternalID:     messageID,
			Status:         notification.StatusSent,
		}, nil
	}

	errorMsg := fmt.Sprintf("SendGrid returned status %d: %s", response.StatusCode, response.Body)
	log.Printf("Email notification %s failed: %s", notif.ID, errorMsg)

	return &notification.DeliveryReport{
		NotificationID: notif.ID,
		Status:         notification.StatusFailed,
		ErrorMessage:   errorMsg,
	}, fmt.Errorf("sendgrid error: %s", errorMsg)
}

// GetChannelType returns the channel type
func (e *EmailChannel) GetChannelType() string {
	return "email"
}
