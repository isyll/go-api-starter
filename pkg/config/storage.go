package config

// S3-compatible object store (MinIO in dev).
type StorageConfig struct {
	Endpoint      string `yaml:"endpoint"`
	AccessKey     string `yaml:"access_key"`
	SecretKey     string `yaml:"secret_key"`
	Bucket        string `yaml:"bucket"`
	UseSSL        bool   `yaml:"use_ssl"`
	PublicBaseURL string `yaml:"public_base_url"`
}

func (c *StorageConfig) Enabled() bool {
	return c != nil && c.Endpoint != ""
}
