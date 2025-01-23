package models

// ChatRequest represents the incoming chat request structure
type ChatRequest struct {
	UserInput   string        `json:"user_input"`
	ChatHistory []ChatMessage `json:"chat_history"`
}

// ChatMessage represents a single message in the chat history
type ChatMessage struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

// ChatResponse represents the response structure
type ChatResponse struct {
	AssistantResponse string `json:"assistant_response"`
}
