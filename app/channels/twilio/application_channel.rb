module Twilio
  class ApplicationChannel < ActionCable::Channel::Base
    # This state is used to carry messages from Rails to AnyCable.
    state_attr_accessor :anycable_response

    delegate :call_sid, :stream_sid, to: :connection

    # Send a response to AnyCable server (not to the client directly)
    # Use this method to send commands or configuration.
    def reply_with(event, data)
      # Here we use the channel state feature of AnyCable.
      # Server reads the message from the state and clears it upon retrieval.
      self.anycable_response = {event:, data:}
    end
  end
end
