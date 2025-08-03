package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/alexnthnz/notification-system/api/proto/gen"
	"github.com/alexnthnz/notification-system/internal/notification"
)

// channelFromProto converts proto Channel to string
func channelFromProto(channel pb.Channel) string {
	switch channel {
	case pb.Channel_CHANNEL_EMAIL:
		return "email"
	case pb.Channel_CHANNEL_SMS:
		return "sms"
	case pb.Channel_CHANNEL_PUSH:
		return "push"
	default:
		return ""
	}
}

// channelToProto converts string channel to proto Channel
func channelToProto(channel string) pb.Channel {
	switch channel {
	case "email":
		return pb.Channel_CHANNEL_EMAIL
	case "sms":
		return pb.Channel_CHANNEL_SMS
	case "push":
		return pb.Channel_CHANNEL_PUSH
	default:
		return pb.Channel_CHANNEL_UNSPECIFIED
	}
}

// statusFromProto converts proto NotificationStatus to internal status
func statusFromProto(status pb.NotificationStatus) notification.NotificationStatus {
	switch status {
	case pb.NotificationStatus_NOTIFICATION_STATUS_PENDING:
		return notification.StatusPending
	case pb.NotificationStatus_NOTIFICATION_STATUS_SENT:
		return notification.StatusSent
	case pb.NotificationStatus_NOTIFICATION_STATUS_DELIVERED:
		return notification.StatusDelivered
	case pb.NotificationStatus_NOTIFICATION_STATUS_FAILED:
		return notification.StatusFailed
	case pb.NotificationStatus_NOTIFICATION_STATUS_CANCELLED:
		return notification.StatusCancelled
	default:
		return notification.StatusPending
	}
}

// statusToProto converts internal status to proto NotificationStatus
func statusToProto(status notification.NotificationStatus) pb.NotificationStatus {
	switch status {
	case notification.StatusPending:
		return pb.NotificationStatus_NOTIFICATION_STATUS_PENDING
	case notification.StatusSent:
		return pb.NotificationStatus_NOTIFICATION_STATUS_SENT
	case notification.StatusDelivered:
		return pb.NotificationStatus_NOTIFICATION_STATUS_DELIVERED
	case notification.StatusFailed:
		return pb.NotificationStatus_NOTIFICATION_STATUS_FAILED
	case notification.StatusCancelled:
		return pb.NotificationStatus_NOTIFICATION_STATUS_CANCELLED
	default:
		return pb.NotificationStatus_NOTIFICATION_STATUS_UNSPECIFIED
	}
}

// notificationToProto converts internal Notification to proto Notification
func notificationToProto(n *notification.Notification) *pb.Notification {
	protoNotif := &pb.Notification{
		Id:          n.ID,
		UserId:      n.UserID,
		Channel:     channelToProto(n.Channel),
		Recipient:   n.Recipient,
		Subject:     n.Subject,
		Body:        n.Body,
		Status:      statusToProto(n.Status),
		ExternalId:  n.ExternalID,
		ErrorMessage: n.ErrorMessage,
		RetryCount:  int32(n.RetryCount),
		CreatedAt:   timestamppb.New(n.CreatedAt),
		UpdatedAt:   timestamppb.New(n.UpdatedAt),
		Metadata:    n.Metadata,
	}

	// Handle optional timestamps
	if n.ScheduledAt != nil {
		protoNotif.ScheduledAt = timestamppb.New(*n.ScheduledAt)
	}
	if n.SentAt != nil {
		protoNotif.SentAt = timestamppb.New(*n.SentAt)
	}
	if n.DeliveredAt != nil {
		protoNotif.DeliveredAt = timestamppb.New(*n.DeliveredAt)
	}

	return protoNotif
}

// userPreferenceToProto converts internal UserPreference to proto UserPreference
func userPreferenceToProto(p *notification.UserPreference) *pb.UserPreference {
	var frequency pb.Frequency
	switch p.Frequency {
	case "immediate":
		frequency = pb.Frequency_FREQUENCY_IMMEDIATE
	case "hourly":
		frequency = pb.Frequency_FREQUENCY_HOURLY
	case "daily":
		frequency = pb.Frequency_FREQUENCY_DAILY
	default:
		frequency = pb.Frequency_FREQUENCY_UNSPECIFIED
	}

	return &pb.UserPreference{
		Id:        p.ID,
		UserId:    p.UserID,
		Channel:   channelToProto(p.Channel),
		Enabled:   p.Enabled,
		Frequency: frequency,
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}
}

// userPreferenceFromProto converts proto UserPreference to internal UserPreference
func userPreferenceFromProto(p *pb.UserPreference) *notification.UserPreference {
	var frequency string
	switch p.Frequency {
	case pb.Frequency_FREQUENCY_IMMEDIATE:
		frequency = "immediate"
	case pb.Frequency_FREQUENCY_HOURLY:
		frequency = "hourly"
	case pb.Frequency_FREQUENCY_DAILY:
		frequency = "daily"
	default:
		frequency = "immediate"
	}

	return &notification.UserPreference{
		ID:        p.Id,
		UserID:    p.UserId,
		Channel:   channelFromProto(p.Channel),
		Enabled:   p.Enabled,
		Frequency: frequency,
		CreatedAt: p.CreatedAt.AsTime(),
		UpdatedAt: p.UpdatedAt.AsTime(),
	}
}