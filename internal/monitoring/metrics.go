package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the notification service
type Metrics struct {
	NotificationsSent          *prometheus.CounterVec
	NotificationsFailed        *prometheus.CounterVec
	NotificationsDelivered     *prometheus.CounterVec
	NotificationLatency        *prometheus.HistogramVec
	ChannelProcessingDuration  *prometheus.HistogramVec
	QueueSize                  prometheus.Gauge
	ActiveConnections          prometheus.Gauge
	DatabaseConnections        *prometheus.GaugeVec
	RetryCount                 *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	metrics := &Metrics{
		NotificationsSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notifications_sent_total",
				Help: "Total number of notifications sent",
			},
			[]string{"channel", "status"},
		),
		NotificationsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notifications_failed_total",
				Help: "Total number of failed notifications",
			},
			[]string{"channel", "error_type"},
		),
		NotificationsDelivered: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notifications_delivered_total",
				Help: "Total number of delivered notifications",
			},
			[]string{"channel"},
		),
		NotificationLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "notification_processing_duration_seconds",
				Help:    "Time taken to process notifications",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"channel", "operation"},
		),
		ChannelProcessingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "channel_processing_duration_seconds",
				Help:    "Time taken by channels to send notifications",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"channel"},
		),
		QueueSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "notification_queue_size",
				Help: "Current size of the notification queue",
			},
		),
		ActiveConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_connections",
				Help: "Number of active connections to the service",
			},
		),
		DatabaseConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of active database connections",
			},
			[]string{"database", "state"},
		),
		RetryCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_retries_total",
				Help: "Total number of notification retries",
			},
			[]string{"channel", "retry_reason"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(
		metrics.NotificationsSent,
		metrics.NotificationsFailed,
		metrics.NotificationsDelivered,
		metrics.NotificationLatency,
		metrics.ChannelProcessingDuration,
		metrics.QueueSize,
		metrics.ActiveConnections,
		metrics.DatabaseConnections,
		metrics.RetryCount,
	)

	return metrics
}

// RecordNotificationSent records a sent notification
func (m *Metrics) RecordNotificationSent(channel, status string) {
	m.NotificationsSent.WithLabelValues(channel, status).Inc()
}

// RecordNotificationFailed records a failed notification
func (m *Metrics) RecordNotificationFailed(channel, errorType string) {
	m.NotificationsFailed.WithLabelValues(channel, errorType).Inc()
}

// RecordNotificationDelivered records a delivered notification
func (m *Metrics) RecordNotificationDelivered(channel string) {
	m.NotificationsDelivered.WithLabelValues(channel).Inc()
}

// RecordProcessingDuration records processing duration
func (m *Metrics) RecordProcessingDuration(channel, operation string, duration float64) {
	m.NotificationLatency.WithLabelValues(channel, operation).Observe(duration)
}

// RecordChannelDuration records channel processing duration
func (m *Metrics) RecordChannelDuration(channel string, duration float64) {
	m.ChannelProcessingDuration.WithLabelValues(channel).Observe(duration)
}

// SetQueueSize sets the current queue size
func (m *Metrics) SetQueueSize(size float64) {
	m.QueueSize.Set(size)
}

// IncrementActiveConnections increments active connections
func (m *Metrics) IncrementActiveConnections() {
	m.ActiveConnections.Inc()
}

// DecrementActiveConnections decrements active connections
func (m *Metrics) DecrementActiveConnections() {
	m.ActiveConnections.Dec()
}

// SetDatabaseConnections sets database connection metrics
func (m *Metrics) SetDatabaseConnections(database, state string, count float64) {
	m.DatabaseConnections.WithLabelValues(database, state).Set(count)
}

// RecordRetry records a notification retry
func (m *Metrics) RecordRetry(channel, reason string) {
	m.RetryCount.WithLabelValues(channel, reason).Inc()
}

// Handler returns the Prometheus metrics HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}