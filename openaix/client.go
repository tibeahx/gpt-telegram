package openaix

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type OpenAi struct {
	client  *openai.Client
	logger  *logrus.Logger
	session []openai.Message
}

func NewOpenAi(token string, logger *logrus.Logger) OpenAi {
	client := openai.NewClient(token)
	return OpenAi{client: client, logger: logger}
}
