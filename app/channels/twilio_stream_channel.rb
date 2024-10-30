class TwilioStreamChannel < ApplicationChannel
  state_attr_accessor :ai_voice

  def subscribed
    broadcast_log "Media stream has started"

    self.ai_voice = AIService::VOICES.sample

    greeting = "Hi, I'm Aike. Here is my favourite Simpsons quote: #{Faker::TvShows::Simpsons.quote}"

    transmit_message(:greeting, greeting)
  end

  def unsubscribed
    broadcast_log "Media stream has stopped"
  end

  private

  def transmit_message(id, message)
    # This is an example of how you can send audio to the media stream from the
    # web app.
    transmit({
      event: "media",
      streamSid: params[:stream_sid],
      media: {
        payload: ai.generate_twilio_audio(message, voice: ai_voice),
      }
    })

    broadcast_log "> #{message}"

    # Mark message is required to keep track of the played audio.
    # Twilio will send its mark event with the same name as soon as the audio
    # has been played.
    transmit({
      event: "mark",
      streamSid: params[:stream_sid],
      mark: {
        name: id
      }
    })
  end

  def broadcast_log(msg) = twilio.broadcast_logs(params[:call_sid], msg)

  def twilio = @twilio ||= TwilioService.new

  def ai = @ai ||= AIService.new
end
