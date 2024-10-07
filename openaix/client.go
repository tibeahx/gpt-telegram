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
	errNilClient = errors.New("nil ai client")
	errNoPrompt  = errors.New("got empty prompt")
	errNoValues  = errors.New("got empty session.values")
	errNoPath    = errors.New("got empty path")
)

type OpenAi struct {
	client *openai.Client
	logger *logrus.Logger
	chat   chat
}

func NewOpenAi(token string, logger *logrus.Logger) *OpenAi {
	client := openai.NewClient(token)
	chat := newChat()
	return &OpenAi{
		client: client,
		logger: logger,
		chat:   chat,
	}
}

func (ai *OpenAi) ReadPromptFromContext(
	ctx context.Context,
	prompt string,
	c telebot.Context,
	session *session.Session,
	senderId int64,
) (openai.ChatCompletionChoice, error) {
	if ai.client == nil {
		return openai.ChatCompletionChoice{}, errNilClient
	}

	if prompt == "" {
		return openai.ChatCompletionChoice{}, errNoPrompt
	}

	session.Add(senderId, prompt)

	messages := session.Values(senderId)
	if messages == nil {
		return openai.ChatCompletionChoice{}, errNoValues
	}

	msgx := ai.chat.toCompletion(messages)

	err := c.Send("`waiting for openAI response...`")
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

	ai.logger.Info("got text response from api")

	return resp.Choices[0], nil
}

func (ai *OpenAi) Transcription(
	ctx context.Context,
	path string,
	prompt string,
	c telebot.Context,
	session *session.Session,
	senderId int64,
) (string, error) {
	if ai.client == nil {
		return "", errNilClient
	}

	if path == "" {
		return "", errNoPath
	}
	// положили историю сообщений текущее транкриптнутую
	session.Add(senderId, prompt)

	// получили все сообщения чтобы кастануть их непонятно во что
	messages := session.Values(senderId)
	if messages == nil {
		return "", errNoValues
	}

	// вот тут надо как то кастовать

	req := openai.AudioRequest{
		Model:    Whisper1,
		FilePath: path,
		Format:   openai.AudioResponseFormatText,
	}

	err := c.Send("`waiting for openAI response...`")
	ai.logger.Info("sent audio prompt to api")
	if err != nil {
		return "", err
	}

	trans, err := ai.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}

	ai.logger.Info("got audio response from api")

	return trans.Text, nil
}
