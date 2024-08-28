package telegram

import (
	"context"
	"errors"
	"go-tg/clients/telegram"
	"go-tg/lib/e"
	"go-tg/storage"
	"log"
	"net/url"
	"strings"
)

const (
	RndCmd = "/rnd"
	HelpCmd = "/help"
	StartCmd = "/start"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)
	
	log.Printf("got new command '%s' from '%s'", text, username)

	if isAddCommand(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username)
	case HelpCmd:
		return p.SendHelp(chatID)
	case StartCmd:
		return p.SendHello(chatID)
	default:
		return p.tg.SendMessage(chatID, msgUnkownCommand)

	}
}

func (p *Processor) savePage(chatID int, pageUrl string, username string) (err error) {
	defer func() {err = e.WrapIfErr("can't do command: save page", err)} ()

	sendMsg := NewMessageSender(chatID, p.tg)

	page := &storage.Page{
		URL: pageUrl,
		UserName: username,
	}

	isExist, err := p.storage.IsExists(context.Background(), page)

	if err != nil {
		return err
	}

	if isExist {
		return sendMsg(msgAlreadyExists)
	}

	if err := p.storage.Save(context.Background(), page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() {err = e.WrapIfErr("can't do command: send random", err)} ()

	page, err := p.storage.PickRandom(context.Background(), username)

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(chatID, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(context.Background(), page)
}

func (p *Processor) SendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}
func (p *Processor) SendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}


func NewMessageSender(chatID int, tg *telegram.Client) func(string) error {
	return func(msg string) error {
		return tg.SendMessage(chatID, msg)
	}
}

func isAddCommand(text string) bool {
	return isUrl(text)
}

func isUrl(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
} 