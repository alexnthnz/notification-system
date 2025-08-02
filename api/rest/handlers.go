package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	
	"github.com/alexnthnz/notification-system/internal/notification"
	"github.com/alexnthnz/notification-system/internal/monitoring"
)

// Handler holds dependencies for REST API handlers
type Handler struct {
	notificationService *notification.Service
	metrics            *monitoring.Metrics
	logger             *zap.Logger
	validator          *validator.Validate
}

// NewHandler creates a new REST API handler
func NewHandler(
	notificationService *notification.Service,
	metrics *monitoring.Metrics,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		notificationService: notificationService,
		metrics:            metrics,
		logger:             logger,
		validator:          validator.New(),
	}
}

// CreateNotificationRequest represents the request body for creating notifications
type CreateNotificationRequest struct {
	UserID      string            `json:"user_id" validate:"required"`
	Channel     string            `json:"channel" validate:"required,oneof=email sms push"`
	Recipient   string            `json:"recipient" validate:"required"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body" validate:"required"`
	Priority    int               `json:"priority,omitempty"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	Template    string            `json:"template,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CreateNotificationResponse represents the response for creating notifications
type CreateNotificationResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// CreateNotification handles POST /notifications
func (h *Handler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		h.metrics.RecordProcessingDuration("api", "create_notification", duration)
	}()

	h.metrics.IncrementActiveConnections()
	defer h.metrics.DecrementActiveConnections()

	var req CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.writeErrorResponse(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Convert to notification request
	notifReq := notification.NotificationRequest{
		UserID:      req.UserID,
		Channel:     req.Channel,
		Recipient:   req.Recipient,
		Subject:     req.Subject,
		Body:        req.Body,
		Priority:    req.Priority,
		ScheduledAt: req.ScheduledAt,
		Variables:   req.Variables,
		Metadata:    req.Metadata,
	}

	// Create notification
	notif, err := h.notificationService.CreateNotification(r.Context(), notifReq)
	if err != nil {
		h.logger.Error("Failed to create notification", zap.Error(err))
		h.metrics.RecordNotificationFailed(req.Channel, "creation_error")
		h.writeErrorResponse(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	h.metrics.RecordNotificationSent(req.Channel, "created")
	h.logger.Info("Notification created", 
		zap.String("id", notif.ID),
		zap.String("channel", notif.Channel),
		zap.String("user_id", notif.UserID),
	)

	response := CreateNotificationResponse{
		ID:      notif.ID,
		Status:  string(notif.Status),
		Message: "Notification created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetNotification handles GET /notifications/{id}
func (h *Handler) GetNotification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		h.metrics.RecordProcessingDuration("api", "get_notification", duration)
	}()

	h.metrics.IncrementActiveConnections()
	defer h.metrics.DecrementActiveConnections()

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeErrorResponse(w, "Notification ID is required", http.StatusBadRequest)
		return
	}

	notif, err := h.notificationService.GetNotification(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get notification", zap.Error(err), zap.String("id", id))
		if err.Error() == "notification not found" {
			h.writeErrorResponse(w, "Notification not found", http.StatusNotFound)
		} else {
			h.writeErrorResponse(w, "Failed to retrieve notification", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notif)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "notification-api",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Metrics handles GET /metrics (Prometheus metrics)
func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	h.metrics.Handler().ServeHTTP(w, r)
}

// writeErrorResponse writes an error response
func (h *Handler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// SetupRoutes sets up all REST API routes
func (h *Handler) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/notifications", h.CreateNotification).Methods("POST")
	api.HandleFunc("/notifications/{id}", h.GetNotification).Methods("GET")

	// Health and metrics
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
	router.HandleFunc("/metrics", h.Metrics).Methods("GET")

	// Add middleware
	router.Use(h.loggingMiddleware)
	router.Use(h.corsMiddleware)

	return router
}

// loggingMiddleware logs HTTP requests
func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response recorder to capture status code
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(recorder, r)
		
		duration := time.Since(start)
		h.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", recorder.statusCode),
			zap.Duration("duration", duration),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}

// corsMiddleware adds CORS headers
func (h *Handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseRecorder wraps http.ResponseWriter to capture status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}