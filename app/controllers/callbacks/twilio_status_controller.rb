module Callbacks
  # Handle Twilio webhooks
  class TwilioStatusController < ApplicationController
    # TODO: Check X-Twilio-Signature
    # See https://www.twilio.com/docs/usage/webhooks/getting-started-twilio-webhooks#validate-that-webhook-requests-are-coming-from-twilio
    # before_aciton :verify_twilio_request

    def create
      phone_call = Twilio::PhoneCall.new(
        sid: params[:CallSid],
        status: params[:CallStatus],
        from: params[:From],
        to: params[:To]
      )

      Rails.logger.debug "Twilio call status=#{phone_call.status} callSid=#{phone_call.sid} from=#{phone_call.from} to=#{phone_call.to}"

      # Broadcast an update
      Turbo::StreamsChannel.broadcast_prepend_to "phone_calls", target: "callList", partial: "phone_calls/phone_call", locals: {phone_call:}

      if phone_call.status == "ringing"
        return render plain: TwilioService.new.setup_stream_response, content_type: "text/xml"
      end

      head :ok
    end
  end
end
