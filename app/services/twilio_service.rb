class TwilioService
  private attr_reader :config

  def initialize(config = TwilioConfig)
    @config = config
  end

  def make_call(to:, phrase: nil, timeout: 60)
    call = client.calls.create(
      twiml: setup_stream_response(phrase),
      to: to,
      from: config.phone_number,
      timeout: timeout
    )
    call.sid
  end

  def setup_stream_response(message = nil)
    Twilio::TwiML::VoiceResponse.new do |r|
      r.say(message:) if message
      r.connect do
        _1.stream(url: config.stream_callback)
      end
      r.say(message: "I'm sorry, I cannot connect you at this time.")
    end.to_s
  end

  def broadcast_logs(call_sid, msg, id: SecureRandom.hex(4))
    Turbo::StreamsChannel.broadcast_append_to(
      "phone_calls",
      target: "call-#{call_sid}-logs",
      partial: "phone_calls/phone_call_log",
      locals: {text: msg, id:}
    )
  end

  def broadcast_phone_call(phone_call, status)
    Turbo::StreamsChannel.broadcast_prepend_to(
      "phone_calls",
      target: "callList",
      partial: "phone_calls/phone_call",
      locals: {phone_call:, status:}
    )
  end

  private

  def client = @client ||= Twilio::REST::Client.new(config.account_sid, config.auth_token)
end
