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
	_ = godotenv.Load()

	values := map[string]*string{
		"DB_HOST":                nil,
		"DB_PORT":                nil,
		"APTORA_DB_NAME":         nil,
		"APTORA_DB_USER":         nil,
		"APTORA_DB_PASSWORD":     nil,
		"EXTENSIONS_DB_NAME":     nil,
		"EXTENSIONS_DB_USER":     nil,
		"EXTENSIONS_DB_PASSWORD": nil,
	}

	for key := range values {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			return Settings{}, fmt.Errorf("environment variable %s is required", key)
		}
		v := value
		values[key] = &v
	}

	return Settings{
		DBHost:               *values["DB_HOST"],
		DBPort:               *values["DB_PORT"],
		AptoraDBName:         *values["APTORA_DB_NAME"],
		AptoraDBUser:         *values["APTORA_DB_USER"],
		AptoraDBPassword:     *values["APTORA_DB_PASSWORD"],
		ExtensionsDBName:     *values["EXTENSIONS_DB_NAME"],
		ExtensionsDBUser:     *values["EXTENSIONS_DB_USER"],
		ExtensionsDBPassword: *values["EXTENSIONS_DB_PASSWORD"],
	}, nil
}
