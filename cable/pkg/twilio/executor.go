package twilio

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/node"
	"github.com/anycable/anycable-go/ws"

	"github.com/palkan/twilio-ai-cable/pkg/config"
)

const channelName = "TwilioStreamChannel"

// Handling Twilio events and transforming them into Action Cable commands
type Executor struct {
	node node.AppNode
	conf *config.Config
}

var _ node.Executor = (*Executor)(nil)

func NewExecutor(node node.AppNode, c *config.Config) *Executor {
	return &Executor{node: node, conf: c}
}

func (ex *Executor) HandleCommand(s *node.Session, msg *common.Message) error {
	if msg.Command == ConnectedEvent {
		if s.Connected {
			return errors.New("Already connected")
		}

		s.Connected = true
		return nil
	}

	if msg.Command == StopEvent {
		s.Log.Debug("stop received, disconnecting")
		s.Disconnect("stream stopped", ws.CloseNormalClosure)
		return nil
	}

	if !s.Connected {
		return errors.New("Must be connected before receiving commands")
	}

	// That's the first message with some additional information.
	// Here we should perform authentication (#kick_off)
	if msg.Command == StartEvent {
		start, ok := msg.Data.(StartPayload)

		s.Log.Debug("incoming start message", "msg", start)

		if !ok {
			return fmt.Errorf("Malformed start message: %v", msg.Data)
		}

		s.InternalState = make(map[string]interface{})
		s.InternalState["callSid"] = start.CallSID
		s.InternalState["streamSid"] = start.StreamSID

		// We add account SID as a header to the sesssion.
		// So, we can access it via request.headers['x-twilio-account'] in Ruby.
		s.GetEnv().SetHeader("x-twilio-account", start.AccountSID)
		res, err := ex.node.Authenticate(s)

		if res != nil && res.Status == common.FAILURE {
			return nil
		}

		if err != nil {
			return err
		}

		// We need to perform an additional RPC call to initialize the channel subscription
		_, err = ex.node.Subscribe(s, &common.Message{Identifier: channelId(start.CallSID, start.StreamSID), Command: "subscribe"})

		if err != nil {
			return err
		}

		return nil
	}

	if msg.Command == MediaEvent {
		twilioMsg := msg.Data.(MediaPayload)

		// Ignore robot streams
		if twilioMsg.Track == "outbound" {
			return nil
		}

		// TODO: implement audio processing

		return nil
	}

	if msg.Command == MarkEvent {
		s.Log.Debug("mark received", "msg", msg.Data)
		return nil
	}

	return fmt.Errorf("Unknown command: %s", msg.Command)
}

func (ex *Executor) Disconnect(s *node.Session) error {
	// TODO: implement AI session cleanup
	return ex.node.Disconnect(s)
}

func channelId(callSid string, streamSid string) string {
	msg := struct {
		Channel   string `json:"channel"`
		CallSid   string `json:"call_sid"`
		StreamSid string `json:"stream_sid"`
	}{Channel: channelName, CallSid: callSid, StreamSid: streamSid}

	b, err := json.Marshal(msg)

	if err != nil {
		panic("Failed to build channel identifier 😲")
	}

	return string(b)
}
