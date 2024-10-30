module Twilio
  class MediaStreamChannel < ApplicationChannel
    state_attr_accessor :ai_voice

    def subscribed
      broadcast_log "Media stream has started"

      self.ai_voice = AIService::VOICES.sample

      greeting = "Hi, I'm #{ai_voice.humanize}. Here is my favourite Simpsons quote: #{Faker::TvShows::Simpsons.quote}. How can I help you today?"

      transmit_message(:greeting, greeting)
    end

    def handle_dtmf(data)
      broadcast_log "< Pressed ##{data["digit"]}"
    end

    # OpenAI tools
    def configure_openai
      config = OpenAIConfig

      api_key = config.api_key
      voice = ai_voice
      prompt = config.prompt

      reply_with("openai.configuration", {api_key:, voice:, prompt:})
    end

    def handle_transcript(data)
      direction = data["role"] == "user" ? "<" : ">"

      broadcast_log "#{direction} #{data["text"]}", id: data["id"]
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
        streamSid: stream_sid,
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
        streamSid: stream_sid,
        mark: {
          name: id
        }
      })
    end

    def broadcast_log(...) = twilio.broadcast_logs(call_sid, ...)

    def twilio = @twilio ||= TwilioService.new

    def ai = @ai ||= AIService.new
  end
end
