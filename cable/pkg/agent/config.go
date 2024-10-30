package agent

type Config struct {
	URL    string
	Key    string
	Model  string
	Voice  string
	Prompt string
	// we just pass them as is to the AI
	Tools interface{}
}

func NewConfig(key string) *Config {
	return &Config{
		URL:   "wss://api.openai.com/v1/realtime",
		Key:   key,
		Model: "gpt-4o-realtime-preview-2024-10-01",
		Voice: "alloy",
	}
}
