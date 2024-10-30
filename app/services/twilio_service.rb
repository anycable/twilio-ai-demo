class TwilioService
  private attr_reader :config

  def initialize(config = TwilioConfig)
    @config = config
  end

  def make_call(to:, phrase:, timeout: 30)
    call = client.calls.create(
      twiml: setup_stream_response(phrase),
      to: to,
      from: config.phone_number,
      timeout: timeout
    )
    call.sid
  end

  DEFAULT_WELCOME_PHRASE = "Hey! Let me connect you to our AI agent..."

  def setup_stream_response(message = DEFAULT_WELCOME_PHRASE)
    Twilio::TwiML::VoiceResponse.new do |r|
      r.say(message:)
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
