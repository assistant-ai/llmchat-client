package gpt

const API_URL = "https://api.openai.com/v1/chat/completions"

type GPTModel struct {
	Name      string `json:"name"`
	MaxTokens int    `json:"max_tokens"`
}

var ModelGPT4 = &GPTModel{
	// gpt-4-0613 is a special model that fine tunned to support JSON output formatting
	Name:      "gpt-4-0613",
	MaxTokens: 8000,
}

var ModelGPT4Big = &GPTModel{
	Name:      "gpt-4-32k",
	MaxTokens: 32000,
}

var ModelGPT4Turbo = &GPTModel{
	Name:      "gpt-4-1106-preview",
	MaxTokens: 128000,
}

var ModelGPT4Vision = &GPTModel{
	Name:      "gpt-4-vision-preview",
	MaxTokens: 128000,
}

var ModelGPT3Turbo = &GPTModel{
	Name:      "gpt-3.5-turbo",
	MaxTokens: 4000,
}

var ModelGPT3TurboBig = &GPTModel{
	Name:      "gpt-3.5-turbo-16k",
	MaxTokens: 16000,
}
