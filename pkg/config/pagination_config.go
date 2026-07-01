package config

// PaginationConfig holds the default and maximum page sizes used by
// all list endpoints via the api.Paginate helper.
type PaginationConfig struct {
	// DefaultPageSize is the number of items returned when the client
	// does not supply a page_size parameter.
	DefaultPageSize int `yaml:"default_page_size"`
	// MaxPageSize is the upper bound on page_size to prevent runaway
	// queries.
	MaxPageSize int `yaml:"max_page_size"`
}
