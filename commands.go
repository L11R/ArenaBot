package main

import "github.com/go-telegram-bot-api/telegram-bot-api"

func StartCommand(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "WIP")
	msg.ParseMode = "HTML"

	bot.Send(msg)

	_, err := InsertUser(update.Message.From.ID)
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
		ActiveFight: Fight{ID: res.GeneratedKeys[0]},
	})
	if err != nil {
		log.Warn(err)
		return
	}

	// Update second player info
	_, err = UpdateUser(User{
		ID:          update.Message.ReplyToMessage.From.ID,
		ActiveFight: Fight{ID: res.GeneratedKeys[0]},
	})
	if err != nil {
		log.Warn(err)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "<b>Дуэль начинается!</b>")
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
