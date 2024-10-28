package cli

import (
	"fmt"
	"log/slog"
	"net/http"

	acli "github.com/anycable/anycable-go/cli"
	aconfig "github.com/anycable/anycable-go/config"
	"github.com/anycable/anycable-go/metrics"
	"github.com/anycable/anycable-go/node"
	"github.com/anycable/anycable-go/server"
	"github.com/anycable/anycable-go/ws"
	"github.com/gorilla/websocket"

	"github.com/palkan/twilio-ai-cable/internal/fake_rpc"
	"github.com/palkan/twilio-ai-cable/pkg/config"
	"github.com/palkan/twilio-ai-cable/pkg/custom"
	"github.com/palkan/twilio-ai-cable/pkg/version"
)

func Run(conf *config.Config, anyconf *aconfig.Config) error {
	// Configure your logger here
	logger := slog.Default()

	anycable, err := initAnyCableRunner(conf, anyconf, logger)

	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Starting custom AnyCable v%s", version.Version()))

	return anycable.Run()
}

func initAnyCableRunner(appConf *config.Config, anyConf *aconfig.Config, l *slog.Logger) (*acli.Runner, error) {
	opts := []acli.Option{
		acli.WithName("AnyCable"),
		acli.WithLogger(l),
		acli.WithDefaultSubscriber(),
		acli.WithDefaultBroker(),
		acli.WithDefaultBroadcaster(),
		acli.WithWebSocketEndpoint("/ws", myWebsocketHandler(appConf)),
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

func myWebsocketHandler(config *config.Config) func(n *node.Node, c *aconfig.Config, lg *slog.Logger) (http.Handler, error) {
	return func(n *node.Node, c *aconfig.Config, lg *slog.Logger) (http.Handler, error) {
		extractor := server.DefaultHeadersExtractor{Headers: c.RPC.ProxyHeaders, Cookies: c.RPC.ProxyCookies}

		executor := custom.NewExecutor(n)

		lg.Info(fmt.Sprintf("Handle custom WebSocket connections at ws://%s:%d/ws", c.Server.Host, c.Server.Port))

		return ws.WebsocketHandler([]string{}, &extractor, &c.WS, lg, func(wsc *websocket.Conn, info *server.RequestInfo, callback func()) error {
			wrappedConn := ws.NewConnection(wsc)
			session := node.NewSession(
				n, wrappedConn, info.URL, info.Headers, info.UID,
				node.WithEncoder(custom.Encoder{}), node.WithExecutor(executor),
			)

			_, err := n.Authenticate(session)

			if err != nil {
				return err
			}

			return session.Serve(callback)
		}), nil
	}
}
