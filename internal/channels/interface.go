package channels

import (
	"context"
	"github.com/alexnthnz/notification-system/internal/notification"
)

// Channel represents a notification channel interface
type Channel interface {
	SendNotification(ctx context.Context, notif notification.Notification) (*notification.DeliveryReport, error)
	GetChannelType() string
}

// ChannelManager manages all notification channels
type ChannelManager struct {
	channels map[string]Channel
}

// NewChannelManager creates a new channel manager
func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		channels: make(map[string]Channel),
	}
}

// RegisterChannel registers a channel with the manager
func (cm *ChannelManager) RegisterChannel(channel Channel) {
	cm.channels[channel.GetChannelType()] = channel
}

// GetChannel retrieves a channel by type
func (cm *ChannelManager) GetChannel(channelType string) (Channel, bool) {
	channel, exists := cm.channels[channelType]
	return channel, exists
}

// SendNotification sends a notification through the appropriate channel
func (cm *ChannelManager) SendNotification(ctx context.Context, notif notification.Notification) (*notification.DeliveryReport, error) {
	channel, exists := cm.GetChannel(notif.Channel)
	if !exists {
		return &notification.DeliveryReport{
			NotificationID: notif.ID,
			Status:         notification.StatusFailed,
			ErrorMessage:   "unsupported channel type: " + notif.Channel,
		}, nil
	}

	return channel.SendNotification(ctx, notif)
}