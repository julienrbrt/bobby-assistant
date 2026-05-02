package llm

type Schema struct {
	Type        string            `json:"type,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Description string            `json:"description,omitempty"`
	Enum        []string          `json:"enum,omitempty"`
	Format      string            `json:"format,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
}

type FunctionDecl struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Parameters  *Schema `json:"parameters,omitempty"`
}

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
