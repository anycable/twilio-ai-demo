class TwilioStreamChannel < ApplicationChannel
  def subscribed
    TwilioService.new.broadcast_logs(params[:call_sid], "Media stream has started")

    # transmit({
    #   event: "media",
    #   streamSid: params[:stream_sid],
    #   media: {
    #     payload:
    #   }
    # })

    # transmit({
    #   event: "mark",
    #   streamSid: params[:stream_sid],
    #   mark: {
    #     name: "test-mark"
    #   }
    # })
  end

  def unsubscribed
    TwilioService.new.broadcast_logs(params[:call_sid], "Media stream has stopped")
  end
end
