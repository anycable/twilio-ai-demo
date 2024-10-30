package agent

// Struct representing various OpenAI events
// See https://platform.openai.com/docs/api-reference/realtime-server-events

type InputAudioTranscriptionCompletedEvent struct {
	EventId      string `json:"event_id"`
	Type         string `json:"type"`
	ItemId       string `json:"item_id"`
	ContentIndex int    `json:"content_index"`
	Transcript   string `json:"transcript"`
}
