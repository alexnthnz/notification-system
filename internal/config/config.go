package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config holds all configuration for the notification service
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	API      APIConfig      `mapstructure:"api"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Channels ChannelsConfig `mapstructure:"channels"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

// APIConfig holds API server configuration
type APIConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	GRPCPort int `mapstructure:"grpc_port"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}

// ChannelsConfig holds third-party provider configurations
type ChannelsConfig struct {
	SendGrid SendGridConfig `mapstructure:"sendgrid"`
	Twilio   TwilioConfig   `mapstructure:"twilio"`
	Firebase FirebaseConfig `mapstructure:"firebase"`
}

// SendGridConfig holds SendGrid email configuration
type SendGridConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// TwilioConfig holds Twilio SMS configuration
type TwilioConfig struct {
	AccountSID string `mapstructure:"account_sid"`
	AuthToken  string `mapstructure:"auth_token"`
}

// FirebaseConfig holds Firebase push notification configuration
type FirebaseConfig struct {
	CredentialsPath string `mapstructure:"credentials_path"`
}

// MetricsConfig holds monitoring configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set default values
	setDefaults()

	// Read from environment variables
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		log.Println("Config file not found, using environment variables and defaults")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.database", "notifications")
	viper.SetDefault("database.ssl_mode", "disable")

	// Redis defaults
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Kafka defaults
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topic", "notifications")

	// API defaults
	viper.SetDefault("api.host", "0.0.0.0")
	viper.SetDefault("api.port", 8080)
	viper.SetDefault("api.grpc_port", 9090)

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", 9091)
	viper.SetDefault("metrics.path", "/metrics")

	// Map environment variables
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.database", "DB_NAME")
	viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("kafka.brokers", "KAFKA_BROKERS")
	viper.BindEnv("auth.jwt_secret", "JWT_SECRET")
	viper.BindEnv("channels.sendgrid.api_key", "SENDGRID_API_KEY")
	viper.BindEnv("channels.twilio.account_sid", "TWILIO_ACCOUNT_SID")
	viper.BindEnv("channels.twilio.auth_token", "TWILIO_AUTH_TOKEN")
	viper.BindEnv("channels.firebase.credentials_path", "FIREBASE_CREDENTIALS_PATH")
}