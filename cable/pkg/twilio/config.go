package twilio

type Config struct {
	AccountSID string
}

func NewConfig() *Config {
	return &Config{}
}
