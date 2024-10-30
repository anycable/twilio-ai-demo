Rails.application.config.to_prepare do
  # Make sure the factory is set on reloading
  AnyCable.connection_factory = AnyCable::Rails::ConnectionFactory.new do
    map "/cable" do
      ApplicationConnection
    end
    map "/twilio" do
      Twilio::ApplicationConnection
    end
  end
end
