package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	r "gopkg.in/gorethink/gorethink.v3"
	"os"
	"strings"
)

var log = logrus.New()

var (
	bot     *tgbotapi.BotAPI
	session *r.Session
)

func main() {
	log.Formatter = new(logrus.TextFormatter)
	log.Info("Arena Bot started!")

	var err error

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN env variable not specified!")
	}

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	// Database pool init
	go InitConnectionPool()

	// Debug log
	bot.Debug = true

	log.Infof("authorized on account @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		// Inline handler
		if update.CallbackQuery != nil {
			go PlayerMove(update)
		}

		if update.Message == nil {
			continue
		}

		// userId for logger
		commandLogger := log.WithFields(logrus.Fields{"user_id": update.Message.From.ID})

		// Commands
		if strings.HasPrefix(update.Message.Text, "/start") {
			commandLogger.Info("command /start triggered")
			go StartCommand(update)
		}

		if strings.HasPrefix(update.Message.Text, "/duel") {
			commandLogger.Info("command /duel triggered")
			if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.IsBot == false {
				go DuelCommand(update)
			} else {
				go DuelErrorCommand(update)
			}
		}
	}
}
