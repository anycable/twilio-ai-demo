module Twilio
  class ApplicationConnection < ActionCable::Connection::Base
    identified_by :call_sid, :stream_sid

    def connect
      raise "Must not be called; AnyCable server should perform authentication"
    end
  end
end
