package twilio

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/anycable/anycable-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const identifier = "{\"channel\":\"twilio\",\"streamId\":\"abc2021\"}"

func TestEncoderEncode(t *testing.T) {
	coder := Encoder{}

	t.Run("Ping", func(t *testing.T) {
		msg := &common.PingMessage{Type: "ping", Message: time.Now().Unix()}
		actual, err := coder.Encode(msg)

		require.NoError(t, err)
		assert.Nil(t, actual)
	})

	t.Run("Disconnect", func(t *testing.T) {
		msg := &common.DisconnectMessage{Type: "disconnect", Reason: "unauthorized", Reconnect: false}

		actual, err := coder.Encode(msg)

		require.NoError(t, err)
		assert.Nil(t, actual)
	})

	t.Run("Media", func(t *testing.T) {
		msg := &common.Reply{
			Type:       MediaEvent,
			Identifier: "tw2021",
			Message:    MediaPayload{Payload: "<audio>"},
		}

		expected := toJSON(MediaMessage{
			Event:     MediaEvent,
			StreamSID: "tw2021",
			Media:     MediaPayload{Payload: "<audio>"},
		})

		actual, err := coder.Encode(msg)

		require.NoError(t, err)
		assert.Equal(t, string(expected), string(actual.Payload))
	})

	t.Run("Mark", func(t *testing.T) {
		msg := &common.Reply{
			Type:       MarkEvent,
			Identifier: "tw2021",
			Message:    MarkPayload{Name: "Matroskin"},
		}

		expected := toJSON(MarkMessage{
			Event:     MarkEvent,
			StreamSID: "tw2021",
			Mark:      MarkPayload{Name: "Matroskin"}},
		)

		actual, err := coder.Encode(msg)

		require.NoError(t, err)
		assert.Equal(t, string(expected), string(actual.Payload))
	})

	t.Run("Clear", func(t *testing.T) {
		msg := &common.Reply{
			Type:       ClearEvent,
			Identifier: "tw2021",
		}

		expected := toJSON(ClearMessage{
			Event:     ClearEvent,
			StreamSID: "tw2021",
		})

		actual, err := coder.Encode(msg)

		require.NoError(t, err)
		assert.Equal(t, string(expected), string(actual.Payload))
	})
}

func TestEncoderEncodeTransmission(t *testing.T) {
	coder := Encoder{}

	t.Run("welcome", func(t *testing.T) {
		msg := "{\"type\":\"welcome\"}"

		actual, err := coder.EncodeTransmission(msg)

		require.NoError(t, err)
		assert.Nil(t, actual)
	})

	t.Run("message", func(t *testing.T) {
		msg := toJSON(common.Reply{Identifier: identifier, Message: map[string]string{"payload": "audio"}})

		actual, err := coder.EncodeTransmission(string(msg))

		require.NoError(t, err)
		assert.Equal(t, "{\"payload\":\"audio\"}", string(actual.Payload))
	})
}

func TestEncoderDecode(t *testing.T) {
	coder := Encoder{}

	t.Run("connected", func(t *testing.T) {
		msg := toJSON(ConnectedMessage{Event: "connected", Protocol: "Test", Version: "1.0"})

		actual, err := coder.Decode(msg)

		require.NoError(t, err)
		assert.Equal(t, ConnectedEvent, actual.Command)
		assert.Nil(t, actual.Data)
	})

	t.Run("start", func(t *testing.T) {
		start := StartPayload{StreamSID: "tw2021", AccountSID: "ac2021"}
		msg := toJSON(StartMessage{
			Event:     "start",
			Seq:       1,
			StreamSID: "tw2021",
			Start:     start,
		})

		actual, err := coder.Decode(msg)

		require.NoError(t, err)
		assert.Equal(t, StartEvent, actual.Command)
		assert.Equal(t, "tw2021", actual.Identifier)
		assert.Equal(t, start, actual.Data)
	})

	t.Run("stop", func(t *testing.T) {
		stop := StopPayload{StreamSID: "tw2021", AccountSID: "ac2021"}
		msg := toJSON(StopMessage{
			Event:     "stop",
			Seq:       12,
			StreamSID: "tw2021",
			Stop:      stop,
		})

		actual, err := coder.Decode(msg)

		require.NoError(t, err)
		assert.Equal(t, StopEvent, actual.Command)
		assert.Equal(t, "tw2021", actual.Identifier)
		assert.Equal(t, stop, actual.Data)
	})

	t.Run("mark", func(t *testing.T) {
		mark := MarkPayload{Name: "Matroskin"}
		msg := toJSON(MarkMessage{
			Event:     "mark",
			Seq:       22,
			StreamSID: "tw2021",
			Mark:      mark,
		})

		actual, err := coder.Decode(msg)

		require.NoError(t, err)
		assert.Equal(t, MarkEvent, actual.Command)
		assert.Equal(t, "tw2021", actual.Identifier)
		assert.Equal(t, mark, actual.Data)
	})

	t.Run("media", func(t *testing.T) {
		media := MediaPayload{Payload: "<audio>"}
		msg := toJSON(MediaMessage{
			Event:     "media",
			Seq:       24,
			StreamSID: "tw2021",
			Media:     media,
		})

		actual, err := coder.Decode(msg)

		require.NoError(t, err)
		assert.Equal(t, MediaEvent, actual.Command)
		assert.Equal(t, "tw2021", actual.Identifier)
		assert.Equal(t, media, actual.Data)
	})
}

func toJSON(msg interface{}) []byte {
	b, err := json.Marshal(&msg)
	if err != nil {
		panic("Failed to build JSON 😲")
	}

	return b
}
