package config

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080",
	}
}
