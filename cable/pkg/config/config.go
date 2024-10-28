package config

type Config struct {
	FakeRPC bool
}

func NewConfig() *Config {
	return &Config{
		FakeRPC: false,
	}
}
