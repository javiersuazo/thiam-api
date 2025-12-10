package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type (
	// Config -.
	Config struct {
		App     App
		HTTP    HTTP
		Log     Log
		PG      PG
		GRPC    GRPC
		RMQ     RMQ
		Outbox  Outbox
		NATS    NATS
		Metrics Metrics
		Swagger Swagger
		JWT     JWT
	}

	// App -.
	App struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	// HTTP -.
	HTTP struct {
		Port           string `env:"HTTP_PORT,required"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	// Log -.
	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// PG -.
	PG struct {
		PoolMax int    `env:"PG_POOL_MAX,required"`
		URL     string `env:"PG_URL,required"`
	}

	// GRPC -.
	GRPC struct {
		Port string `env:"GRPC_PORT,required"`
	}

	// RMQ -.
	RMQ struct {
		ServerExchange string `env:"RMQ_RPC_SERVER,required"`
		ClientExchange string `env:"RMQ_RPC_CLIENT,required"`
		URL            string `env:"RMQ_URL,required"`
		EventExchange  string `env:"RMQ_EVENT_EXCHANGE" envDefault:"domain.events"`
	}

	// Outbox -.
	Outbox struct {
		Enabled      bool `env:"OUTBOX_ENABLED" envDefault:"true"`
		PollInterval int  `env:"OUTBOX_POLL_INTERVAL_MS" envDefault:"1000"`
		BatchSize    int  `env:"OUTBOX_BATCH_SIZE" envDefault:"100"`
		MaxRetries   int  `env:"OUTBOX_MAX_RETRIES" envDefault:"5"`
	}

	// NATS -.
	NATS struct {
		ServerExchange string `env:"NATS_RPC_SERVER,required"`
		URL            string `env:"NATS_URL,required"`
	}

	// Metrics -.
	Metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	}

	// Swagger -.
	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}

	// JWT -.
	JWT struct {
		Secret string `env:"JWT_SECRET,required"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
