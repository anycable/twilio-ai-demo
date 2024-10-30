AnyCable.configure_server do
  AnyCable.connection_factory = AnyCable::Rails::ConnectionFactory.new do
    map "/cable" do
      ApplicationConnection
    end
    map "/twilio" do
      Twilio::ApplicationConnection
    end
  end
end
