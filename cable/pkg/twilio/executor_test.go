package twilio

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"testing"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/metrics"
	"github.com/anycable/anycable-go/mocks"
	"github.com/anycable/anycable-go/node"
	"github.com/anycable/anycable-go/node_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleCommandConnected(t *testing.T) {
	app := &node_mocks.AppNode{}
	n := NewMockNode()
	c := NewConfig()
	executor := NewExecutor(app, c)

	t.Run("when not connected", func(t *testing.T) {
		conn := mocks.NewMockConnection()
		session := buildSession(conn, n, executor, false)

		err := executor.HandleCommand(session, &common.Message{Command: ConnectedEvent})

		require.NoError(t, err)
		assert.True(t, session.Connected)
	})

	t.Run("when already connected", func(t *testing.T) {
		conn := mocks.NewMockConnection()
		session := buildSession(conn, n, executor, true)

		err := executor.HandleCommand(session, &common.Message{Command: ConnectedEvent})

		require.Error(t, err)
		assert.Equal(t, "Already connected", err.Error())
	})
}

func TestHandleCommandStart(t *testing.T) {
	app := &node_mocks.AppNode{}
	n := NewMockNode()
	c := NewConfig()
	executor := NewExecutor(app, c)

	t.Run("when not connected", func(t *testing.T) {
		conn := mocks.NewMockConnection()
		session := buildSession(conn, n, executor, false)

		err := executor.HandleCommand(session, &common.Message{Command: StartEvent})

		require.Error(t, err)
		assert.Equal(t, "Must be connected before receiving commands", err.Error())
	})

	t.Run("calls Authenticate with a header and subscribes", func(t *testing.T) {
		conn := mocks.NewMockConnection()
		session := buildSession(conn, n, executor, true)

		start := StartPayload{AccountSID: "ac42", CallSID: "ca123", StreamSID: "sm123"}

		// We should call authenticate
		app.On("Authenticated", session, mock.Anything).Run(func(args mock.Arguments) {
			identifiersJSON := args.Get(1).(string)

			var identifiers map[string]string
			err := json.Unmarshal([]byte(identifiersJSON), &identifiers)

			require.NoError(t, err)

			assert.Equal(t, "ca123", identifiers["call_sid"])
			assert.Equal(t, "sm123", identifiers["stream_sid"])
		})

		// And also subscribe to a channel (in authenticate passes)
		app.
			On("Subscribe", session, &common.Message{Identifier: channelId(session), Command: "subscribe"}).
			Return(nil, nil)

		// And perform AI configure step
		app.
			On("Perform", session, &common.Message{Identifier: channelId(session), Command: "message", Data: `{"action":"configure_openai"}`}).
			Return(nil, nil)

		err := executor.HandleCommand(session, &common.Message{
			Identifier: "ca123",
			Command:    StartEvent,
			Data:       start,
		})

		require.NoError(t, err)
	})
}

func TestHandleCommandMedia(t *testing.T) {
	n := NewMockNode()
	c := NewConfig()
	executor := NewExecutor(n, c)

	t.Run("when not connected", func(t *testing.T) {
		conn := mocks.NewMockConnection()
		session := buildSession(conn, n, executor, false)

		err := executor.HandleCommand(session, &common.Message{Command: MediaEvent})

		require.Error(t, err)
		assert.Equal(t, "Must be connected before receiving commands", err.Error())
	})
}

func TestHandleCommandMark(t *testing.T) {
	n := NewMockNode()
	c := NewConfig()
	executor := NewExecutor(n, c)

	t.Run("when not connected", func(t *testing.T) {
		conn := mocks.NewMockConnection()
		session := buildSession(conn, n, executor, false)

		err := executor.HandleCommand(session, &common.Message{Command: MarkEvent})

		require.Error(t, err)
		assert.Equal(t, "Must be connected before receiving commands", err.Error())
	})

	t.Run("returns no error", func(t *testing.T) {
		conn := mocks.NewMockConnection()
		session := buildSession(conn, n, executor, true)

		err := executor.HandleCommand(session, &common.Message{Command: MarkEvent})

		require.NoError(t, err)
	})
}

var (
	sessionCounter = 1
)

func buildSession(conn node.Connection, n *node.Node, executor node.Executor, connected bool) *node.Session {
	sessionCounter++
	s := node.NewSession(n, conn, "ws://anycable.io/twilio", nil, strconv.Itoa(sessionCounter), node.WithEncoder(Encoder{}), node.WithExecutor(executor))
	s.Connected = connected
	s.Log = slog.With("context", "test")
	return s
}

// NewMockNode build new node with mock controller
func NewMockNode() *node.Node {
	controller := mocks.NewMockController()
	config := node.NewConfig()
	config.HubGopoolSize = 2
	node := node.NewNode(&config, node.WithController(&controller), node.WithInstrumenter(metrics.NewMetrics(nil, 10, slog.Default())))
	return node
}
