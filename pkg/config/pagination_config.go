package config

type PaginationConfig struct {
	DefaultPageSize int `yaml:"default_page_size"`
	MaxPageSize     int `yaml:"max_page_size"`
}
