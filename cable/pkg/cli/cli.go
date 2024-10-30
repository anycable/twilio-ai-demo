package cli

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	acli "github.com/anycable/anycable-go/cli"
	aconfig "github.com/anycable/anycable-go/config"
	"github.com/anycable/anycable-go/logger"
	"github.com/anycable/anycable-go/metrics"
	"github.com/anycable/anycable-go/node"
	"github.com/anycable/anycable-go/server"
	"github.com/anycable/anycable-go/ws"
	"github.com/gorilla/websocket"

	"github.com/palkan/twilio-ai-cable/internal/fake_rpc"
	"github.com/palkan/twilio-ai-cable/pkg/config"
	"github.com/palkan/twilio-ai-cable/pkg/twilio"
	"github.com/palkan/twilio-ai-cable/pkg/version"
)

func Run(conf *config.Config, anyconf *aconfig.Config) error {
	// Configure your logger here
	logHandler, err := logger.InitLogger(anyconf.Log.LogFormat, anyconf.Log.LogLevel)
	log := slog.New(logHandler)

	if err != nil {
		return err
	}

	anycable, err := initAnyCableRunner(conf, anyconf, log)

	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Starting Twilio AnyCable v%s", version.Version()))

	return anycable.Run()
}

func initAnyCableRunner(appConf *config.Config, anyConf *aconfig.Config, l *slog.Logger) (*acli.Runner, error) {
	opts := []acli.Option{
		acli.WithName("AnyCable"),
		acli.WithDefaultSubscriber(),
		acli.WithDefaultBroker(),
		acli.WithDefaultBroadcaster(),
		acli.WithWebSocketEndpoint("/twilio", twilioWebsocketHandler(appConf)),
	}

	if appConf.FakeRPC {
		opts = append(opts, acli.WithController(func(m *metrics.Metrics, c *aconfig.Config, lg *slog.Logger) (node.Controller, error) {
			return fake_rpc.NewController(lg), nil
		}))
	} else {
		opts = append(opts, acli.WithDefaultRPCController())
	}

	return acli.NewRunner(anyConf, opts)
}

func twilioWebsocketHandler(config *config.Config) func(n *node.Node, c *aconfig.Config, lg *slog.Logger) (http.Handler, error) {
	return func(n *node.Node, c *aconfig.Config, lg *slog.Logger) (http.Handler, error) {
		extractor := server.DefaultHeadersExtractor{Headers: c.RPC.ProxyHeaders, Cookies: c.RPC.ProxyCookies}

		executor := twilio.NewExecutor(n, config.Twilio)

		lg.Info(fmt.Sprintf("Handle Twilio Media Streams connections at ws://%s:%d/twilio", c.Server.Host, c.Server.Port))

		return ws.WebsocketHandler([]string{}, &extractor, &c.WS, lg, func(wsc *websocket.Conn, info *server.RequestInfo, callback func()) error {
			wrappedConn := ws.NewConnection(wsc)
			session := node.NewSession(
				n, wrappedConn, info.URL, info.Headers, info.UID,
				node.WithEncoder(twilio.Encoder{}), node.WithExecutor(executor),
				node.WithHandshakeMessageDeadline(time.Now().Add(5*time.Second)),
			)

			return session.Serve(callback)
		}), nil
	}
}
