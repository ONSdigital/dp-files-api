package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// MongoConfig contains the config required to connect to MongoDB.
type MongoConfig struct {
	URI                string        `envconfig:"MONGODB_BIND_ADDR"   json:"-"`
	Collection         string        `envconfig:"MONGODB_FILES_COLLECTION"`
	Database           string        `envconfig:"MONGODB_FILES_DATABASE"`
	Username           string        `envconfig:"MONGODB_USERNAME"    json:"-"`
	Password           string        `envconfig:"MONGODB_PASSWORD"    json:"-"`
	IsSSL              bool          `envconfig:"MONGODB_IS_SSL"`
	EnableReadConcern  bool          `envconfig:"MONGODB_ENABLE_READ_CONCERN"`
	EnableWriteConcern bool          `envconfig:"MONGODB_ENABLE_WRITE_CONCERN"`
	QueryTimeout       time.Duration `envconfig:"MONGODB_QUERY_TIMEOUT"`
	ConnectionTimeout  time.Duration `envconfig:"MONGODB_CONNECT_TIMEOUT"`
}


// Config represents service configuration for dp-files-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	MongoConfig
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:26900",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		MongoConfig: MongoConfig{
			URI:                "localhost:27017",
			Database:           "datasets",
			Collection:         "datasets",
			QueryTimeout:       15 * time.Second,
			ConnectionTimeout:  5 * time.Second,
			EnableWriteConcern: true,
		},
	}

	return cfg, envconfig.Process("", cfg)
}
