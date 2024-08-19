package main

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
)

// 设置 Gemini API 密钥
var geminiKey = "AIzaSyDXXHJH3QtY_Ap7rTYGVtT01EaU_W92vGw"

// 设置聊天温度
var ChatTemperture = 0.5

// GeminiChat: Input a message and get the response string.
func GeminiChat(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200") // Replace with the actual origin of your client application
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Generate response using GenerateContentStream
	iter := model.GenerateContentStream(ctx, genai.Text(string(body)))

	// Send response as plain text
	// Instead of sending the entire response immediately,
	// use a polling mechanism to send chunks of text.
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			return // Return normally when iteration is done
		}
		if err != nil {
			http.Error(w, "Error getting response from model", http.StatusInternalServerError)
			return // Return on error
		}
		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			http.Error(w, "Empty response from model", http.StatusInternalServerError)
			return // Return on empty response
		}

		// Print each chunk as it arrives
		for _, c := range resp.Candidates {
			for _, p := range c.Content.Parts {
				fmt.Fprintf(w, "%s ", p)
				fmt.Println("chunk message:", p) // Print the chunk message
			}
		}
		fmt.Fprint(w, "\n")
	}
}

// Print response
func printResponse(resp *genai.GenerateContentResponse) string {
	var ret string
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			ret = ret + fmt.Sprintf("%v", part)
			log.Println(part)
		}
	}
	return ret
}

func main() {
	http.HandleFunc("/chat", GeminiChat)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
