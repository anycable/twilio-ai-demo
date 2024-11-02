class OpenAIConfig < ApplicationConfig
  attr_config :api_key, :organization_id, :prompt,
              realtime_enabled: true
end
