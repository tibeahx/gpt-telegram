package openaix

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"github.com/tibeahx/gpt-helper/session"
	"gopkg.in/telebot.v3"
)

const (
	GPT4o     = "gpt-4o"
	GPT4oMini = "gpt-4o-mini"
	Whisper1  = "whisper-1"
)

var (
	errNoPrompt = errors.New("got empty prompt")
	errNoValues = errors.New("got empty session.values")
	errNoPath   = errors.New("got empty path")
)

type OpenAI struct {
	client *openai.Client
	logger *logrus.Logger
	chat   chat
}

func NewOpenAi(token string, logger *logrus.Logger) *OpenAI {
	client := openai.NewClient(token)
	ai := &OpenAI{
		chat:   newChat(),
		logger: logger,
	}
	if client != nil {
		ai.client = client
	}
	return ai
}

type Options struct {
	Prompt   string
	C        telebot.Context
	Session  *session.InMemorySession
	SenderID int64
	Path     string
}

func (ai *OpenAI) ReadPromptFromContext(ctx context.Context, opts Options) (openai.ChatCompletionChoice, error) {
	if opts.Prompt == "" {
		return openai.ChatCompletionChoice{}, errNoPrompt
	}
	opts.Session.Add(opts.SenderID, opts.Prompt)
	messages := opts.Session.Values(opts.SenderID)
	if messages == nil {
		return openai.ChatCompletionChoice{}, errNoValues
	}
	msgx := ai.chat.toCompletion(messages)

	err := opts.C.Send("`waiting for openAI response...`")
	ai.logger.Info("sent text prompt to api")
	if err != nil {
		return openai.ChatCompletionChoice{}, err
	}
	resp, err := ai.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     GPT4oMini,
		MaxTokens: 500,
		Messages:  msgx,
	})
	if err != nil {
		return openai.ChatCompletionChoice{}, err
	}
	opts.Session.Add(opts.SenderID, resp.Choices[0].Message.Content)
	ai.logger.Info("got text response from api")
	return resp.Choices[0], nil
}

func (ai *OpenAI) Transcription(ctx context.Context, opts Options) (string, error) {
	if opts.Prompt == "" {
		return "", errNoPath
	}
	req := openai.AudioRequest{
		Model:    Whisper1,
		FilePath: opts.Prompt,
		Format:   openai.AudioResponseFormatText,
	}

	err := opts.C.Send("`waiting for openAI response...`")
	ai.logger.Info("sent audio prompt to api")
	if err != nil {
		return "", err
	}
	trans, err := ai.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}
	opts.Session.Add(opts.SenderID, trans.Text)
	ai.logger.Info("got audio response from api")
	return trans.Text, nil
}
