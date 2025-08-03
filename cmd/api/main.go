package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcapi "github.com/alexnthnz/notification-system/api/grpc"
	pb "github.com/alexnthnz/notification-system/api/proto/gen"
	"github.com/alexnthnz/notification-system/api/rest"
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

	logger.Info("Starting Notification API Service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize metrics
	metrics := monitoring.NewMetrics()
	logger.Info("Metrics initialized")

	// Connect to PostgreSQL
	postgres, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer postgres.Close()

	// Initialize database schema
	if err := postgres.InitSchema(); err != nil {
		logger.Fatal("Failed to initialize database schema", zap.Error(err))
	}
	logger.Info("Database connected and schema initialized")

	// Connect to Redis
	redis, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redis.Close()
	logger.Info("Redis connected")

	// Initialize Kafka producer
	producer := queue.NewProducer(cfg.Kafka)
	defer producer.Close()
	logger.Info("Kafka producer initialized")

	// Initialize notification service
	notificationService := notification.NewService(postgres, redis, producer)
	logger.Info("Notification service initialized")

	// Initialize REST API handler
	handler := rest.NewHandler(notificationService, metrics, logger)
	router := handler.SetupRoutes()

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	grpcHandler := grpcapi.NewServer(notificationService, metrics, logger)
	
	// Register the notification service
	pb.RegisterNotificationServiceServer(grpcServer, grpcHandler)
	
	// Enable reflection for grpcurl and other tools
	reflection.Register(grpcServer)

	// Start gRPC server in a goroutine
	go func() {
		grpcAddr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.GRPCPort)
		listener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC", zap.Error(err))
		}

		logger.Info("Starting gRPC server", zap.String("addr", grpcAddr))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Start metrics server if enabled
	if cfg.Metrics.Enabled {
		metricsServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
			Handler: metrics.Handler(),
		}

		go func() {
			logger.Info("Starting metrics server", zap.Int("port", cfg.Metrics.Port))
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("Metrics server error", zap.Error(err))
			}
		}()
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	logger.Info("Servers exited")
}