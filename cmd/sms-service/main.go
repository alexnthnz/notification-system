package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/alexnthnz/notification-system/internal/channels"
	"github.com/alexnthnz/notification-system/internal/config"
	"github.com/alexnthnz/notification-system/internal/database"
	"github.com/alexnthnz/notification-system/internal/monitoring"
	"github.com/alexnthnz/notification-system/internal/notification"
	"github.com/alexnthnz/notification-system/internal/queue"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting SMS Service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize metrics
	metrics := monitoring.NewMetrics()

	// Connect to PostgreSQL
	postgres, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

	// Connect to Redis
	redis, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redis.Close()

	// Initialize notification service
	notificationService := notification.NewService(postgres, redis, nil)

	// Initialize SMS channel
	smsChannel := channels.NewSMSChannel(cfg.Channels.Twilio)
	logger.Info("SMS channel initialized")

	// Initialize Kafka consumer
	consumer := queue.NewConsumer(cfg.Kafka, "sms-service")
	defer consumer.Close()
	logger.Info("Kafka consumer initialized")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consuming notifications
	go func() {
		logger.Info("Starting to consume SMS notifications")
		err := consumer.ConsumeNotifications(ctx, func(msg queue.NotificationMessage) error {
			return processSMSNotification(ctx, msg, smsChannel, notificationService, metrics, logger)
		})
		if err != nil && err != context.Canceled {
			logger.Error("Consumer error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down SMS service...")
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(5 * time.Second)
	logger.Info("SMS service exited")
}

func processSMSNotification(
	ctx context.Context,
	msg queue.NotificationMessage,
	smsChannel *channels.SMSChannel,
	notificationService *notification.Service,
	metrics *monitoring.Metrics,
	logger *zap.Logger,
) error {
	// Skip if not SMS channel
	if msg.Channel != "sms" {
		return nil
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RecordChannelDuration("sms", duration)
	}()

	logger.Info("Processing SMS notification", 
		zap.String("id", msg.ID),
		zap.String("recipient", msg.Recipient),
	)

	// Get full notification details
	notif, err := notificationService.GetNotification(ctx, msg.ID)
	if err != nil {
		logger.Error("Failed to get notification details", zap.Error(err), zap.String("id", msg.ID))
		return err
	}

	// Send SMS
	report, err := smsChannel.SendNotification(ctx, *notif)
	if err != nil {
		logger.Error("Failed to send SMS", zap.Error(err), zap.String("id", msg.ID))
		metrics.RecordNotificationFailed("sms", "send_error")
		
		// Update notification status
		notificationService.UpdateNotificationStatus(ctx, msg.ID, notification.StatusFailed, "", err.Error())
		return err
	}

	// Update notification status based on report
	if report.Status == notification.StatusSent {
		metrics.RecordNotificationSent("sms", "sent")
		err = notificationService.UpdateNotificationStatus(ctx, msg.ID, notification.StatusSent, report.ExternalID, "")
	} else {
		metrics.RecordNotificationFailed("sms", "provider_error")
		err = notificationService.UpdateNotificationStatus(ctx, msg.ID, report.Status, report.ExternalID, report.ErrorMessage)
	}

	if err != nil {
		logger.Error("Failed to update notification status", zap.Error(err), zap.String("id", msg.ID))
		return err
	}

	logger.Info("SMS notification processed successfully", zap.String("id", msg.ID))
	return nil
}