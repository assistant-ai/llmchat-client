package llama2_client

import (
	"os"
	"fmt"
	"github.com/nikolaydubina/llama2.go/llama2"
	nn "github.com/nikolaydubina/llama2.go/exp/nnfast"
	"github.com/sirupsen/logrus"
	"github.com/assistant-ai/llmchat-client/client"
	"github.com/assistant-ai/llmchat-client/db"
)

type Llama2Client struct {
	CheckpointPath string
	TokenizerPath     string
	Temperature float64
	Steps    int
	Topp float64
	Logger    *logrus.Logger
}

func NewLlama2Client(checkpointPath string, tokenizerPath string, temperature float64, steps int, topp float64, logger *logrus.Logger, contextDepth int, defaultContext string) *client.Client {
	return &client.Client{
		Client: &Llama2Client{
			CheckpointPath: checkpointPath,
			TokenizerPath:     tokenizerPath,
			Temperature: temperature,
			Steps:    steps,
			Topp:    topp,
			Logger:    logger,
		},
		ContextDepth:   contextDepth,
		DefaultContext: defaultContext,
		Logger:         logger,
	}
}

func (g *Llama2Client) SendMessages(messages []db.Message, context []string) ([]db.Message, error) {
	fmt.Printf("here")
	checkpointFile, err := os.OpenFile(g.CheckpointPath, os.O_RDONLY, 0)
	if err != nil {
		if (g.Logger != nil) {
			g.Logger.Fatalf("cannot read config: %s", err)
		}
		return nil, err
	}
	defer checkpointFile.Close()

	prompt := ""
	answer := ""

	for _, contextMsg := range context {
		prompt = prompt + "\n" + contextMsg
	}

	for _, msg := range messages {
		prompt = prompt + "\n" + msg.Content
	}

	config, err := llama2.NewConfigFromCheckpoint(checkpointFile)
	if err != nil {
		if (g.Logger != nil) {
			g.Logger.Fatalf("cannot read config: %s", err)
		}
	}
	if (g.Logger != nil) {
		g.Logger.Printf("config: %#v\n", config)
	}

	// "negative vocab size is hacky way of signaling unsahred weights. biy yikes" — @karpathy
	isSharedWeights := config.VocabSize > 0
	if config.VocabSize < 0 {
		config.VocabSize = -config.VocabSize
	}

	tokenizerFile, err := os.OpenFile(g.TokenizerPath, os.O_RDONLY, 0)
	if err != nil {
		if (g.Logger != nil) {
			g.Logger.Fatal(err)
		}
		return nil, err
	}
	defer tokenizerFile.Close()

	vocab := llama2.NewVocabFromFile(config.VocabSize, tokenizerFile)

	w := llama2.NewTransformerWeightsFromCheckpoint(config, checkpointFile, isSharedWeights)

	// right now we cannot run for more than config.SeqLen steps
	if g.Steps <= 0 || g.Steps > config.SeqLen {
		g.Steps = config.SeqLen
	}

	runState := llama2.NewRunState(config)

	promptTokens := vocab.Encode(prompt)

	var token int = 1 // 1 = BOS token in llama-2 sentencepiece
	var pos = 0
	for pos < g.Steps {
		fmt.Printf("pos %d/%d\n", pos, g.Steps)
		// forward the transformer to get logits for the next token
		llama2.Transformer(token, pos, config, runState, w)

		var next int
		if pos < len(promptTokens) {
			next = promptTokens[pos]
		} else {
			// sample the next token
			if g.Temperature == 0 {
				// greedy argmax sampling
				next = nn.ArgMax(runState.Logits)
			} else {
				// apply the temperature to the logits
				for q := 0; q < config.VocabSize; q++ {
					runState.Logits[q] /= float32(g.Temperature)
				}
				// apply softmax to the logits to the probabilities for next token
				nn.SoftMax(runState.Logits)
				// we now want to sample from this distribution to get the next token
				if g.Topp <= 0 || g.Topp >= 1 {
					// simply sample from the predicted probability distribution
					next = nn.Sample(runState.Logits)
				} else {
					// top-p (nucleus) sampling, clamping the least likely tokens to zero
					next = nn.SampleTopP(runState.Logits, float32(g.Topp))
				}
			}
		}
		pos++

		// data-dependent terminating condition: the BOS (1) token delimits sequences
		if next == 1 {
			break
		}

		// following BOS (1) token, sentencepiece decoder strips any leading whitespace
		var tokenStr string
		if token == 1 && vocab.Words[next][0] == ' ' {
			tokenStr = vocab.Words[next][1:]
		} else {
			tokenStr = vocab.Words[next]
		}
		answer = answer + "\n" + tokenStr

		// advance forward
		token = next
	}

	newMessage := db.CreateNewMessage(db.AssistentRoleNeam, answer, messages[0].ContextId)
	messages = append(messages, newMessage)

	return messages, nil
}
