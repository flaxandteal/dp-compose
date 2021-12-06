package config

import (
	"time"

	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

// KafkaTLSProtocolFlag informs service to use TLS protocol for kafka
const KafkaTLSProtocolFlag = "TLS"

// Config represents the app configuration
type Config struct {
	ImportAPIAddr              string        `envconfig:"IMPORT_API_ADDR"`
	DatasetAPIAddr             string        `envconfig:"DATASET_API_ADDR"`
	DatasetAPIMaxWorkers       int           `envconfig:"DATASET_API_MAX_WORKERS"` // maximum number of concurrent go-routines requesting items to datast api at the same time
	DatasetAPIBatchSize        int           `envconfig:"DATASET_API_BATCH_SIZE"`  // maximum size of a response by dataset api when requesting items in batches
	ShutdownTimeout            time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"                   json:"-"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	KafkaConfig                KafkaConfig
}

// KafkaConfig contains the config required to connect to Kafka
type KafkaConfig struct {
	Addr             []string `envconfig:"KAFKA_ADDR"                            json:"-"`
	Version          string   `envconfig:"KAFKA_VERSION"`
	OffsetOldest     bool     `envconfig:"KAFKA_OFFSET_OLDEST"`
	NumWorkers       int      `envconfig:"KAFKA_NUM_WORKERS"`
	MaxBytes         int      `envconfig:"KAFKA_MAX_BYTES"`
	SecProtocol      string   `envconfig:"KAFKA_SEC_PROTO"`
	SecCACerts       string   `envconfig:"KAFKA_SEC_CA_CERTS"`
	SecClientKey     string   `envconfig:"KAFKA_SEC_CLIENT_KEY"                  json:"-"`
	SecClientCert    string   `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	SecSkipVerify    bool     `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	ExportStartTopic string   `envconfig:"KAFKA_TOPIC_CANTABULAR_EXPORT_START"`
}

// NewConfig creates the config object
func NewConfig() (*Config, error) {
	cfg := Config{
		ServiceAuthToken:           "AB0A5CFA-3C55-4FA8-AACC-F98039BED0AC",
		ImportAPIAddr:              "http://localhost:21800",
		DatasetAPIAddr:             "http://localhost:22000",
		DatasetAPIMaxWorkers:       100,
		DatasetAPIBatchSize:        1000,
		ShutdownTimeout:            5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		KafkaConfig: KafkaConfig{
			Addr:             []string{"localhost:9092"},
			Version:          "1.0.2",
			OffsetOldest:     true,
			NumWorkers:       1,
			MaxBytes:         2000000,
			SecProtocol:      "",
			SecCACerts:       "",
			SecClientKey:     "",
			SecClientCert:    "",
			SecSkipVerify:    false,
			ExportStartTopic: "cantabular-export-start",
		},
	}
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	cfg.ServiceAuthToken = "Bearer " + cfg.ServiceAuthToken

	return &cfg, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
