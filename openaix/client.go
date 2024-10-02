package openaix

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

const (
	GPT4o     = "gpt-4o"
	GPT4oMini = "gpt-4o-mini"
)

type OpenAi struct {
	client  *openai.Client
	logger  *logrus.Logger
	context []openai.ChatCompletionMessage
}

func NewOpenAi(token string, logger *logrus.Logger) *OpenAi {
	client := openai.NewClient(token)
	return &OpenAi{client: client, logger: logger}
}

func (ai *OpenAi) ReadPromptFromContext(
	ctx context.Context,
	prompt string,
	messages []string,
) (openai.ChatCompletionResponse, error) {
	resp, err := ai.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     GPT4oMini,
		MaxTokens: 500,
		Messages:  ai.context,
	})
	ai.logger.Info("sending prompt to openai...")
	if err != nil {
		return resp, err
	}
	return resp, nil
}
