package twilio

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/node"
	"github.com/anycable/anycable-go/utils"
	"github.com/anycable/anycable-go/ws"

	"github.com/palkan/twilio-ai-cable/pkg/agent"
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

		s.WriteInternalState("callSid", start.CallSID)
		s.WriteInternalState("streamSid", start.StreamSID)

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

		identifier := channelId(start.CallSID, start.StreamSID)

		// We need to perform an additional RPC call to initialize the channel subscription
		_, err = ex.node.Subscribe(s, &common.Message{Identifier: identifier, Command: "subscribe"})

		if err != nil {
			return err
		}

		err = ex.initAgent(s, identifier)

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

		ai := ex.getAI(s)

		if ai == nil {
			return nil
		}

		audioBytes, err := base64.StdEncoding.DecodeString(twilioMsg.Payload)

		if err != nil {
			return err
		}

		err = ai.EnqueueAudio(audioBytes)

		return err
	}

	if msg.Command == MarkEvent {
		s.Log.Debug("mark received", "msg", msg.Data)
		// TODO: Here we can track which media messages has been processed,
		// so we can implement some clearing logic.
		// See https://www.twilio.com/docs/voice/media-streams/websocket-messages#send-a-clear-message
		return nil
	}

	if msg.Command == DTMFEvent {
		// DTMF is sent over RPC

		dtfm := msg.Data.(DTMFPayload)
		_, err := ex.performRPC(s, "handle_dtmf", map[string]string{"digit": dtfm.Digit})

		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Unknown command: %s", msg.Command)
}

func (ex *Executor) Disconnect(s *node.Session) error {
	ai := ex.getAI(s)

	if ai != nil {
		ai.Close()
	}

	return ex.node.Disconnect(s)
}

func (ex *Executor) initAgent(s *node.Session, identifier string) error {
	openAIKey := strChannelState(s, identifier, "openai_key")

	if openAIKey == "" {
		return errors.New("OpenAI key is not set")
	}

	conf := agent.NewConfig(openAIKey)

	openAIModel := strChannelState(s, identifier, "openai_model")
	if openAIModel != "" {
		conf.Model = openAIModel
	}

	openAIVoice := strChannelState(s, identifier, "ai_voice")
	if openAIVoice != "" {
		conf.Voice = openAIVoice
	}

	agent := agent.NewAgent(conf, s.Log)

	err := agent.KickOff(context.Background())
	if err != nil {
		return err
	}

	s.WriteInternalState("agent", agent)

	return nil
}

func (ex *Executor) getAI(s *node.Session) *agent.Agent {
	var ai *agent.Agent

	if rawAgent, ok := s.ReadInternalState("agent"); ok {
		ai = rawAgent.(*agent.Agent)
	}

	return ai
}

func (ex *Executor) performRPC(s *node.Session, action string, data map[string]string) (*common.CommandResult, error) {
	var callSID string
	var streamSID string

	if val, found := s.ReadInternalState("callSid"); found {
		callSID = val.(string)
	} else {
		return nil, errors.New("Call SID not found")
	}

	if val, found := s.ReadInternalState("streamSid"); found {
		streamSID = val.(string)
	} else {
		return nil, errors.New("Stream SID not found")
	}

	data["action"] = action

	payload := utils.ToJSON(data)

	return ex.node.Perform(s, &common.Message{
		Identifier: channelId(callSID, streamSID),
		Command:    "message",
		Data:       string(payload),
	})
}

func strChannelState(s *node.Session, identifier string, key string) string {
	jsonVal := s.GetEnv().GetChannelStateField(identifier, key)

	if jsonVal == "" {
		return ""
	}

	var val string
	err := json.Unmarshal([]byte(jsonVal), &val)

	if err != nil {
		return ""
	}

	return val
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
