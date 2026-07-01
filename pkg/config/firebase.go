package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// FirebaseConfig holds the project credentials and storage settings
// for a single Firebase environment (dev or prod).
type FirebaseConfig struct {
	// ProjectID is the Google Cloud project identifier.
	ProjectID string `yaml:"project_id"`
	// StorageBucket is the Cloud Storage bucket used for avatar uploads.
	StorageBucket string `yaml:"storage_bucket"`
	// CredentialsFile is the path to the service-account JSON key file.
	CredentialsFile string `yaml:"credentials_file"`
	// MaxFileSizeMB is the per-upload size cap in megabytes.
	MaxFileSizeMB int `yaml:"max_file_size_mb"`
	// AllowedExtensions is the list of permitted file suffixes
	// (e.g. ["jpg", "png", "webp"]).
	AllowedExtensions []string `yaml:"allowed_extensions"`
	// AvatarFolder is the Storage prefix used for user avatar objects.
	AvatarFolder string `yaml:"avatar_folder"`
	// UploadTokenExpiresIn is how long a signed upload URL remains valid.
	UploadTokenExpiresIn time.Duration `yaml:"upload_token_expires_in"`
}

// FirebaseConfigs groups dev and prod Firebase credentials so
// LoadFirebaseConfig can select the right set at startup.
type FirebaseConfigs struct {
	Dev  FirebaseConfig `yaml:"dev"`
	Prod FirebaseConfig `yaml:"prod"`
}

// LoadFirebaseConfig reads configs/firebase.yaml and returns the
// FirebaseConfig matching env ("production" or any other string for
// dev). An empty env defaults to dev.
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
