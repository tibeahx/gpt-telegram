package openaix

import (
	"context"
	"errors"
	"net/http"

	"github.com/tibeahx/go-openai"
	"github.com/tibeahx/gpt-helper/config"
	"github.com/tibeahx/gpt-helper/logger"
	"github.com/tibeahx/gpt-helper/session"
	"gopkg.in/telebot.v3"
)

const (
	GPT4o     = "gpt-4o"
	GPT4oMini = "gpt-4o-mini"
	Whisper1  = "whisper-1"
	TTS1      = "tts-1"
)

var (
	errNoPrompt = errors.New("got empty prompt")
	errNoValues = errors.New("got empty session.values")
	errNoPath   = errors.New("got empty path")
)

var log = logger.GetLogger()

type OpenAI struct {
	client *openai.Client
	chat   chat
}

func NewOpenAi(cfg *config.Config, httpClient *http.Client) *OpenAI {
	client := openai.NewClientWithConfig(openai.ConfigWithCustomHttpClient(
		cfg.BaseUrl,
		cfg.AiApikey,
		httpClient,
	))
	ai := &OpenAI{
		chat: newChat(),
	}
	if client != nil {
		ai.client = client
	}
	return ai
}

func (ai *OpenAI) UpdateHttpClient(cfg *config.Config, httpClient *http.Client) {
	aiConfig := openai.ConfigWithCustomHttpClient(cfg.BaseUrl, cfg.AiApikey, httpClient)
	ai.client = openai.NewClientWithConfig(aiConfig)
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
	log.Info("sent text prompt to api")
	if err != nil {
		return openai.ChatCompletionChoice{}, err
	}
	resp, err := ai.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     GPT4oMini, //change to GPT4o if needed
		MaxTokens: 500,
		Messages:  msgx,
	})
	if err != nil {
		return openai.ChatCompletionChoice{}, err
	}
	opts.Session.Add(opts.SenderID, resp.Choices[0].Message.Content)
	log.Info("got text response from api")
	return resp.Choices[0], nil
}

func (ai *OpenAI) Transcription(ctx context.Context, opts Options) (string, error) {
	if opts.Path == "" {
		return "", errNoPath
	}
	req := openai.AudioRequest{
		Model:    Whisper1,
		FilePath: opts.Path,
		Format:   openai.AudioResponseFormatText,
	}

	err := opts.C.Send("`waiting for openAI response...`")
	log.Info("sent audio prompt to api")
	if err != nil {
		return "", err
	}
	trans, err := ai.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}
	opts.Session.Add(opts.SenderID, trans.Text)
	log.Info("got audio response from api")
	return trans.Text, nil
}
