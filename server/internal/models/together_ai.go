package models

// TogetherAIRequest represents the request structure for Together AI API
type TogetherAIRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// TogetherAIResponse represents the response structure from Together AI
type TogetherAIResponse struct {
	Output struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	} `json:"output"`
}
