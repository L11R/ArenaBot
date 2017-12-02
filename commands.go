package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func StartCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "WIP")
	msg.ParseMode = "HTML"

	bot.Send(msg)

	_, err := InsertUser(User{
		ID:        update.Message.From.ID,
		Username:  update.Message.From.UserName,
		FirstName: update.Message.From.FirstName,
		LastName:  update.Message.From.LastName,
	})
	if err != nil {
		log.Warn(err)
		return
	}
}

func DuelCommand(update tgbotapi.Update) {
	hitHead := "hit:head"
	hitBody := "hit:body"
	hitLegs := "hit:legs"
	blockHead := "block:head"
	blockBody := "block:body"
	blockLegs := "block:legs"

	res, err := InsertFight(update.Message.From.ID, update.Message.ReplyToMessage.From.ID)
	if err != nil {
		log.Warn(err)
		return
	}

	// Update first player info
	_, err = UpdateUser(User{
		ID:          update.Message.From.ID,
		Username:    update.Message.From.UserName,
		FirstName:   update.Message.From.FirstName,
		LastName:    update.Message.From.LastName,
		ActiveFight: Fight{ID: res.GeneratedKeys[0]},
	})
	if err != nil {
		log.Warn(err)
		return
	}

	// Update second player info
	_, err = UpdateUser(User{
		ID:          update.Message.ReplyToMessage.From.ID,
		Username:    update.Message.ReplyToMessage.From.UserName,
		FirstName:   update.Message.ReplyToMessage.From.FirstName,
		LastName:    update.Message.ReplyToMessage.From.LastName,
		ActiveFight: Fight{ID: res.GeneratedKeys[0]},
	})
	if err != nil {
		log.Warn(err)
		return
	}

	text := fmt.Sprintf(
		"<b>Дуэль начинается!</b>\n@%s VS @%s",
		strings.ToUpper(update.Message.From.UserName),
		strings.ToUpper(update.Message.ReplyToMessage.From.UserName))

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"

	DuelKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{
				Text:         "🗡в голову",
				CallbackData: &hitHead,
			},
			tgbotapi.InlineKeyboardButton{
				Text:         "🛡головы",
				CallbackData: &blockHead,
			},
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{
				Text:         "🗡по корпусу",
				CallbackData: &hitBody,
			},
			tgbotapi.InlineKeyboardButton{
				Text:         "🛡корпуса",
				CallbackData: &blockBody,
			},
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{
				Text:         "🗡по ногам",
				CallbackData: &hitLegs,
			},
			tgbotapi.InlineKeyboardButton{
				Text:         "🛡ног",
				CallbackData: &blockLegs,
			},
		),
	)
	msg.ReplyMarkup = DuelKeyboard

	bot.Send(msg)
}

func DuelErrorCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не ответили, либо ответили боту.")
	msg.ParseMode = "HTML"

	bot.Send(msg)
}
