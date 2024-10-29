class TwilioService
  private attr_reader :config

  def initialize(config = TwilioConfig)
    @config = config
  end

  def setup_stream_response
    Twilio::TwiML::VoiceResponse.new do |r|
      r.say(message: "Hey! Let's see what's on your plate. Connecting you to an agent...")
      r.connect do
        _1.stream(url: config.stream_callback)
      end
      r.say(message: "I'm sorry, I cannot connect you at this time.")
    end.to_s
  end

  def broadcast_logs(call_sid, msg)
    Turbo::StreamsChannel.broadcast_append_to(
      "phone_calls",
      target: "call-#{call_sid}-logs",
      partial: "phone_calls/phone_call_log",
      locals: {text: msg}
    )
  end

  private

  def client = @client ||= Twilio::REST::Client.new(config.account_sid, config.auth_token)
end
