package main

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"github.com/gorilla/websocket"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"log"
	"net/http"
)

// 设置 Gemini API 密钥
var geminiKey = "AIzaSyDXXHJH3QtY_Ap7rTYGVtT01EaU_W92vGw"

// 设置聊天温度
var ChatTemperture = 0.5

// Upgrade to WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

// GeminiChat: Input a message and get the response string.
func GeminiChat(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	ctx := context.Background()
	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// For text-only input, use the gemini-pro model
	model := client.GenerativeModel("gemini-1.5-flash-latest")
	value := float32(ChatTemperture) // Set the temperature to 0.5
	model.Temperature = &value

	for {
		// Read message from WebSocket
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// Handle received message
		if messageType == websocket.TextMessage {
			question := string(message)

			// Generate response using GenerateContentStream
			iter := model.GenerateContentStream(ctx, genai.Text(question))

			// Send response as a stream of events through WebSocket
			for {
				resp, err := iter.Next()
				if err == iterator.Done {
					break // Return normally when iteration is done
				}
				if err != nil {
					conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error getting response from model: %s", err.Error())))
					break // Return on error
				}
				if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
					conn.WriteMessage(websocket.TextMessage, []byte("Empty response from model"))
					break // Return on empty response
				}

				// Send each chunk as a separate message
				for _, c := range resp.Candidates {
					for _, p := range c.Content.Parts {
						// Convert Part to []byte
						partStr := fmt.Sprintf("%v", p)
						partBytes := []byte(partStr)
						conn.WriteMessage(websocket.TextMessage, partBytes)
						fmt.Println("chunk message:", p) // Print the chunk message
					}
				}
			}
		}
	}
}

func main() {
	http.HandleFunc("/chat", GeminiChat)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
