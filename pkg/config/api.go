package config

type AppConfig struct {
	Info struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"app"`

	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	I18n struct {
		DefaultLanguage string `yaml:"default_language"`
		LocalesDir      string `yaml:"locales_dir"`
	} `yaml:"i18n"`
}

func (c *AppConfig) GetServerAddress() string {
	return "0.0.0.0:" + c.Server.Port
}
