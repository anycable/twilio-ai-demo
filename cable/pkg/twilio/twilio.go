package twilio

import "encoding/json"

// See https://www.twilio.com/docs/voice/media-streams/websocket-messages
const (
	ConnectedEvent = "connected"
	StartEvent     = "start"
	MediaEvent     = "media"
	MarkEvent      = "mark"
	StopEvent      = "stop"
	ClearEvent     = "clear"
	DTMFEvent      = "dtmf"
)

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

type ConnectedMessage struct {
	Event    string `json:"event"`
	Protocol string `json:"protocol"`
	Version  string `json:"version"`
}

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

type MarkPayload struct {
	Name string `json:"name"`
}

type MarkMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid,omitempty"`
	Seq       int64  `json:"sequenceNumber,omitempty"`

	Mark MarkPayload `json:"mark"`
}

type ClearMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid"`
}

type DTMFPayload struct {
	Track string `json:"track"`
	Digit string `json:"digit"`
}

type DTMFMessage struct {
	Event     string `json:"event"`
	StreamSID string `json:"streamSid,omitempty"`
	Seq       int64  `json:"sequenceNumber,omitempty"`

	DTMF DTMFPayload `json:"dtmf"`
}
