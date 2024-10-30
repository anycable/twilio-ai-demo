package config

import "github.com/palkan/twilio-ai-cable/pkg/twilio"

type Config struct {
	FakeRPC bool
	Twilio  *twilio.Config
}

func NewConfig() *Config {
	return &Config{
		FakeRPC: false,
		Twilio:  twilio.NewConfig(),
	}
}
