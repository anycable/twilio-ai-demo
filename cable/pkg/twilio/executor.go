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
	"github.com/joomcode/errorx"

	"github.com/palkan/twilio-ai-cable/pkg/agent"
)

const channelName = "Twilio::MediaStreamChannel"
const responseState = "anycable_response"

type AppResponse struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// Handling Twilio events and transforming them into Action Cable commands
type Executor struct {
	node node.AppNode
	conf *Config
}

var _ node.Executor = (*Executor)(nil)

func NewExecutor(node node.AppNode, c *Config) *Executor {
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
	// Here we should perform authentication and start the AI session.
	if msg.Command == StartEvent {
		start, ok := msg.Data.(StartPayload)

		s.Log.Debug("incoming start message", "msg", start)

		if !ok {
			return fmt.Errorf("Malformed start message: %v", msg.Data)
		}

		// Check if account SID matches and reject the connection if not
		if ex.conf.AccountSID != "" && ex.conf.AccountSID != start.AccountSID {
			s.Log.Debug("unauthenticated stream", "account_sid", start.AccountSID[0:5]+"***")
			s.Disconnect("Auth Failed", ws.CloseNormalClosure)
			return nil
		}

		// Mark as authenticated and store the identifiers
		callSid := start.CallSID
		streamSid := start.StreamSID

		// Store identifiers in the session
		s.WriteInternalState("callSid", callSid)
		s.WriteInternalState("streamSid", streamSid)

		identifiers := string(utils.ToJSON(map[string]string{"call_sid": callSid, "stream_sid": streamSid}))

		ex.node.Authenticated(s, identifiers)

		// Now, subscribe to the channel to initialize the session
		identifier := channelId(s)
		_, err := ex.node.Subscribe(s, &common.Message{Identifier: identifier, Command: "subscribe"})

		if err != nil {
			return err
		}

		err = ex.initAgent(s)

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

		// TODO: handle response (e.g., send some command to the AI agent)

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

// Supported RPC response types
const configEvent = "openai.configuration"

type OpenAIConfigData struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model,omitempty"`
	Voice  string `json:"voice,omitempty"`
	Prompt string `json:"prompt,omitempty"`
	Tools  string `json:"tools,omitempty"`
}

func (ex *Executor) initAgent(s *node.Session) error {
	// Retrieve AI configuration from the main app
	res, err := ex.performRPC(s, "configure_openai", nil)

	if err != nil {
		return err
	}

	if res == nil {
		// No response from the main app, do not start the AI
		return nil
	}

	if res.Event != configEvent {
		return fmt.Errorf("unexpected response type from RPC: %s", res.Event)
	}

	var data OpenAIConfigData

	err = json.Unmarshal(res.Data, &data)
	if err != nil {
		return errorx.Decorate(err, "failed to parse OpenAI config from RPC")
	}

	conf := agent.NewConfig(data.APIKey)

	if data.Model != "" {
		conf.Model = data.Model
	}

	if data.Voice != "" {
		conf.Voice = data.Voice
	}

	if data.Prompt != "" {
		conf.Prompt = data.Prompt
	}

	if data.Tools != "" {
		conf.Tools = json.RawMessage(data.Tools)
	}

	agent := agent.NewAgent(conf, s.Log)

	agent.HandleTranscript(func(role string, text string, id string) {
		_, err := ex.performRPC(s, "handle_transcript", map[string]string{"role": role, "text": text, "id": id})

		if err != nil {
			s.Log.Error("failed to perform handle_transcript rpc", "error", err)
		}
	})

	agent.HandleAudio(func(encodedAudio string, id string) {
		var streamSid string
		if val, ok := s.ReadInternalState("streamSid"); ok {
			streamSid = val.(string)
		} else {
			return
		}

		s.Send(&common.Reply{Type: MediaEvent, Message: MediaPayload{Payload: encodedAudio}, Identifier: streamSid})
		s.Send(&common.Reply{Type: MarkEvent, Message: MarkPayload{Name: `ai-delta-` + id}, Identifier: streamSid})
	})

	agent.HandleFunctionCall(func(name string, args string, id string) {
		res, err := ex.performRPC(s, "handle_function_call", map[string]string{"name": name, "arguments": args})

		if err != nil {
			s.Log.Error("failed to perform handle_function_call rpc", "error", err)
		}

		if res != nil && res.Event == "openai.function_call_result" {
			agent.HandleFunctionCallResult(id, string(res.Data))
		}
	})

	err = agent.KickOff(context.Background())
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

func (ex *Executor) performRPC(s *node.Session, action string, data map[string]string) (*AppResponse, error) {
	if data == nil {
		data = make(map[string]string)
	}

	data["action"] = action

	payload := utils.ToJSON(data)

	identifier := channelId(s)

	res, err := ex.node.Perform(s, &common.Message{
		Identifier: identifier,
		Command:    "message",
		Data:       string(payload),
	})

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	// Fetch response from the RPC
	rawRes := res.IState[responseState]
	if rawRes == "" {
		return nil, nil
	}

	// Cleanup the channel state — we don't need to carry this state around
	st := (*s.GetEnv().ChannelStates)[identifier]
	delete(st, responseState)

	var rpcRes AppResponse
	err = json.Unmarshal([]byte(rawRes), &rpcRes)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to parse RPC response")
	}

	return &rpcRes, nil
}

func channelId(s *node.Session) string {
	msg := struct {
		Channel string `json:"channel"`
	}{Channel: channelName}

	return string(utils.ToJSON(msg))
}
