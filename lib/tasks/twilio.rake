namespace :twilio do
  desc "Call the provided number"
  task :call, [:to_number] => :environment do |_task, args|
    twilio = TwilioService.new
    twilio.make_call(to: args[:to_number])
  end
end
