package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Settings contains all required configuration values for the backend server.
type Settings struct {
	DBHost               string
	DBPort               string
	AptoraDBName         string
	AptoraDBUser         string
	AptoraDBPassword     string
	ExtensionsDBName     string
	ExtensionsDBUser     string
	ExtensionsDBPassword string
}

// Load loads environment variables from the runtime environment. When a local
// .env file is present (development mode), it is loaded automatically.
func Load() (Settings, error) {
	// Try to load .env file - check parent directory first (for when running from backend/),
	// then current directory. If both fail, continue without .env (production may use actual env vars).
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			// Neither file exists - that's OK, we may be using actual environment variables
		}
	}

	missing := []string{}
	get := func(k string) string {
		v := strings.TrimSpace(os.Getenv(k))
		if v == "" {
			missing = append(missing, k)
		}
		return v
	}

	settings := Settings{
		DBHost:               get("DB_HOST"),
		DBPort:               get("DB_PORT"),
		AptoraDBName:         get("APTORA_DB_NAME"),
		AptoraDBUser:         get("APTORA_DB_USER"),
		AptoraDBPassword:     get("APTORA_DB_PASSWORD"),
		ExtensionsDBName:     get("EXTENSIONS_DB_NAME"),
		ExtensionsDBUser:     get("EXTENSIONS_DB_USER"),
		ExtensionsDBPassword: get("EXTENSIONS_DB_PASSWORD"),
	}

	if len(missing) > 0 {
		return Settings{}, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return settings, nil
}
