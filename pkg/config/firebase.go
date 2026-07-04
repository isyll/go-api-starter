package config

import "time"

type FirebaseConfig struct {
	ProjectID            string        `yaml:"project_id"`
	StorageBucket        string        `yaml:"storage_bucket"`
	CredentialsFile      string        `yaml:"credentials_file"`
	MaxFileSizeMB        int           `yaml:"max_file_size_mb"`
	AllowedExtensions    []string      `yaml:"allowed_extensions"`
	AvatarFolder         string        `yaml:"avatar_folder"`
	UploadTokenExpiresIn time.Duration `yaml:"upload_token_expires_in"`
}
