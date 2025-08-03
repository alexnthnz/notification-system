package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/alexnthnz/notification-system/api/proto/gen"
	"github.com/alexnthnz/notification-system/internal/monitoring"
	"github.com/alexnthnz/notification-system/internal/notification"
)

// Server implements the NotificationService gRPC server
type Server struct {
	pb.UnimplementedNotificationServiceServer
	notificationService *notification.Service
	metrics            *monitoring.Metrics
	logger             *zap.Logger
}

// NewServer creates a new gRPC server
func NewServer(
	notificationService *notification.Service,
	metrics *monitoring.Metrics,
	logger *zap.Logger,
) *Server {
	return &Server{
		notificationService: notificationService,
		metrics:            metrics,
		logger:             logger,
	}
}

// CreateNotification creates a new notification
func (s *Server) CreateNotification(ctx context.Context, req *pb.CreateNotificationRequest) (*pb.CreateNotificationResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		s.metrics.RecordProcessingDuration("grpc", "create_notification", duration)
	}()

	s.logger.Info("gRPC CreateNotification request",
		zap.String("user_id", req.UserId),
		zap.String("channel", req.Channel.String()),
		zap.String("recipient", req.Recipient),
	)

	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Channel == pb.Channel_CHANNEL_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}
	if req.Recipient == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient is required")
	}
	if req.Body == "" {
		return nil, status.Error(codes.InvalidArgument, "body is required")
	}

	// Convert gRPC request to internal request
	notifReq := notification.NotificationRequest{
		UserID:    req.UserId,
		Channel:   channelFromProto(req.Channel),
		Recipient: req.Recipient,
		Subject:   req.Subject,
		Body:      req.Body,
		Priority:  int(req.Priority),
		Variables: req.Variables,
		Metadata:  req.Metadata,
	}

	// Handle scheduled_at
	if req.ScheduledAt != nil {
		scheduledAt := req.ScheduledAt.AsTime()
		notifReq.ScheduledAt = &scheduledAt
	}

	// Create notification
	notif, err := s.notificationService.CreateNotification(ctx, notifReq)
	if err != nil {
		s.logger.Error("Failed to create notification", zap.Error(err))
		s.metrics.RecordNotificationFailed(notifReq.Channel, "creation_error")
		return nil, status.Error(codes.Internal, "failed to create notification")
	}

	s.metrics.RecordNotificationSent(notifReq.Channel, "created")
	s.logger.Info("Notification created via gRPC",
		zap.String("id", notif.ID),
		zap.String("channel", notif.Channel),
	)

	return &pb.CreateNotificationResponse{
		Id:        notif.ID,
		Status:    statusToProto(notif.Status),
		Message:   "Notification created successfully",
		CreatedAt: timestamppb.New(notif.CreatedAt),
	}, nil
}

// GetNotification retrieves a notification by ID
func (s *Server) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.GetNotificationResponse, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		s.metrics.RecordProcessingDuration("grpc", "get_notification", duration)
	}()

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	notif, err := s.notificationService.GetNotification(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get notification", zap.Error(err), zap.String("id", req.Id))
		if err.Error() == "notification not found" {
			return nil, status.Error(codes.NotFound, "notification not found")
		}
		return nil, status.Error(codes.Internal, "failed to retrieve notification")
	}

	return &pb.GetNotificationResponse{
		Notification: notificationToProto(notif),
	}, nil
}

// ListNotifications lists notifications for a user (placeholder implementation)
func (s *Server) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	// This is a placeholder implementation
	// In a real system, you'd implement pagination and filtering
	return &pb.ListNotificationsResponse{
		Notifications:   []*pb.Notification{},
		NextPageToken:   "",
		TotalCount:      0,
	}, nil
}

// UpdateNotificationStatus updates the status of a notification
func (s *Server) UpdateNotificationStatus(ctx context.Context, req *pb.UpdateNotificationStatusRequest) (*pb.UpdateNotificationStatusResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Status == pb.NotificationStatus_NOTIFICATION_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	err := s.notificationService.UpdateNotificationStatus(
		ctx,
		req.Id,
		statusFromProto(req.Status),
		req.ExternalId,
		req.ErrorMessage,
	)
	if err != nil {
		s.logger.Error("Failed to update notification status", zap.Error(err), zap.String("id", req.Id))
		return nil, status.Error(codes.Internal, "failed to update notification status")
	}

	return &pb.UpdateNotificationStatusResponse{
		Success: true,
		Message: "Notification status updated successfully",
	}, nil
}

// GetUserPreferences retrieves user notification preferences (placeholder)
func (s *Server) GetUserPreferences(ctx context.Context, req *pb.GetUserPreferencesRequest) (*pb.GetUserPreferencesResponse, error) {
	// Placeholder implementation
	return &pb.GetUserPreferencesResponse{
		Preferences: []*pb.UserPreference{},
	}, nil
}

// UpdateUserPreferences updates user notification preferences (placeholder)
func (s *Server) UpdateUserPreferences(ctx context.Context, req *pb.UpdateUserPreferencesRequest) (*pb.UpdateUserPreferencesResponse, error) {
	// Placeholder implementation
	return &pb.UpdateUserPreferencesResponse{
		Success: true,
		Message: "User preferences updated successfully",
	}, nil
}