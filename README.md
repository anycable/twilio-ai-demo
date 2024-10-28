# AnyCable Twilio Streams x AI demo

This application demonstrates how to use [AnyCable](https://anycable.io) to interact with voice calls via Twilio Streams
and OpenAI Realtime API.

## Requirements

- Ruby 3.3 for the application server
- Go 1.23 to run AnyCable
- Twilio and OpenAI accounts

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
to be accessible from the Internet. You can use [ngrok](https://ngrok.com) for that:

```sh
ngrok http 8080
```

Use the generated URL as a `TWILIO_WEBHOOK_URL` environment variable when running the app:

```sh
TWILIO_WEBHOOK_URL=https://<your-ngrok-id>.ngrok.io bin/dev
```

Now, you can initiate a call to the provided phone number from the web app ("Make a call"),
and interact with the app using your voice. Example commands:

- What's on my list for today?
- What's my plans for the weekend?
- Create a "buy milk" task for tomorrow.
