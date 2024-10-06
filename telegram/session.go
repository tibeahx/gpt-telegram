package telegram

import (
	"errors"
	"sync"

	"gopkg.in/telebot.v3"
)

var errMissingKey = errors.New("missing key")

type session struct {
	mu    sync.RWMutex
	store map[int64][]string
}

func newSession(ctx telebot.Context) *session {
	s := &session{store: make(map[int64][]string)}
	s.store[ctx.Sender().ID] = []string{}
	return s
}

func (s *session) flush(key int64) error {
	if key == 0 {
		return errMissingKey
	}
	s.mu.Lock()
	delete(s.store, key)
	s.mu.Unlock()
	return nil
}

func (s *session) add(key int64, value string) {
	if key == 0 || value == "" {
		return
	}
	s.mu.Lock()
	if _, ok := s.store[key]; !ok {
		s.store[key] = []string{}
	}
	s.store[key] = append(s.store[key], value)
	s.mu.Unlock()
}

func (s *session) values(key int64) []string {
	if key == 0 {
		return nil
	}
	s.mu.RLock()
	if messages, ok := s.store[key]; ok {
		return messages
	}
	s.mu.RUnlock()
	return []string{}
}
