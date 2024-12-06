module Twilio
  class MediaStreamChannel < ApplicationChannel
    include OpenAITools

    state_attr_accessor :ai_voice

    def subscribed
      broadcast_call_status "active"

      broadcast_log "Media stream has started"

      self.ai_voice = AIService::VOICES.sample

      greeting =
        if OpenAIConfig.realtime_enabled?
          "Hi, I'm #{ai_voice.humanize}. " \
          "I can tell you about your planned tasks and " \
          "help to you manage them. What would you like to do?"
        else
          "Hi, I'm #{ai_voice.humanize}. " \
          "Press 1 to check tasks for today. " \
          "Press 2 to check tasks for tomorrow. " \
          "Press 3 to check tasks for this week."
        end

      transmit_message(:greeting, greeting)
    end

    def handle_dtmf(data)
      digit = data["digit"].to_i
      broadcast_log "< Pressed ##{digit}"

      todos, period =
        case digit
        when 1 then [Todo.incomplete.where(deadline: Date.current.all_day), "today"]
        when 2 then [Todo.incomplete.where(deadline: Date.tomorrow.all_day), "tomorrow"]
        when 3 then [Todo.incomplete.where(deadline: Date.current.all_week), "this week"]
        end

      return unless todos

      phrase = if todos.any?
        "Here is what you have for #{period}:\n#{todos.map(&:description).join(",")}"
      else
        "You don't have any tasks for #{period}"
      end

      transmit_message(:"dtmf_response_#{digit}", phrase)
    end

    # OpenAI tools
    def configure_openai
      config = OpenAIConfig
      return unless config.realtime_enabled?

      api_key = config.api_key
      voice = ai_voice
      prompt = config.prompt

      # Get tools configuration from this class
      # NOTE: with pass tools configuration as JSON to not deal with serialization/deserialization at the server side.
      tools = self.class.openai_tools_schema.to_json

      reply_with("openai.configuration", {api_key:, voice:, prompt:, tools:})
    end

    def handle_transcript(data)
      direction = data["role"] == "user" ? "<" : ">"

      broadcast_log "#{direction} #{data["text"]}", id: data["id"]
    end

    def handle_function_call(data)
      name = data["name"].to_sym
      args = JSON.parse(data["arguments"], symbolize_names: true)

      return unless self.class.openai_tools.include?(name)

      broadcast_log "# Invoke: #{name}(#{data["arguments"]})"

      result = public_send(name, **args)

      reply_with("openai.function_call_result", result)
    end

    # Fetch user's tasks for a given period of time.
    # @rbs (period: (:today | :tomorrow | :week)) -> Array[Todo]
    tool def get_tasks(period:)
      range = case period
      when "today"
        Date.current.all_day
      when "tomorrow"
        Date.tomorrow.all_day
      when "week"
        Date.current.all_week
      end

      {todos: Todo.incomplete.where(deadline: range).as_json(only: [:id, :deadline, :description])}
    end

    # Create a new task for a specified date
    # @rbs (deadline: Date, description: String) -> {status: (:created | :failed), ?todo: Todo}
    tool def create_task(deadline:, description:)
      todo = Todo.new(deadline:, description:)
      if todo.save
        {status: :created, todo: todo.as_json(only: [:id, :deadline, :description])}
      else
        {status: :failed, message: todo.errors.full_messages.join(", ")}
      end
    end

    # Mark a task as completed
    # @rbs (id: Integer) -> {status: (:completed | :failed), ?message: String}
    tool def complete_task(id:)
      todo = Todo.find_by(id: id)
      if todo
        todo.update!(completed: true)
        {status: :completed}
      else
        {status: :failed, message: "Task not found"}
      end
    end

    # Deleate a task
    # @rbs (id: Integer) -> {status: (:completed | :failed), ?message: String}
    tool def delete_task(id:)
      todo = Todo.find_by(id: id)
      if todo
        todo.destroy!
        {status: :completed}
      else
        {status: :failed, message: "Task not found"}
      end
    end

    def unsubscribed
      broadcast_log "Media stream has stopped"

      broadcast_call_status "completed"
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

    def broadcast_call_status(status)
      phone_call = Twilio::PhoneCall.new(
        sid: call_sid,
        from: nil,
        to: nil
      )

      twilio.broadcast_phone_call(phone_call, status)
    end

    def broadcast_log(...) = twilio.broadcast_logs(call_sid, ...)

    def twilio = @twilio ||= TwilioService.new

    def ai = @ai ||= AIService.new
  end
end
