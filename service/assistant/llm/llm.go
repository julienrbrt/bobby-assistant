package llm

type FunctionCall struct {
	ID   string         `json:"id,omitempty"`
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type FunctionResponse struct {
	CallID   string         `json:"callId,omitempty"`
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type ChatMessage struct {
	Role             string
	Content          string
	FunctionCall     *FunctionCall
	FunctionResponse *FunctionResponse
}
