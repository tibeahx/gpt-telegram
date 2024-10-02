package openaix

import "github.com/sashabaranov/go-openai"

type chat struct {
	role    string
	context []openai.ChatCompletionMessage
}

func newChat() chat {
	role := openai.ChatMessageRoleUser
	return chat{context: make([]openai.ChatCompletionMessage, 0), role: role}
}

func (c chat) toCompletion(messages []string) []openai.ChatCompletionMessage {
	c.context = make([]openai.ChatCompletionMessage, len(messages))

	for _, msg := range messages {
		c.context = append(c.context, openai.ChatCompletionMessage{
			Role:    c.role,
			Content: msg,
		})
	}
	return c.context
}
