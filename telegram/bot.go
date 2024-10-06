package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tibeahx/gpt-helper/openaix"
	"gopkg.in/telebot.v3"
)

var (
	errInvalidSender        = errors.New("invalid sender")
	errEmptyMsg             = errors.New("got empty message")
	errfailedProcessMessage = errors.New("failed to process message")
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

func (b *Bot) manageSession(c telebot.Context) (int64, error) {
	if b.session == nil {
		b.session = newSession(c)
	}

	var (
		sender      = c.Sender()
		senderId    = c.Sender().ID
		messageText = c.Message().Text
	)

	if senderId == 0 || sender == nil {
		return 0, errInvalidSender
	}

	if len(messageText) == 0 {
		return 0, errEmptyMsg
	}

	if messageText[0] == '/' {
		b.logger.Warn("got command, will skip adding to session ctx")
		return senderId, nil
	}

	if len(b.session.values(senderId)) > maxSessionCtxLenght {
		b.logger.Infof("session will be flushed due to oversize\n current len: %d", len(b.session.values(senderId)))
		b.session.flush(senderId)
	}

	return senderId, nil
}

func (b *Bot) processMessage(msg *telebot.Message, c telebot.Context) error {
	if msg.Text != "" {
		if msg.Text == "/clear" {
			return b.HandleClear(c)
		}
		return b.HandleText(c)
	}

	if msg.Voice != nil {
		return b.HandleVoice(c)
	}

	if msg.Media().MediaType() == "photo" {
		return b.HandlePhoto(c)
	}

	return nil
}

func (b *Bot) HandlePrompt(c telebot.Context) error {
	senderId, err := b.manageSession(c)
	if err != nil {
		return err
	}

	var (
		msg         = c.Message()
		messageText = c.Message().Text
	)

	if messageText[0] == '/' && messageText == "/prompt" {
		b.waitingForText[senderId] = true
		err := c.Send("`enter your prompt`")
		if err != nil {
			return err
		}
	}

	if err := b.processMessage(msg, c); err != nil {
		return errfailedProcessMessage
	}

	return nil
}

func (b *Bot) HandleText(c telebot.Context) error {
	var (
		messageText = c.Message().Text
		senderId    = c.Sender().ID
	)

	if messageText != "" && !strings.HasPrefix(messageText, "/") { //&& b.waitingForText[senderId]
		b.logger.Infof("got message: %s", messageText)

		// вот тут ловится дедлок, надо курить
		b.session.add(senderId, messageText)
		b.logger.Infof("added text: %s in session for used:%d", messageText, senderId)

		err := c.Send("`sending your message to openAI`")
		if err != nil {
			return err
		}

		completion, err := b.openAi.ReadPromptFromContext(
			context.Background(),
			messageText,
			b.session.values(c.Sender().ID),
			c,
		)
		if err != nil {
			return err
		}

		b.waitingForText[senderId] = false
		return c.Send(completion.Choices[0])
	}

	return nil
}

func (b *Bot) HandleVoice(c telebot.Context) error { return nil }

func (b *Bot) HandlePhoto(c telebot.Context) error { return nil }

func (b *Bot) HandleClear(c telebot.Context) error {
	senderId, err := b.manageSession(c)
	if err != nil {
		return err
	}

	messages := b.session.values(senderId)
	if len(messages) == 0 {
		return c.Send("noting to delete, your saved messages == 0")
	}

	b.logger.Info("about to clear session messages")
	b.session.flush(senderId)

	return c.Send(fmt.Sprintf("flushed %d messages", len(messages)))
}

func (b *Bot) HandleCommands(c telebot.Context) error {
	msg := b.commands()
	if _, err := b.manageSession(c); err != nil {
		return err
	}
	return c.Send(msg)
}

func (b *Bot) start() {
	b.logger.Info("bot started...")
	b.tele.Start()
}

func (b *Bot) Run() {
	b.tele.Handle("/commands", func(c telebot.Context) error {
		return b.HandleCommands(c)
	})

	b.tele.Handle("/clear", func(c telebot.Context) error {
		return b.HandleClear(c)
	})

	b.tele.Handle("/prompt", func(c telebot.Context) error {
		return b.HandlePrompt(c)
	})

	b.tele.Handle(telebot.OnText, func(c telebot.Context) error {
		return b.HandleText(c)
	})

	b.tele.Handle(telebot.OnVoice, func(c telebot.Context) error {
		return b.HandleVoice(c)
	})

	b.tele.Handle(telebot.OnPhoto, func(c telebot.Context) error {
		return b.HandlePhoto(c)
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
