module Twilio
  # Handle Twilio status webhooks
  class StatusesController < ApplicationController
    skip_before_action :verify_authenticity_token

    # TODO: Check X-Twilio-Signature
    # See https://www.twilio.com/docs/usage/webhooks/getting-started-twilio-webhooks#validate-that-webhook-requests-are-coming-from-twilio
    # before_aciton :verify_twilio_request

    def create
      twilio = TwilioService.new
      status = params[:CallStatus]

      phone_call = Twilio::PhoneCall.new(
        sid: params[:CallSid],
        from: params[:From],
        to: params[:To]
      )

      Rails.logger.debug "Twilio call status=#{status} callSid=#{phone_call.sid} from=#{phone_call.from} to=#{phone_call.to}"

      # Broadcast an update
      twilio.broadcast_phone_call(phone_call, status)

      if status == "ringing"
        return render plain: twilio.setup_stream_response, content_type: "text/xml"
      end

      head :ok
    end
  end
end
