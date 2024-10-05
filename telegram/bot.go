package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tibeahx/gpt-helper/openaix"
	"gopkg.in/telebot.v3"
)

var (
	errInvalidSender = errors.New("invalid sender")
	errEmptyMsg      = errors.New("got empty message")
)

const (
	maxSessionCtxLenght = 100
	prompt              = "/prompt"
	clear               = "/clear"
	commands            = "/commands"
)

type Bot struct {
	tele           *telebot.Bot
	logger         *logrus.Logger
	openAi         *openaix.OpenAi
	cmdList        []string
	session        *session
	waitingForText map[int64]bool
}

func NewBot(token string, logger *logrus.Logger, openAi *openaix.OpenAi) (*Bot, error) {
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
		tele:           bot,
		logger:         logger,
		openAi:         openAi,
		cmdList:        []string{"/start", "/prompt", "/clear"},
		waitingForText: make(map[int64]bool),
	}, nil
}

func (b *Bot) manageSession(ctx telebot.Context) (int64, error) {
	if b.session == nil {
		b.session = newSession(ctx)
	}

	var (
		sender       = ctx.Sender()
		senderId     = ctx.Sender().ID
		textToAppend = ctx.Message().Text
	)

	if senderId == 0 || sender == nil {
		return 0, errInvalidSender
	}

	if len(textToAppend) == 0 {
		return 0, errEmptyMsg
	}

	if textToAppend[0] == '/' {
		b.logger.Warn("got command, will skip to add to session ctx")
		return senderId, nil
	}

	if len(b.session.values(senderId)) > maxSessionCtxLenght {
		b.logger.Infof("session will be flushed due to oversize\n current len: %d", len(b.session.values(senderId)))
		b.session.flush(senderId)
	}

	return senderId, nil
}

func (b *Bot) HandlePrompt(ctx telebot.Context) error {
	senderId, err := b.manageSession(ctx)
	if err != nil {
		return err
	}

	if ctx.Message().Text[0] == '/' && ctx.Message().Text == "/prompt" {
		err := ctx.Send("`enter your prompt`")
		if err != nil {
			return err
		}
		b.waitingForText[senderId] = true
	}

	if b.waitingForText[senderId] {
		b.session.add(senderId, ctx.Message().Text)

		completion, err := b.openAi.ReadPromptFromContext(
			context.Background(),
			ctx.Message().Text,
			b.session.values(ctx.Sender().ID),
		)
		if err != nil {
			return err
		}

		b.waitingForText[senderId] = false
		return ctx.Send(completion.Choices[0])
	}

	return nil
}

func (b *Bot) HandleClear(ctx telebot.Context) error {
	senderId, err := b.manageSession(ctx)
	if err != nil {
		return err
	}

	messages := b.session.values(senderId)
	if len(messages) == 0 {
		return ctx.Send("noting to delete, your saved messages == 0")
	}

	b.logger.Info("about to clear session messages")
	b.session.flush(senderId)

	return ctx.Send(fmt.Sprintf("flushed %d messages", len(messages)))
}

func (b *Bot) HandleCommands(ctx telebot.Context) error {
	msg := b.commands()
	if _, err := b.manageSession(ctx); err != nil {
		return err
	}
	return ctx.Send(msg)
}

func (b *Bot) start() {
	b.logger.Info("bot started...")
	b.tele.Start()
}

func (b *Bot) Run() {
	b.tele.Handle("/commands", func(ctx telebot.Context) error {
		return b.HandleCommands(ctx)
	})
	b.tele.Handle("/clear", func(ctx telebot.Context) error {
		return b.HandleClear(ctx)
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
		str += fmt.Sprintf("`\n%s`", cmd)
	}
	return str
}
