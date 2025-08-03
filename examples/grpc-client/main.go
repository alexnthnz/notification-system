package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/alexnthnz/notification-system/api/proto/gen"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := pb.NewNotificationServiceClient(conn)

	// Example 1: Create an email notification
	emailReq := &pb.CreateNotificationRequest{
		UserId:    "user123",
		Channel:   pb.Channel_CHANNEL_EMAIL,
		Recipient: "user@example.com",
		Subject:   "Welcome!",
		Body:      "Thank you for signing up for our service.",
		Priority:  pb.Priority_PRIORITY_HIGH,
		Metadata: map[string]string{
			"campaign_id": "welcome_series",
			"template":    "welcome_email",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	emailResp, err := client.CreateNotification(ctx, emailReq)
	if err != nil {
		log.Fatalf("Failed to create email notification: %v", err)
	}

	log.Printf("Email notification created: ID=%s, Status=%s", 
		emailResp.Id, emailResp.Status.String())

	// Example 2: Create an SMS notification
	smsReq := &pb.CreateNotificationRequest{
		UserId:    "user123",
		Channel:   pb.Channel_CHANNEL_SMS,
		Recipient: "+1234567890",
		Body:      "Your verification code is: 123456",
		Priority:  pb.Priority_PRIORITY_HIGH,
	}

	smsResp, err := client.CreateNotification(ctx, smsReq)
	if err != nil {
		log.Fatalf("Failed to create SMS notification: %v", err)
	}

	log.Printf("SMS notification created: ID=%s, Status=%s", 
		smsResp.Id, smsResp.Status.String())

	// Example 3: Create a scheduled push notification
	scheduledTime := time.Now().Add(5 * time.Minute)
	pushReq := &pb.CreateNotificationRequest{
		UserId:      "user123",
		Channel:     pb.Channel_CHANNEL_PUSH,
		Recipient:   "fcm_token_here",
		Subject:     "Don't forget!",
		Body:        "You have a meeting in 5 minutes.",
		Priority:    pb.Priority_PRIORITY_MEDIUM,
		ScheduledAt: timestamppb.New(scheduledTime),
	}

	pushResp, err := client.CreateNotification(ctx, pushReq)
	if err != nil {
		log.Fatalf("Failed to create push notification: %v", err)
	}

	log.Printf("Push notification scheduled: ID=%s, Status=%s", 
		pushResp.Id, pushResp.Status.String())

	// Example 4: Get notification details
	getReq := &pb.GetNotificationRequest{
		Id: emailResp.Id,
	}

	getResp, err := client.GetNotification(ctx, getReq)
	if err != nil {
		log.Fatalf("Failed to get notification: %v", err)
	}

	notif := getResp.Notification
	log.Printf("Retrieved notification: ID=%s, Channel=%s, Status=%s, CreatedAt=%s",
		notif.Id, notif.Channel.String(), notif.Status.String(), 
		notif.CreatedAt.AsTime().Format(time.RFC3339))

	// Example 5: Update notification status (typically done by channel services)
	updateReq := &pb.UpdateNotificationStatusRequest{
		Id:         emailResp.Id,
		Status:     pb.NotificationStatus_NOTIFICATION_STATUS_SENT,
		ExternalId: "sendgrid_message_id_123",
	}

	updateResp, err := client.UpdateNotificationStatus(ctx, updateReq)
	if err != nil {
		log.Fatalf("Failed to update notification status: %v", err)
	}

	log.Printf("Notification status updated: Success=%t, Message=%s",
		updateResp.Success, updateResp.Message)

	log.Println("All gRPC operations completed successfully!")
}