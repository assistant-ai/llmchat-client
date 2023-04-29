# LLMChat Client (Simplified GPT3/4 Go Client)

This Go library provides a simplified and easy-to-use interface for building chat-based GPT applications using GPT-3 or GPT-4. With just a few lines of code, you can interact with the GPT-3/4 AI to generate context-aware responses.

## Features

- Simplified interface for GPT-3/4 interactions
- Use of Go Modules for dependency management
- Default and customizable configurations for GPTClient
- Message history and context management using local SQLite database

## Main Methods

The main functions provided by the `GptClient` struct include:

- `func NewDefaultGptClient(openAiKey string) *GptClient` - Create a new GptClient with default configurations.
- `func NewDefaultGptClientFromFile(openAiKeyFilePath string) (*GptClient, error)` - Create a new GptClient with default configurations, reading the API key from a file.
- `func NewGptClient(openAiKey string, contextDepth int, model GPTModel, defaultContext string, maxTokens int) *GptClient` - Create a new, fully customized GptClient.
- `func NewGptClientFromFile(openAiKeyFlePath string, contextDepth int, model GPTModel, defaultContext string, maxTokens int) (*GptClient, error)` - Create a new, fully customized GptClient, reading the API key from a file.
- `func (g *GptClient) SenRandomContextMessage(message string) (string, error)` - Send a message using a random context.
- `func (g *GptClient) SendMessage(message string, inputContextId string) (string, error)` - Send a message using a specified or random context if none is provided.

## How to Use

1. Start by importing the `gpt` and `db` packages:

```go
import (
	"github.com/assistant-ai/llmchat-client/gpt"
	"github.com/assistant-ai/llmchat-client/db"
)
```

2. Then, create a new `GptClient` instance, either with default configurations, using a file for the API key, or with custom configurations:

```go
client := gpt.NewDefaultGptClient("your_openai_api_key")
```

3. Finally, send a message and retrieve the GPT-generated response:

```go
response, err := client.SenRandomContextMessage("What's the weather like today?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

## Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/assistant-ai/llmchat-client/gpt"
)

func main() {
	// Create a GptClient with default settings
	client := gpt.NewDefaultGptClient("your_openai_api_key")

	// Send a message and receive GPT-generated response
	response, err := client.SenRandomContextMessage("What's the weather like today?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
}
```

Feel free to dive into the source code (`gpt/client.go` and `db/db_client.go`) for a better understanding of how the GptClient works and various customization options.

This library makes it simple and straightforward to start building your own AI-powered chat applications using GPT-3/4. Enjoy!