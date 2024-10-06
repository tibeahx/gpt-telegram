package telegram

import (
	"sync"

	"gopkg.in/telebot.v3"
)

type session struct {
	store sync.Map
}

func newSession(ctx telebot.Context) *session {
	s := &session{}
	s.store.Store(ctx.Sender().ID, []string{})
	return s
}

func (s *session) flush(key int64) {
	s.store.Delete(key)
}

func (s *session) add(key int64, value string) {
	if key == 0 || value == "" {
		return
	}

	v, _ := s.store.LoadOrStore(key, []string{})
	values, _ := v.([]string)
	s.store.Store(key, append(values, value))

}

func (s *session) values(key int64) []string {
	if key == 0 {
		return nil
	}

	v, ok := s.store.Load(key)
	if !ok {
		return []string{}
	}

	values, _ := v.([]string)
	return values
}
