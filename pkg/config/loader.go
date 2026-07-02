package config

import (
	"log"
	"os"

	"github.com/a8m/envsubst"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

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
