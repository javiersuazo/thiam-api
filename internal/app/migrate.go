//go:build migrate

package app

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts    = 20
	_defaultIntervalSec = 1
)

func init() {
	if enabled := os.Getenv("MIGRATION_ENABLED"); enabled == "false" {
		log.Printf("Migrate: disabled via MIGRATION_ENABLED=false")
		return
	}

	databaseURL, ok := os.LookupEnv("PG_URL")
	if !ok || len(databaseURL) == 0 {
		log.Fatalf("migrate: environment variable not declared: PG_URL")
	}

	attempts := getEnvInt("MIGRATION_RETRY_ATTEMPTS", _defaultAttempts)
	intervalSec := getEnvInt("MIGRATION_RETRY_INTERVAL_SEC", _defaultIntervalSec)
	interval := time.Duration(intervalSec) * time.Second

	var (
		err error
		m   *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", databaseURL)
		if err == nil {
			break
		}

		log.Printf("Migrate: postgres is trying to connect, attempts left: %d", attempts)
		time.Sleep(interval)
		attempts--
	}

	if err != nil {
		log.Fatalf("Migrate: postgres connect error: %s", err)
	}

	err = m.Up()
	defer m.Close()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Migrate: up error: %s", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	version, dirty, _ := m.Version()
	log.Printf("Migrate: up success (version: %d, dirty: %v)", version, dirty)
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
