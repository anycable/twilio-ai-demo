class TwilioConfig < ApplicationConfig
  attr_config :account_sid, :auth_token,
    :phone_number,
    :status_callback,
    :stream_callback
end
