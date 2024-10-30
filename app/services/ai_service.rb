class AIService
  VOICES = %w[
    alloy
    echo
    fable
    onyx
    nova
    shimmer
  ].freeze

  private attr_reader :config

  def initialize(config = OpenAIConfig)
    @config = config
  end

  def generate_twilio_audio(phrase, voice: "alloy")
    Rails.cache.fetch("ai:audio:#{voice}:#{phrase.parameterize}") do
      client.audio.speech(
        parameters: {
          model: "tts-1",
          input: phrase,
          voice:,
          response_format: "pcm"
        }
      ).then do |pcm|
        # downsamle from 20kHz to 8kHz
        samples = pcm.unpack("C*")
        pcm_samples = []
        (0..(samples.size - 1)).step(3) do |i|
          pcm_samples << samples[i]
        end
        pcm_samples
      end.then do |pcm_samples|
        # convert to ulaw
        G711.encode_ulaw(pcm_samples).pack("C*")
      end.then do |ulaw|
        Base64.strict_encode64(ulaw)
      end
    end
  end

  private

  def client = @client ||= OpenAI::Client.new(access_token: config.api_key, organization_id: config.organization_id)
end
