# AnyCable Go scaffold

This is a template repository containing a scaffolding code to build real-time Go applications on top of [AnyCable][anycable-go].

Why building a WebSocket application with AnyCable (and not other Go libraries)?

- Connect your application to Ruby/Rails/JS/whatever apps with ease by using AnyCable RPC protocol.
- Many features out-of-the-box including different pub/sub adapters (including [embedded NATS][enats]), built-in instrumentation.
- Bulletproof code, which has been used production for years.

Read more about [AnyCable][anycable-docs].

## Examples

- [AnyCable + Twilio Streams](https://github.com/anycable/anycable-twilio-hanami-demo)

## Installation

Clone this repository:

```sh
git clone --depth 1 https://github.com/anycable/anycable-go-scaffold my-cable-project
```

Run the following command to set up the project:

```sh
make init
```

It would rename the project from `anycable/mycable` to whatever name you want (`<org>/<project>`) updating the `go.mod`, `*.go` files, etc.

## Development

**NOTE:** Make sure Go 1.23+ installed.

The following commands are available:

```shell
# Build the Go binary (will be available in dist/twilio-anycable)
make

# Run Golang tests
make test
```

We use [golangci-lint](https://golangci-lint.run) to lint Go source code:

```sh
make lint
```

### Git hooks

To automatically lint and test code before commits/pushes it is recommended to install [Lefthook][lefthook]:

```sh
brew install lefthook

lefthook install
```

[anycable-go]: https://github.com/anycable/anycable-go
[anycable-docs]: https://docs.anycable.io/anycable-go/getting_started
[lefthook]: https://github.com/evilmartians/lefthook
[enats]: https://docs.anycable.io/anycable-go/embedded_nats
