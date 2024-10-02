package openaix

import "github.com/sashabaranov/go-openai"

type chat struct {
	role    string
	context []openai.ChatCompletionMessage
}

func newChat() chat {
	role := openai.ChatMessageRoleUser
	return chat{role: role}
}

func (c chat) toCompletion(messages []string) []openai.ChatCompletionMessage {
	c.context = make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		c.context[i] = openai.ChatCompletionMessage{
			Role:    c.role,
			Content: msg,
		}
	}
	return c.context
}
