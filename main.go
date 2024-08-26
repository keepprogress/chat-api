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
	"os"
)

var geminiKey = os.Getenv("GENAI_API_KEY")

var ChatTemperature = 0.5

// Upgrade to WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {

	// upgrade this connection to a WebSocket
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// helpful Log statement to show connections
	log.Println("Client Connected")

	reader(ws)
}

func reader(conn *websocket.Conn) {
	ctx := context.Background()
	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// For text-only input, use the gemini-pro model
	model := client.GenerativeModel("gemini-1.5-flash-latest")
	value := float32(ChatTemperature) // Set the temperature to 0.5
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
	http.HandleFunc("/chat", wsEndpoint)
	fmt.Println("started ...") // Print the chunk message
	log.Fatal(http.ListenAndServe(":8080", nil))
}
