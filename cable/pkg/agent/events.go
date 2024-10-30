package agent

// Struct representing various OpenAI events
// See https://platform.openai.com/docs/api-reference/realtime-server-events

type Item struct {
	Id     string `json:"id,omitempty"`
	Object string `json:"object,omitempty"`
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
	Role   string `json:"role,omitempty"`
	// Function call fields
	CallID    string `json:"call_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Output    string `json:"output,omitempty"`
}

type Usage struct {
	TotalTokens       int `json:"total_tokens"`
	InputTokens       int `json:"input_tokens"`
	OutputTokens      int `json:"output_tokens"`
	InputTokenDetails struct {
		CachedTokens int `json:"cached_tokens"`
		TextTokens   int `json:"text_tokens"`
		AudioTokens  int `json:"audio_tokens"`
	} `json:"input_token_details"`
	OutputTokenDetails struct {
		TextTokens  int `json:"text_tokens"`
		AudioTokens int `json:"audio_tokens"`
	} `json:"output_token_details"`
}

type Response struct {
	ID            string `json:"id,omitempty"`
	Status        string `json:"status,omitempty"`
	StatusDetails struct {
		Type   string `json:"type,omitempty"`
		Reason string `json:"reason,omitempty"`
		Error  struct {
			Type    string `json:"type,omitempty"`
			Code    string `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error,omitempty"`
	} `json:"status_details,omitempty"`
	Usage *Usage `json:"usage,omitempty"`
}

type ResponseEvent struct {
	EventId  string `json:"event_id"`
	Type     string `json:"type"`
	Response *Response
}

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

type OutputItemDoneEvent struct {
	Item *Item `json:"item"`
	OutputItemEvent
}
