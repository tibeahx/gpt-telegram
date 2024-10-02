package telegram

import "gopkg.in/telebot.v3"

type session map[int64][]string

func newSession(ctx telebot.Context) session {
	s := make(map[int64][]string)
	s[ctx.Sender().ID] = []string{}
	return s
}

func (s session) flush() session {
	s = make(session)
	return s
}

func (s session) values() []string {
	messages := make([]string, 0)
	for _, msg := range s {
		if len(msg) > 0 {
			messages = append(messages, msg...)
		}
	}
	return messages
}
