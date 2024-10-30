package agent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/anycable/anycable-go/logger"
	"github.com/anycable/anycable-go/utils"
	"github.com/gorilla/websocket"
	"github.com/joomcode/errorx"
)

type TranscriptHandler = func(role string, text string, id string)
type AudioHandler = func(data string, id string)
type FunctionHandler = func(name string, args string, id string)

// Agent represents a single Twilio Stream consumer connected
// to OpenAI realtime API
type Agent struct {
	conf *Config
	buf  *bytes.Buffer

	conn   *websocket.Conn
	sendCh chan []byte

	log *slog.Logger

	transcriptHandler TranscriptHandler
	audioHandler      AudioHandler
	functionHandler   FunctionHandler

	cancelFn context.CancelFunc
	connMu   sync.RWMutex
	mu       sync.Mutex
}

const (
	// 320 is the number of bytes in a single packet (20ms),
	// thus, flush every 300ms
	bytesPerFlush = 320 * 15
)

// NewAgent creates a new Agent instance with the given configuration.
func NewAgent(c *Config, l *slog.Logger) *Agent {
	return &Agent{
		conf:   c,
		buf:    bytes.NewBuffer(nil),
		sendCh: make(chan []byte, 128),
		log:    l.With("component", "openai"),
	}
}

func (a *Agent) HandleTranscript(handler TranscriptHandler) {
	a.transcriptHandler = handler
}

func (a *Agent) HandleAudio(handler AudioHandler) {
	a.audioHandler = handler
}

func (a *Agent) HandleFunctionCall(handler FunctionHandler) {
	a.functionHandler = handler
}

// KickOff starts the OpenAI WebSocket connection.
func (a *Agent) KickOff(ctx context.Context) error {
	url := a.conf.URL + "?model=" + a.conf.Model
	header := http.Header{
		"Authorization": []string{"Bearer " + a.conf.Key},
		"OpenAI-Beta":   []string{"realtime=v1"},
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, header)

	if err != nil {
		return errorx.Decorate(err, "could not dial OpenAI WebSocket")
	}

	ctx, cancel := context.WithCancel(ctx)

	a.connMu.Lock()
	a.cancelFn = cancel
	a.conn = conn
	a.connMu.Unlock()

	a.log.Debug("connected to OpenAI WebSocket")

	// Send session.update message to configure the session
	sessionConfig := map[string]interface{}{
		"type": "session.update",
		"session": map[string]interface{}{
			"input_audio_format":  "g711_ulaw",
			"output_audio_format": "g711_ulaw",
			"input_audio_transcription": map[string]string{
				"model": "whisper-1",
			},
		},
	}

	if a.conf.Prompt != "" {
		sessionConfig["session"].(map[string]interface{})["instructions"] = a.conf.Prompt
	}

	if a.conf.Tools != nil {
		sessionConfig["session"].(map[string]interface{})["tools"] = a.conf.Tools
	}

	configMessage := utils.ToJSON(sessionConfig)
	a.sendMsg(configMessage)

	go a.readMessages()
	go a.writeMessages(ctx)

	return nil
}

func (a *Agent) HandleFunctionCallResult(callID string, data string) {
	item := &Item{Type: "function_call_output", CallID: callID, Output: data}

	msg := struct {
		Type string `json:"type"`
		Item *Item  `json:"item"`
	}{"conversation.item.create", item}

	encoded := utils.ToJSON(msg)

	a.log.Warn("sending function call result", "data", string(encoded))

	a.sendMsg(encoded)
	// Send `response.create` message right away to trigger model inference
	a.sendMsg([]byte(`{"type":"response.create"}`))
}

func (a *Agent) EnqueueAudio(audio []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.buf.Write(audio)

	if a.buf.Len() > bytesPerFlush {
		if err := a.sendAudio(a.buf.Bytes()); err != nil {
			return errorx.Decorate(err, "could not send audio")
		}

		a.buf.Reset()
	}

	return nil
}

func (a *Agent) Close() {
	a.connMu.RLock()
	defer a.connMu.RUnlock()

	if a.cancelFn != nil {
		a.cancelFn()
	}

	if a.conn != nil {
		a.conn.Close()
	}
}

func (a *Agent) readMessages() {
	for {
		_, msg, err := a.conn.ReadMessage()
		if err != nil {
			a.log.Error("could not read message from OpenAI WebSocket", "err", err)
			return
		}

		a.log.Debug("received message from OpenAI WebSocket", "msg", logger.CompactValue(string(msg)))

		var typedMessage struct {
			Type string `json:"type"`
		}

		_ = json.Unmarshal(msg, &typedMessage)

		switch typedMessage.Type {
		case "session.created":
		case "session.updated":
		case "input_audio_buffer.speech_started":
		case "input_audio_buffer.speech_stopped":
		case "input_audio_buffer.committed":
		case "conversation.item.input_audio_transcription.completed":
			var event *InputAudioTranscriptionCompletedEvent
			_ = json.Unmarshal(msg, &event)

			a.handleTranscript(event)
		case "response.created":
		case "rate_limits.updated":
		case "response.output_item.added":
		case "conversation.item.created":
		case "response.content_part.added":
		case "response.audio.delta":
			var event *AudioDeltaEvent
			_ = json.Unmarshal(msg, &event)

			a.handleAudio(event)
		case "response.audio_transcript.delta":
			var event *AudioTranscriptDeltaEvent
			_ = json.Unmarshal(msg, &event)

			a.handleTranscript(event)
		case "response.audio.done":
		case "response.audio_transcript.done":
			var event *AudioTranscriptDoneEvent
			_ = json.Unmarshal(msg, &event)

			a.handleTranscript(event)
		case "response.function_call_arguments.delta":
		case "response.function_call_arguments.done":
		case "response.content_part.done":
		case "response.output_item.done":
			var event *OutputItemDoneEvent
			_ = json.Unmarshal(msg, &event)

			if event.Item.Type == "function_call" {
				a.handleFunctionCall(event.Item)
			}
		case "response.done":
			var event *ResponseEvent
			_ = json.Unmarshal(msg, &event)

			// Log errors
			if event.Response.Status == "failed" {
				a.log.Error("request failed", "error", event.Response.StatusDetails.Error)
			}
		case "error":
			a.log.Error("server error", "err", string(msg))
		default:
			a.log.Warn("unhandled message type", "type", typedMessage.Type)
		}
	}
}

func (a *Agent) writeMessages(ctx context.Context) {
	for {
		select {
		case msg := <-a.sendCh:
			if err := a.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				a.log.Error("could not write message to OpenAI WebSocket", "err", err)
				return
			}
		case <-ctx.Done():
			_ = a.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

func (a *Agent) sendMsg(msg []byte) {
	a.sendCh <- msg
}

func (a *Agent) sendAudio(audio []byte) error {
	encoded := base64.StdEncoding.EncodeToString(audio)

	msg := []byte(`{"type":"input_audio_buffer.append","audio": "` + encoded + `"}`)
	a.sendMsg(msg)

	return nil
}

func (a *Agent) handleTranscript(ev TranscriptEvent) {
	text := ev.GetTranscript()

	if text == "" {
		return
	}

	role := ev.GetRole()
	id := ev.GetItemId()

	a.log.Info("transcript", "text", text, "role", role, "id", id)

	if a.transcriptHandler != nil {
		a.transcriptHandler(role, text, id)
	}
}

func (a *Agent) handleAudio(ev *AudioDeltaEvent) {
	if a.audioHandler != nil {
		a.audioHandler(ev.Delta, ev.ItemId)
	}
}

func (a *Agent) handleFunctionCall(item *Item) {
	a.log.Debug("agent is trying to call a function", "name", item.Name, "args", item.Arguments, "id", item.CallID)

	if a.functionHandler != nil {
		a.functionHandler(item.Name, item.Arguments, item.CallID)
	}
}
