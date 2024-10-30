package agent

// Struct representing various OpenAI events
// See https://platform.openai.com/docs/api-reference/realtime-server-events

type ItemEvent struct {
	EventId      string `json:"event_id"`
	Type         string `json:"type"`
	ItemId       string `json:"item_id"`
	ContentIndex int    `json:"content_index"`
}

func (ev *ItemEvent) GetItemId() string {
	return ev.ItemId
}

func (ev *ItemEvent) GetType() string {
	return ev.Type
}

type OutputItemEvent struct {
	ResponseId  string `json:"response_id"`
	OutputIndex int    `json:"output_index"`
	ItemEvent
}

type TranscriptEvent interface {
	GetRole() string
	GetItemId() string
	GetTranscript() string
}

type InputAudioTranscriptionCompletedEvent struct {
	Transcript string `json:"transcript"`
	ItemEvent
}

func (ev *InputAudioTranscriptionCompletedEvent) GetRole() string {
	return "user"
}

func (ev *InputAudioTranscriptionCompletedEvent) GetTranscript() string {
	return ev.Transcript
}

var _ TranscriptEvent = (*InputAudioTranscriptionCompletedEvent)(nil)

type AudioTranscriptDeltaEvent struct {
	Delta string `json:"delta"`
	OutputItemEvent
}

func (ev *AudioTranscriptDeltaEvent) GetRole() string {
	return "assistant"
}

func (ev *AudioTranscriptDeltaEvent) GetTranscript() string {
	return ev.Delta
}

var _ TranscriptEvent = (*AudioTranscriptDeltaEvent)(nil)

type AudioTranscriptDoneEvent struct {
	Transcript string `json:"transcript"`
	OutputItemEvent
}

func (ev *AudioTranscriptDoneEvent) GetRole() string {
	return "assistant"
}

func (ev *AudioTranscriptDoneEvent) GetTranscript() string {
	return ev.Transcript
}

var _ TranscriptEvent = (*AudioTranscriptDoneEvent)(nil)

type AudioDeltaEvent struct {
	Delta string `json:"delta"`
	OutputItemEvent
}
