package openaix

import "github.com/sashabaranov/go-openai"

type chat struct {
	role    string
	history []openai.ChatCompletionMessage
	voices  []openai.AudioRequest
}

func newChat() chat {
	role := openai.ChatMessageRoleUser
	return chat{role: role}
}

func (c chat) toCompletion(messages []string) []openai.ChatCompletionMessage {
	c.history = make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		c.history[i] = openai.ChatCompletionMessage{
			Role:    c.role,
			Content: msg,
		}
	}
	return c.history
}
