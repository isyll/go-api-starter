package config

// GatewayConfig configures the optional HTTP/JSON transcoding gateway. It is
// off by default: the gateway binary refuses to start unless Enabled is true.
type GatewayConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ListenAddr string `yaml:"listen_addr"`
	// UpstreamAddr is the gRPC server to proxy to. Empty means the local
	// server on the configured server port.
	UpstreamAddr string `yaml:"upstream_addr"`
}
