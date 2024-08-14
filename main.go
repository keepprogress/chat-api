package main

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

	// Generate response
	resp, err := model.GenerateContent(ctx, genai.Text(string(body))) // 直接使用 body 内容
	if err != nil {
		http.Error(w, "Error generating response: %v", http.StatusInternalServerError)
		return
	}

	// Send response as plain text
	// Instead of sending the entire response immediately,
	// use a polling mechanism to send chunks of text.
	for _, cand := range resp.Candidates {
		log.Println("here is at for loop level 1 ...")
		for _, part := range cand.Content.Parts {
			fmt.Fprintf(w, "%v", part)
			log.Println(part)
			w.(http.Flusher).Flush()           // Flush the response to the client
			time.Sleep(100 * time.Millisecond) // Adjust the polling interval
		}
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
