package config

import (
	"log"
	"os"

	"github.com/a8m/envsubst"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// LoadConfig reads the YAML file at path, expands ${ENV_VAR} references via
// envsubst, and unmarshals the result into a new T.
func LoadConfig[T any](path string) (*T, error) {
	data, err := envsubst.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := new(T)
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadEnvFile loads the .env file named by the ENV_FILE environment variable
// (defaults to ".env"), logging a warning when the file is absent.
func LoadEnvFile() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Printf(
			"Environment file not found, relying on system environment: %v",
			err,
		)
	}
}
