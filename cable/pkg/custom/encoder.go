package custom

import (
	"encoding/json"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/encoders"
	"github.com/anycable/anycable-go/utils"
	"github.com/anycable/anycable-go/ws"
)

// Encoder converts messages from/to custom format to AnyCable format
type Encoder struct {
}

var _ encoders.Encoder = (*Encoder)(nil)

const customEncoderID = "custom"

func (Encoder) ID() string {
	return customEncoderID
}

// Encode converts an outgoing message from server to client from AnyCable format to
// custom format.
// If nil is returned, the message is not sent to the client at all.
func (Encoder) Encode(msg encoders.EncodedMessage) (*ws.SentFrame, error) {
	b := utils.ToJSON(msg)

	return &ws.SentFrame{FrameType: ws.TextFrame, Payload: b}, nil
}

// EncodeTransmission converts an outgoing message from RPC to client from AnyCable format as a JSON string to
// custom format.
// If nil is returned, the message is not sent to the client at all.
func (enc Encoder) EncodeTransmission(raw string) (*ws.SentFrame, error) {
	msg := common.Reply{}

	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return nil, err
	}

	return enc.Encode(&msg)
}

// Decode converts an incoming message from client to AnyCable format.
func (Encoder) Decode(raw []byte) (*common.Message, error) {
	msg := &common.Message{}

	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}

	return msg, nil
}
