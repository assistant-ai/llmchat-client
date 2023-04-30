package gpt

const API_URL = "https://api.openai.com/v1/chat/completions"

type GPTModel struct {
	Name      string `json:"name"`
	MaxTokens int    `json:"max_tokens"`
}

var ModelGPT4 = &GPTModel{
	Name:      "gpt-4",
	MaxTokens: 8000,
}

var ModelGPT3Turbo = &GPTModel{
	Name:      "gpt-3.5-turbo",
	MaxTokens: 4000,
}
