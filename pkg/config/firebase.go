package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type FirebaseConfig struct {
	ProjectID            string        `yaml:"project_id"`
	StorageBucket        string        `yaml:"storage_bucket"`
	CredentialsFile      string        `yaml:"credentials_file"`
	MaxFileSizeMB        int           `yaml:"max_file_size_mb"`
	AllowedExtensions    []string      `yaml:"allowed_extensions"`
	AvatarFolder         string        `yaml:"avatar_folder"`
	UploadTokenExpiresIn time.Duration `yaml:"upload_token_expires_in"`
}

type FirebaseConfigs struct {
	Dev  FirebaseConfig `yaml:"dev"`
	Prod FirebaseConfig `yaml:"prod"`
}

func LoadFirebaseConfig(env string) (*FirebaseConfig, error) {
	data, err := os.ReadFile("configs/firebase.yaml")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read firebase config: %w",
			err,
		)
	}

	var configs FirebaseConfigs
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf(
			"failed to parse firebase config: %w",
			err,
		)
	}

	if env == "" {
		env = "development"
	}

	var config FirebaseConfig
	switch env {
	case "production":
		config = configs.Prod
	default:
		config = configs.Dev
	}

	return &config, nil
}
