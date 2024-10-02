package telegram

import (
	"errors"

	"gopkg.in/telebot.v3"
)

var errMissingKey = errors.New("missing key")

type session map[int64][]string

func newSession(ctx telebot.Context) session {
	s := make(map[int64][]string)
	s[ctx.Sender().ID] = []string{}
	return s
}

func (s session) flush(key int64) error {
	if key == 0 {
		return errMissingKey
	}
	delete(s, key)
	return nil
}

func (s session) add(key int64, value string) {
	if key <= 0 || value == "" {
		return
	}
	if _, ok := s[key]; !ok {
		s[key] = []string{}
	}
	s[key] = append(s[key], value)
}

func (s session) values(key int64) []string {
	if key == 0 {
		return nil
	}
	if messages, ok := s[key]; ok {
		return messages
	}
	return []string{}
}
