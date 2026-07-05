package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isyll/go-grpc-starter/pkg/config"
)

type DummyConfig struct {
	Name    string `yaml:"name"`
	Version int    `yaml:"version"`
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_config.yaml")

	yamlContent := []byte(`
name: "AppTest"
version: 1
`)
	err := os.WriteFile(filePath, yamlContent, 0o644)
	require.NoError(t, err)

	// Load valid config
	cfg, err := config.LoadConfig[DummyConfig](filePath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "AppTest", cfg.Name)
	assert.Equal(t, 1, cfg.Version)
}

func TestLoadConfig_EnvSubst(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_config_env.yaml")

	yamlContent := []byte(`
name: "${APP_NAME:-DefaultName}"
version: 2
`)
	err := os.WriteFile(filePath, yamlContent, 0o644)
	require.NoError(t, err)

	// Test fallback
	cfg1, err := config.LoadConfig[DummyConfig](filePath)
	require.NoError(t, err)
	assert.Equal(t, "DefaultName", cfg1.Name)

	// Test env override
	t.Setenv("APP_NAME", "EnvName")
	cfg2, err := config.LoadConfig[DummyConfig](filePath)
	require.NoError(t, err)
	assert.Equal(t, "EnvName", cfg2.Name)
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := config.LoadConfig[DummyConfig]("non_existent_file.yaml")
	assert.Error(t, err)
}

func TestLoadConfig_InvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_config_invalid.yaml")

	yamlContent := []byte(`
name: "AppTest"
version: NOT_A_NUMBER
	invalid_yaml: {
`)
	err := os.WriteFile(filePath, yamlContent, 0o644)
	require.NoError(t, err)

	_, err = config.LoadConfig[DummyConfig](filePath)
	assert.Error(t, err)
}

func TestLoadEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.test")

	err := os.WriteFile(
		envPath,
		[]byte("TEST_ENV_KEY=test_value\n"),
		0o644,
	)
	require.NoError(t, err)

	t.Setenv("ENV_FILE", envPath)
	t.Cleanup(func() { _ = os.Unsetenv("TEST_ENV_KEY") })

	config.LoadEnvFile()

	val := os.Getenv("TEST_ENV_KEY")
	assert.Equal(t, "test_value", val)
}
