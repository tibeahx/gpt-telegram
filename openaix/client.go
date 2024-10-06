package openaix

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

const (
	GPT4o     = "gpt-4o"
	GPT4oMini = "gpt-4o-mini"
)

type OpenAi struct {
	client *openai.Client
	logger *logrus.Logger
	chat   chat
}

func NewOpenAi(token string, logger *logrus.Logger) *OpenAi {
	client := openai.NewClient(token)
	chat := newChat()
	return &OpenAi{client: client, logger: logger, chat: chat}
}

func (ai *OpenAi) ReadPromptFromContext(
	ctx context.Context,
	prompt string,
	messages []string,
	c telebot.Context,
) (openai.ChatCompletionResponse, error) {
	if ai.client == nil {
		return openai.ChatCompletionResponse{}, errors.New("openai client is not initalized")
	}

	msgx := ai.chat.toCompletion(messages)

	err := c.Send("`waiting for openAI response...`")
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}
	resp, err := ai.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     GPT4oMini,
		MaxTokens: 500,
		Messages:  msgx,
	})
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}

	return resp, nil
}
