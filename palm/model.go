package palm

type PalmInstance struct {
	Context string `json:"context"`
	Examples []string `json:"examples"`
	Messages []PalmMessage `json:"messages"`
}

type PalmMessage struct {
	Author string `json:"author"`
	Content string `json:"content"`
}

type PredictPayload struct {
	Instances  []PalmInstance `json:"instances"`
	Parameters Parameters   `json:"parameters"`
}

type Parameters struct {
	Temperature    float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
	TopP            float64 `json:"topP"`
	TopK            int     `json:"topK"`
}

type PredictResponse struct {
	Predictions []struct {
		Candidates []PalmMessage `json:"candidates"`
	} `json:"predictions"`
}
