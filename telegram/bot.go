package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tibeahx/gpt-helper/config"
	"github.com/tibeahx/gpt-helper/logger"
	"github.com/tibeahx/gpt-helper/openaix"
	"github.com/tibeahx/gpt-helper/session"

	"gopkg.in/telebot.v3"
)

var (
	errInvalidSender        = errors.New("invalid sender")
	errEmptyMsg             = errors.New("got empty message")
	errfailedProcessMessage = errors.New("failed to process message")
)

const (
	maxSessionCtxLenght = 100
	requestTimeout      = 5 * time.Second
	prompt              = "/prompt"
	clear               = "/clear"
	commands            = "/commands"
	start               = "/start"
)

type Bot struct {
	tele          *telebot.Bot
	openai        *openaix.OpenAI
	session       *session.InMemorySession
	waitingForMsg map[int64]struct{}
}

func NewBot(cfg *config.Config, openai *openaix.OpenAI) (*Bot, error) {
	opts := telebot.Settings{
		Token:   cfg.BotApikey,
		Poller:  &telebot.LongPoller{Timeout: 10 * time.Second},
		Verbose: true,
	}

	bot, err := telebot.NewBot(opts)
	if err != nil {
		return nil, err
	}
	return &Bot{
		tele:          bot,
		openai:        openai,
		waitingForMsg: make(map[int64]struct{}),
	}, nil
}

var log = logger.GetLogger()

func (b *Bot) manageSession(c telebot.Context) (int64, error) {
	if b.session == nil {
		b.session = session.NewSession(c)
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
		log.Warn("got command, will skip adding to session ctx")
		return senderId, nil
	}
	if len(b.session.Values(senderId)) > maxSessionCtxLenght {
		log.Infof("session will be flushed due to oversize\n current len: %d", len(b.session.Values(senderId)))
		b.session.Flush(senderId)
	}
	return senderId, nil
}

func (b *Bot) processMessage(msg *telebot.Message, c telebot.Context) error {
	if msg.Text != "" {
		if msg.Text == clear {
			return b.HandleClear(c)
		}
		return b.HandleText(c)
	}
	if msg.Voice != nil {
		return b.HandleVoice(c)
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
	if messageText[0] == '/' && messageText == prompt {
		b.waitingForMsg[senderId] = struct{}{}
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
	if b.session == nil {
		b.session = session.NewSession(c)
	}
	var (
		messageText = c.Message().Text
		senderId    = c.Sender().ID
	)

	if _, waiting := b.waitingForMsg[senderId]; waiting && messageText != "" && !strings.HasPrefix(messageText, "/") {
		log.Infof("got message: %s", messageText)
		err := c.Send("`sending your message to openAI`")
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		res, err := b.openai.ReadPromptFromContext(
			ctx,
			openaix.Options{
				Prompt:   messageText,
				C:        c,
				Session:  b.session,
				SenderID: senderId,
			},
		)
		if err != nil {
			return err
		}
		delete(b.waitingForMsg, senderId)
		return c.Send(res)
	}
	return nil
}

func (b *Bot) HandleVoice(c telebot.Context) error {
	if b.session == nil {
		b.session = session.NewSession(c)
	}
	var (
		file     = c.Message().Media().MediaFile()
		senderId = c.Sender().ID
	)

	if _, waiting := b.waitingForMsg[senderId]; waiting {
		log.Info("got new audio")
		err := c.Send("`sending your voice to openAI`")
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		files := openaix.NewFiles(b.tele)
		files.DownloadAsync(ctx, file, "ogg")
		defer files.Cleanup()

		path := files.Filepath()
		log.Infof("got filepath: %s", path)
		res, err := b.openai.Transcription(
			ctx,
			openaix.Options{
				Path:     path,
				C:        c,
				Session:  b.session,
				SenderID: senderId,
			},
		)
		if err != nil {
			return err
		}
		delete(b.waitingForMsg, senderId)
		return c.Send(res)
	}
	return nil
}

func (b *Bot) HandleClear(c telebot.Context) error {
	senderId, err := b.manageSession(c)
	if err != nil {
		return err
	}

	messages := b.session.Values(senderId)
	if len(messages) == 0 {
		return c.Send("nothing to delete, you have 0 messages in context")
	}

	log.Info("about to clear session messages")
	b.session.Flush(senderId)

	delete(b.waitingForMsg, senderId)
	return c.Send(fmt.Sprintf("flushed %d messages from context", len(messages)))
}

func (b *Bot) HandleCommands(c telebot.Context) error {
	msg := b.commands()
	if _, err := b.manageSession(c); err != nil {
		return err
	}
	return c.Send(msg)
}

var cmdList = []string{start, prompt, clear, commands}

func (b *Bot) commands() (str string) {
	str = "current commands are: "
	for _, cmd := range cmdList {
		if cmd == "" {
			return ""
		}
		str += fmt.Sprintf("`\n%s`", cmd)
	}
	return str
}

func (b *Bot) start() {
	log.Info("bot started...")
	b.tele.Start()
}

func (b *Bot) Stop() {
	log.Info("bot stopping...")
	b.tele.Stop()
}

func (b *Bot) initHandlers() {
	handlers := []struct {
		cmd     string
		handler telebot.HandlerFunc
	}{
		{commands, b.HandleCommands},
		{clear, b.HandleClear},
		{prompt, b.HandlePrompt},
		{telebot.OnText, b.HandleText},
		{telebot.OnVoice, b.HandleVoice},
	}

	for _, h := range handlers {
		b.tele.Handle(h.cmd, h.handler)
	}
}

func (b *Bot) Run() {
	b.initHandlers()
	b.start()
}
