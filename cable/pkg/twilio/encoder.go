package twilio

import (
	"encoding/json"
	"fmt"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/encoders"
	"github.com/anycable/anycable-go/ws"
)

// Encoder converts messages from/to Twilio format to AnyCable format
type Encoder struct {
}

// We only need to parse event type and streamSid in the encoder.
// We pass the whole message in the Data field to be deserialized later in the executor.
type DecodeMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid"`

	// All possible payloads
	Start StartPayload `json:"start,omitempty"`
	Media MediaPayload `json:"media,omitempty"`
	Stop  StopPayload  `json:"stop,omitempty"`
	Mark  MarkPayload  `json:"mark,omitempty"`
}

var _ encoders.Encoder = (*Encoder)(nil)

const twilioEncoderID = "twilio"

func (Encoder) ID() string {
	return twilioEncoderID
}

func (Encoder) Encode(msg encoders.EncodedMessage) (*ws.SentFrame, error) {
	mtype := msg.GetType()

	// Ignore pings, disconnects, confirmations, welcome messages
	if mtype == common.PingType || mtype == common.DisconnectType {
		return nil, nil
	}

	r, ok := msg.(*common.Reply)

	if !ok {
		return nil, fmt.Errorf("Unknown message type: %v", msg)
	}

	if r.Type == common.ConfirmedType || r.Type == common.WelcomeType {
		return nil, nil
	}

	var response interface{}

	if r.Type == MediaEvent {
		response = MediaMessage{
			Event:     MediaEvent,
			StreamSID: r.Identifier,
			Media:     r.Message.(MediaPayload),
		}
	}

	if r.Type == MarkEvent {
		response = MarkMessage{
			Event:     MarkEvent,
			StreamSID: r.Identifier,
			Mark:      r.Message.(MarkPayload),
		}
	}

	if r.Type == ClearEvent {
		response = ClearMessage{
			Event:     ClearEvent,
			StreamSID: r.Identifier,
		}
	}

	if response == nil {
		response = r.Message
	}

	b, err := json.Marshal(response)

	if err != nil {
		return nil, err
	}

	return &ws.SentFrame{FrameType: ws.TextFrame, Payload: b}, nil
}

func (enc Encoder) EncodeTransmission(raw string) (*ws.SentFrame, error) {
	msg := common.Reply{}

	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return nil, err
	}

	return enc.Encode(&msg)
}

func (Encoder) Decode(raw []byte) (*common.Message, error) {
	twMsg := &DecodeMessage{}

	if err := json.Unmarshal(raw, &twMsg); err != nil {
		return nil, err
	}

	var data interface{}

	switch twMsg.Event {
	case StartEvent:
		data = twMsg.Start
	case MediaEvent:
		data = twMsg.Media
	case MarkEvent:
		data = twMsg.Mark
	case StopEvent:
		data = twMsg.Stop
	}

	msg := common.Message{Command: twMsg.Event, Identifier: twMsg.StreamSID, Data: data}

	return &msg, nil
}
