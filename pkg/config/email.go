package config

type EmailConfig struct {
	Email struct {
		APIKey  string `yaml:"api_key"`
		Senders struct {
			NoReply  SenderInfo `yaml:"noreply"`
			Security SenderInfo `yaml:"security"`
			News     SenderInfo `yaml:"news"`
		} `yaml:"senders"`
		Worker struct {
			Concurrency int `yaml:"concurrency"`
			RetryMax    int `yaml:"retry_max"`
		} `yaml:"worker"`
		Templates struct {
			BasePath        string `yaml:"base_path"`
			DefaultLanguage string `yaml:"default_language"`
		} `yaml:"templates"`
		Batch struct {
			MaxSize int `yaml:"max_size"`
		} `yaml:"batch"`
	} `yaml:"email"`
}

type SenderInfo struct {
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
}
