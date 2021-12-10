package config

import (
	"time"

	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

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
