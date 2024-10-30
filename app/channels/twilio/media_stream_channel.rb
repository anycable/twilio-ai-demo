module Twilio
  class MediaStreamChannel < ApplicationChannel
    state_attr_accessor :ai_voice

    def subscribed
      broadcast_call_status "active"

      broadcast_log "Media stream has started"

      self.ai_voice = AIService::VOICES.sample

      greeting = "Hi, I'm #{ai_voice.humanize}. " \
        "I can tell you about your planned tasks and " \
        "help to you manage them. What would you like to do?"

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

      # Generate tools configuration
      # TODO: That could be done via metaprogramming based on the #handle_function_call contents.
      # NOTE: with pass tools configuration as JSON to not deal with serialization/deserialization at the server side.
      tools = [
        {
          type: "function",
          name: "get_tasks",
          description: "Fetch user's tasks for a given period of time",
          parameters: {
            type: "object",
            properties: {
              period: {
                type: "string",
                enum: ["today", "tomorrow", "week"]
              }
            },
            required: ["period"]
          }
        },
        {
          type: "function",
          name: "create_task",
          description: "Create a new task for a specified date",
          parameters: {
            type: "object",
            properties: {
              date: {
                type: "string",
                format: "date"
              },
              description: {
                type: "string"
              }
            },
            required: ["date", "description"]
          }
        },
        {
          type: "function",
          name: "complete_task",
          description: "Mark a task as completed",
          parameters: {
            type: "object",
            properties: {
              id: {
                type: "integer"
              }
            },
            required: ["id"]
          }
        }
      ].to_json

      reply_with("openai.configuration", {api_key:, voice:, prompt:, tools:})
    end

    def handle_transcript(data)
      direction = data["role"] == "user" ? "<" : ">"

      broadcast_log "#{direction} #{data["text"]}", id: data["id"]
    end

    def handle_function_call(data)
      name = data["name"]
      args = JSON.parse(data["arguments"], symbolize_names: true)

      broadcast_log "# Invoke: #{name}(#{data["arguments"]})"

      case [name, args]
      in "get_tasks", {period: "today" | "tomorrow" | "week" => period}
        range = case period
        when "today"
          Date.current.all_day
        when "tomorrow"
          Date.tomorrow.all_day
        when "week"
          Date.current.all_week
        end

        todos = Todo.incomplete.where(deadline: range).as_json(only: [:id, :deadline, :description])

        reply_with("openai.function_call_result", {todos:})
      in "create_task", {date: String => deadline, description: String => description}
        todo = Todo.new(deadline:, description:)
        if todo.save
          reply_with("openai.function_call_result", {status: :created, todo: todo.as_json(only: [:id, :deadline, :description])})
        else
          reply_with("openai.function_call_result", {status: :failed, message: todo.errors.full_messages.join(", ")})
        end
      in "complete_task", {id: Integer => id}
        todo = Todo.find_by(id: id)
        if todo
          todo.update!(completed: true)
          reply_with("openai.function_call_result", {status: :completed})
        else
          reply_with("openai.function_call_result", {status: :failed, message: "Task not found"})
        end
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
