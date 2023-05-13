# LLMChat Client (Simplified Client For Chat Applications for GPT3/4 and PaLM Vertex AI Models)

This Go library provides a simplified and easy-to-use interface for building chat-based applications using GPT-3/4 or PaLM model from Google Vertex AI. With just a few lines of code, you can interact with the GPT-3/4 AI or PaLM model to generate context-aware responses.

## Features

- Simplified interface for GPT-3/4 and PaLM model interactions
- Message history and context management using local SQLite database

## Main Methods

The main functions provided by the `GptClient` and `PalmClient` structs include:

- `func NewDefaultGptClient(openAiKey string) *GptClient` - Create a new GptClient with default configurations.
- `func NewDefaultGptClientFromFile(openAiKeyFilePath string) (*GptClient, error)` - Create a new GptClient with default configurations, reading the API key from a file.
- `func NewGptClient(openAiKey string, contextDepth int, model GPTModel, defaultContext string, maxTokens int) *GptClient` - Create a new, fully customized GptClient.
- `func NewGptClientFromFile(openAiKeyFlePath string, contextDepth int, model GPTModel, defaultContext string, maxTokens int) (*GptClient, error)` - Create a new, fully customized GptClient, reading the API key from a file.
- `func NewDefaultTokenPalmClient(GCPProjectId string) (*client.Client, error)` - Create a new PalmClient with default configurations.
- `func NewPalmClient(GCPProjectId string, serviceAccountJsonPath string) (*client.Client, error)` - Create a new PalmClient with custom configuration.
- `func (c *client.Client) SenRandomContextMessage(message string) (string, error)` - Send a message using a random context.
- `func (c *client.Client) SendMessage(message string, inputContextId string) (string, error)` - Send a message using a specified or random context if none is provided.

## How to Use

1. Start by importing the `gpt` or `palm`, and `db` packages:

```go
import (
	"github.com/assistant-ai/llmchat-client/gpt"
	"github.com/assistant-ai/llmchat-client/palm"
	"github.com/assistant-ai/llmchat-client/db"
)
```

2. Then, create a new `GptClient` or `PalmClient` instance, either with default configurations, using a file for the API key, or with custom configurations:

```go
client := gpt.NewDefaultGptClient("your_openai_api_key")
palmClient, err := palm.NewDefaultTokenPalmClient("your_project_id", "service-acc-json-apth")
```

3. Finally, send a message and retrieve the LLM-generated response response:

```go
response, err := client.SenRandomContextMessage("What's the weather like today?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)

response2, err := palmClient.SendMessage("What's the weather like today?", "weather")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response2)
```

## Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/assistant-ai/llmchat-client/gpt"
	"github.com/assistant-ai/llmchat-client/palm"
)

func main() {
	// Create a GptClient with default settings
	client := gpt.NewDefaultGptClient("your_openai_api_key")

	// Create a PalmClient with default settings
	palmClient, err := palm.NewDefaultTokenPalmClient("your_project_id", "service-acc-json-apth")
	if err != nil {
		log.Fatal(err)
	}

	// Send a message and receive GPT-generated response
	response, err := client.SenRandomContextMessage("What's the weather like today?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("GPT Response:", response)

	// Send a message and receive PaLM-generated response
	response2, err := palmClient.SendMessage("What's the weather like today?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PaLM Response:", response2)
}

```

Feel free to dive into the source code (`gpt/client.go`, `palm/client.go`, and `db/db_client.go`) for a better understanding of how the GptClient, PalmClient, and various customization options work.

This library makes it simple and straightforward to start building your own AI-powered chat applications using GPT-3/4 and PaLM model from Google Vertex AI. Enjoy!
