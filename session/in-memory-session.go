package session

import (
	"sync"

	"github.com/google/uuid"
	"github.com/tibeahx/gpt-helper/logger"
	"gopkg.in/telebot.v3"
)

var log = logger.GetLogger()

type InMemorySession struct {
	id      string
	storage sync.Map
}

func NewSession(ctx telebot.Context) *InMemorySession {
	s := &InMemorySession{
		id: uuid.New().String(),
	}
	s.storage.Store(ctx.Sender().ID, []string{})
	return s
}

func (s *InMemorySession) ID() string {
	return s.id
}

func (s *InMemorySession) Flush(key int64) {
	s.storage.Delete(key)
}

func (s *InMemorySession) Add(key int64, value string) {
	if key == 0 || value == "" {
		return
	}
	v, _ := s.storage.LoadOrStore(key, []string{})
	values, _ := v.([]string)
	s.storage.Store(key, append(values, value))
	log.Infof("added value to session: %v", value)
}

func (s *InMemorySession) Values(key int64) []string {
	if key == 0 {
		return nil
	}
	v, ok := s.storage.Load(key)
	if !ok {
		return []string{}
	}
	values, _ := v.([]string)
	log.Infof("current values are: %v", values)
	return values
}
