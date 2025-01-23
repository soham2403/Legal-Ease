package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/internal/models"

	"github.com/joho/godotenv"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Hello World"}
	json.NewEncoder(w).Encode(response)
}

func HandleChat(w http.ResponseWriter, r *http.Request) {

	var chatReq models.ChatRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Format chat history for the prompt
	chatHistoryStr := ""
	for _, msg := range chatReq.ChatHistory {
		chatHistoryStr += fmt.Sprintf("%s: %s\n", msg.Sender, msg.Message)
	}

	// Create prompt using the template from the Python code
	prompt := fmt.Sprintf(`<s>[INST]This is a chat template. As a legal chat bot specializing in Indian Penal Code queries, 
your primary objective is to provide accurate and concise information based on the user's questions. 
You will adhere strictly to the instructions provided, offering relevant context from the knowledge base 
while avoiding unnecessary details. Your responses will be brief and to the point.
CHAT HISTORY: %s
HUMAN: %s
ASSISTANT:
</s>[INST]`, chatHistoryStr, chatReq.UserInput)

	// Make request to Together AI API
	client := &http.Client{}
	togetherAIReq := models.TogetherAIRequest{
		Model:       "mistralai/Mistral-7B-Instruct-v0.2",
		Prompt:      prompt,
		MaxTokens:   1024,
		Temperature: 0.5,
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(togetherAIReq)
	if err != nil {
		http.Error(w, "Failed to create API request", http.StatusInternalServerError)
		return
	}

	// Create request to Together AI API
	req, err := http.NewRequest("POST", "https://api.together.xyz/inference", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, "Failed to create API request", http.StatusInternalServerError)
		return
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	// Access environment variables using os.Getenv
	TOGETHER_AI_API := os.Getenv("TOGETHER_AI_API")
	if TOGETHER_AI_API == "" {
		log.Fatal("API key is empty...")
		return
	}

	req.Header.Set("Authorization", "Bearer "+TOGETHER_AI_API)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get response from AI service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse response from Together AI
	var aiResp models.TogetherAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		http.Error(w, "Failed to parse API response", http.StatusInternalServerError)
		return
	}

	// Extract the text from the response
	var responseText string
	if len(aiResp.Output.Choices) > 0 {
		responseText = aiResp.Output.Choices[0].Text
	} else {
		responseText = "No response generated"
	}

	// Send response back to client
	chatResp := models.ChatResponse{
		AssistantResponse: responseText,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}
