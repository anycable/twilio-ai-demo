# AnyCable x Twilio Media Streams x OpenAI Realtime

This application demonstrates how to use [AnyCable](https://anycable.io) to interact with voice calls via Twilio Streams
and OpenAI Realtime API.

> [!TIP] 
> Read the blog post to learn more about how AnyCable helps to bring phone calls and AI together: [Hey, AnyCable speaking! Needing help with a Twilio-OpenAI connection?](https://evilmartians.com/chronicles/anycable-speaking-needing-help-with-a-twilio-openai-connection).

The app consists of two parts:

- A Ruby on Rails application that provides a web UI to ToDo items.
- A Go application built with AnyCable that handles Twilio Streams and OpenAI interactions in a logic-agnostic way.

The RoR app is the primary (and only) source of truth providing credentials, instructions,and tools
for the Go application (i.e., Twilio and OpenAI). Check out the `app/channels/twilio/media_stream_channel.rb` file to see the voice UX implementation.

## Requirements

- Ruby 3.3 for the application server.
- Go 1.23 to run AnyCable.
- Twilio and OpenAI accounts (see below).

## Configuration

### Twilio / OpenAI setup

You must obtain a phone number from Twilio as well as your account credentials.
Similarly, generate an API key for OpenAI.

Then, you can store them in the local configuration files as follows:

```yml
# config/twilio.local.yml`
phone_number: <you phone number>
account_sid: <your account SID>
auth_token: <your auth token>

# config/openai.local.yml
api_key: <your key>
```

Alternatively, you can use the corresponding environment variables
(`TWILIO_PHONE_NUMBER`, `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `OPENAI_API_KEY`) or [local credentials](https://github.com/palkan/anyway_config#local-files).

## Running the app

The web app is built with Ruby on Rails. To run it, you need to install the dependencies:

```sh
bin/setup
```

The command above also builds the AnyCable projects stored at the `cable/` directory.

Now, you can run it locally as follows:

```sh
bin/dev
```

Then, you can visit the app at [localhost:3000](http://localhost:3000). You can manage ToDo items there.

To interact with the app using a phone, you must make the realtime server (running on the `:8080`) by default
to be accessible from the Internet as well as your Rails server. You can use [ngrok](https://ngrok.com) for that:

```sh
# for AnyCable server
ngrok http 8080
```

Use the generated URL for 8080 as the `TWILIO_STREAM_CALLBACK` environment variable:

```sh
TWILIO_STREAM_CALLBACK=https://<your-ngrok-id>.ngrok.io bin/dev
```

(You can also add the url to the `config/twilio.local.yml` file).

Now, you can initiate a call to the provided phone number running the following Rake command:

```sh
TWILIO_STREAM_CALLBACK=https://<your-ngrok-id>.ngrok.io  \
  bin/rails "twilio:call[+12344442222]"
```

Then, you can interact with an AI agent and give it some instructions:

- What's on my list for today?
- What's my plans for the weekend?
- Create a "buy milk" task for tomorrow.

### Initiating calls from your phone

To call your Twilio number and interact with the AI agent, you MUST configure a status callback webhook for the phone number
pointing to another Ngrok tunnel (for port 3000):

```sh
ngrok http 3000
```

Then, use the `TWILIO_STATUS_CALLBACK` env var to provide the webhook URL to the application:

```sh
TWILIO_STATUS_CALLBACK=https://<your-rails-ngrok-id>.ngrok.io \
 TWILIO_STREAM_CALLBACK=https://<your-cable-ngrok-id>.ngrok.io \
 bin/dev
```

## Calls monitoring

Go to the [localhost:3000/phone_calls](http://localhost:3000/phone_calls) to see some live logs of your calls.
