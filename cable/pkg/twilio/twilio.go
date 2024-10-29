package twilio

import "encoding/json"

const (
	ConnectedEvent = "connected"
	StartEvent     = "start"
	MediaEvent     = "media"
	MarkEvent      = "mark"
	StopEvent      = "stop"
	ClearEvent     = "clear"
)

// https://www.twilio.com/docs/voice/twiml/stream#message-start
type StartPayload struct {
	AccountSID string `json:"accountSid"`
	StreamSID  string `json:"streamSid"`
	CallSID    string `json:"callSid"`
}

func (p *StartPayload) ToJSON() ([]byte, error) {
	b, err := json.Marshal(&p)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type StartMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid"`
	Seq       int64  `json:"sequenceNumber"`

	Start StartPayload `json:"start"`
}

// https://www.twilio.com/docs/voice/twiml/stream#message-connected
type ConnectedMessage struct {
	Event    string `json:"event"`
	Protocol string `json:"protocol"`
	Version  string `json:"version"`
}

// https://www.twilio.com/docs/voice/twiml/stream#message-media
// https://www.twilio.com/docs/voice/twiml/stream#message-media-to-twilio
type MediaPayload struct {
	Payload string `json:"payload"`
	Track   string `json:"track"`
}

type MediaMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid,omitempty"`
	Seq       int64  `json:"sequenceNumber,omitempty"`

	Media MediaPayload `json:"media"`
}

// https://www.twilio.com/docs/voice/twiml/stream#message-stop
type StopPayload struct {
	AccountSID string `json:"accountSid"`
	StreamSID  string `json:"streamSid"`
}

type StopMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid"`
	Seq       int64  `json:"sequenceNumber"`

	Stop StopPayload `json:"stop"`
}

// https://www.twilio.com/docs/voice/twiml/stream#message-mark
// https://www.twilio.com/docs/voice/twiml/stream#message-mark-to-twilio
type MarkPayload struct {
	Name string `json:"name"`
}

type MarkMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid,omitempty"`
	Seq       int64  `json:"sequenceNumber,omitempty"`

	Mark MarkPayload `json:"mark"`
}

// https://www.twilio.com/docs/voice/twiml/stream#message-clear-to-twilio
type ClearMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid"`
}
