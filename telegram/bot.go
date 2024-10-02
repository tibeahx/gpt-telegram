package telegram

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tibeahx/gpt-helper/openaix"
	"gopkg.in/telebot.v3"
)

// 1) Получаем сообщение (промт) от пользователя.
// 2) Если сессии нет, создаем новую для этого пользователя, иначе обновляем существующую.
// 3) Сохраняем текст сообщения и ID в сессии пользователя.
// 4) Формируем запрос к GPT с учетом контекста из сессии.
// 5) Отправляем запрос на сервер GPT.
// 6) Получаем ответ от GPT и обрабатываем его.
// 7) Сохраняем ответ в сессию, привязывая к ID пользователя.
// 8) Отправляем ответ пользователю из сессии.
// Дополнительно: обработка ошибок, очистка старых сессий, ограничение длины истории сообщений.

var (
	errEmptyMsg      = errors.New("got empty message")
	errInvalidSender = errors.New("invalid sender")
)

type Bot struct {
	tele    *telebot.Bot
	logger  *logrus.Logger
	openAi  openaix.OpenAi
	cmdList []string
	session session
}

func (b *Bot) manageSession(ctx telebot.Context) error {
	if ctx.Sender().ID == 0 || ctx.Sender() == nil {
		return errInvalidSender
	}

	if b.session == nil {
		b.session = newSession(ctx)
	}

	var (
		senderId     = ctx.Sender().ID
		textToAppend = ctx.Message().Text
	)

	if messages, ok := b.session[senderId]; ok {
		b.session[senderId] = append(messages, textToAppend)
		b.logger.Infof("new message received, total messages: %d", len(b.session.values()))
	}

	if len(b.session.values()) > 100 {
		b.logger.Infof("session will be shuled due to oversize\n current len: %d", len(b.session.values()))
		b.session.flush()
	}

	return nil
}

func NewBot(token string, logger *logrus.Logger, openAi openaix.OpenAi) (*Bot, error) {
	opts := telebot.Settings{
		Token:   token,
		Poller:  &telebot.LongPoller{Timeout: 10 * time.Second},
		Verbose: true,
	}

	bot, err := telebot.NewBot(opts)
	if err != nil {
		return nil, err
	}

	return &Bot{
		tele:    bot,
		logger:  logger,
		openAi:  openAi,
		cmdList: []string{"/start", "/tune", "/prompt"},
	}, nil
}

func (b *Bot) HandleCommands(ctx telebot.Context) error {
	msg := b.commands()
	if err := b.manageSession(ctx); err != nil {
		return err
	}
	return ctx.Send(msg)
}

func (b *Bot) HandleTune(ctx telebot.Context) error { return nil }

func (b *Bot) HandlePrompt(ctx telebot.Context) error {
	if ctx.Message().Text == "nil" {
		return errEmptyMsg
	}
	return nil
}

func (b *Bot) start() {
	b.logger.Info("bot started...")
	b.tele.Start()
}

func (b *Bot) Run() {
	b.tele.Handle("/commands", func(ctx telebot.Context) error {
		return b.HandleCommands(ctx)
	})
	b.tele.Handle("/tune", func(ctx telebot.Context) error {
		return b.HandleTune(ctx)
	})
	b.tele.Handle("/prompt", func(ctx telebot.Context) error {
		return b.HandlePrompt(ctx)
	})
	b.start()
}

func (b *Bot) commands() (str string) {
	str = "current commands are: "
	for _, cmd := range b.cmdList {
		if cmd == "" {
			return ""
		}
		str += fmt.Sprintf("\n%s", cmd)
	}
	return str
}
