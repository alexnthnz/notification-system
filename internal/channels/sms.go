package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/alexnthnz/notification-system/internal/config"
	"github.com/alexnthnz/notification-system/internal/notification"
)

// SMSChannel handles SMS notifications using Twilio
type SMSChannel struct {
	config config.TwilioConfig
	client *http.Client
}

// NewSMSChannel creates a new SMS channel
func NewSMSChannel(cfg config.TwilioConfig) *SMSChannel {
	return &SMSChannel{
		config: cfg,
		client: &http.Client{},
	}
}

// TwilioResponse represents the response from Twilio API
type TwilioResponse struct {
	SID         string `json:"sid"`
	Status      string `json:"status"`
	ErrorCode   *int   `json:"error_code,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
}

// SendNotification sends an SMS notification
func (s *SMSChannel) SendNotification(ctx context.Context, notif notification.Notification) (*notification.DeliveryReport, error) {
	log.Printf("Sending SMS notification %s to %s", notif.ID, notif.Recipient)

	// Prepare the request data
	data := url.Values{}
	data.Set("To", notif.Recipient)
	data.Set("From", "+1234567890") // Your Twilio phone number
	data.Set("Body", notif.Body)

	// Create the request
	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.config.AccountSID)
	req, err := http.NewRequestWithContext(ctx, "POST", twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		return &notification.DeliveryReport{
			NotificationID: notif.ID,
			Status:         notification.StatusFailed,
			ErrorMessage:   err.Error(),
		}, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.config.AccountSID, s.config.AuthToken)

	// Send the request
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Failed to send SMS notification %s: %v", notif.ID, err)
		return &notification.DeliveryReport{
			NotificationID: notif.ID,
			Status:         notification.StatusFailed,
			ErrorMessage:   err.Error(),
		}, err
	}
	defer resp.Body.Close()

	// Parse the response
	var twilioResp TwilioResponse
	if err := json.NewDecoder(resp.Body).Decode(&twilioResp); err != nil {
		return &notification.DeliveryReport{
			NotificationID: notif.ID,
			Status:         notification.StatusFailed,
			ErrorMessage:   "Failed to parse Twilio response",
		}, err
	}

	// Check if the request was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Successfully sent SMS notification %s (Twilio SID: %s)", notif.ID, twilioResp.SID)
		return &notification.DeliveryReport{
			NotificationID: notif.ID,
			ExternalID:     twilioResp.SID,
			Status:         notification.StatusSent,
		}, nil
	}

	// Handle error response
	errorMsg := "Unknown Twilio error"
	if twilioResp.ErrorMessage != nil {
		errorMsg = *twilioResp.ErrorMessage
	}

	log.Printf("SMS notification %s failed: %s", notif.ID, errorMsg)
	return &notification.DeliveryReport{
		NotificationID: notif.ID,
		Status:         notification.StatusFailed,
		ErrorMessage:   errorMsg,
	}, fmt.Errorf("twilio error: %s", errorMsg)
}

// GetChannelType returns the channel type
func (s *SMSChannel) GetChannelType() string {
	return "sms"
}