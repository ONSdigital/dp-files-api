package config

import (
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"time"

	"github.com/ONSdigital/dp-mongodb/v3/mongodb"

	"github.com/kelseyhightower/envconfig"
)

type MongoConfig = mongodb.MongoDriverConfig
type AuthConfig = authorisation.Config

// Config represents service configuration for dp-files-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	IsPublishing               bool          `envconfig:"IS_PUBLISHING"`
	MongoConfig
	KafkaConfig
	AuthConfig
}

// KafkaConfig contains the config required to connect to Kafka
type KafkaConfig struct {
	Addr                      []string `envconfig:"KAFKA_ADDR"                            json:"-"`
	ProducerMinBrokersHealthy int      `envconfig:"KAFKA_PRODUCER_MIN_BROKERS_HEALTHY"`
	Version                   string   `envconfig:"KAFKA_VERSION"`
	MaxBytes                  int      `envconfig:"KAFKA_MAX_BYTES"`
	SecProtocol               string   `envconfig:"KAFKA_SEC_PROTO"`
	SecCACerts                string   `envconfig:"KAFKA_SEC_CA_CERTS"`
	SecClientKey              string   `envconfig:"KAFKA_SEC_CLIENT_KEY"                  json:"-"`
	SecClientCert             string   `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	SecSkipVerify             bool     `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	StaticFilePublishedTopic  string   `envconfig:"STATIC_FILE_PUBLISHED_TOPIC"`
}

var cfg *Config

const MetadataCollection = "MetadataCollection"

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
		IsPublishing:               false,
		MongoConfig: MongoConfig{
			ClusterEndpoint:               "localhost:27017",
			Database:                      "files",
			Collections:                   map[string]string{MetadataCollection: "metadata"},
			IsStrongReadConcernEnabled:    false,
			IsWriteConcernMajorityEnabled: true,
			ConnectTimeout:                5 * time.Second,
			QueryTimeout:                  15 * time.Second,
			TLSConnectionConfig: mongodb.TLSConnectionConfig{
				IsSSL: false,
			},
		},
		KafkaConfig: KafkaConfig{
			Addr:                      []string{"kafka:9092"},
			ProducerMinBrokersHealthy: 1,
			Version:                   "2.6.1",
			MaxBytes:                  2000000,
			SecProtocol:               "",
			SecCACerts:                "",
			SecClientKey:              "",
			SecClientCert:             "",
			SecSkipVerify:             false,
			StaticFilePublishedTopic:  "static-file-published-v2",
		},
		AuthConfig: AuthConfig{
				Enabled:                             true,
				PermissionsAPIURL:                   "http://localhost:25400",
				IdentityWebKeySetURL:                "http://localhost:25600",
				PermissionsCacheUpdateInterval:      time.Minute * 5,
				PermissionsMaxCacheTime:             time.Minute * 15,
				PermissionsCacheExpiryCheckInterval: time.Second * 10,
				IdentityClientMaxRetries:            2,
				ZebedeeURL:                          "http://localhost:8082",
		},
	}

	return cfg, envconfig.Process("", cfg)
}
