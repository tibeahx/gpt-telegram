package session

import (
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/telebot.v3"
)

type Session struct {
	id      string
	storage sync.Map
	logger  *logrus.Logger
}

func NewSession(ctx telebot.Context, logger *logrus.Logger) *Session {
	s := &Session{
		id:     uuid.New().String(),
		logger: logger,
	}
	s.storage.Store(ctx.Sender().ID, []string{})
	return s
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Flush(key int64) {
	s.storage.Delete(key)
}

func (s *Session) Add(key int64, value string) {
	if key == 0 || value == "" {
		return
	}
	v, _ := s.storage.LoadOrStore(key, []string{})
	values, _ := v.([]string)
	s.storage.Store(key, append(values, value))
}

func (s *Session) Values(key int64) []string {
	if key == 0 {
		return nil
	}
	v, ok := s.storage.Load(key)
	if !ok {
		return []string{}
	}
	values, _ := v.([]string)
	return values
}
