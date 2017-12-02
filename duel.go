package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
	"time"
)

func PlayerMove(update tgbotapi.Update) {
	data := strings.Split(update.CallbackQuery.Data, ":")
	log.Infof("id%d: %s, %s", update.CallbackQuery.From.ID, data[0], data[1])

	user, err := GetUser(update.CallbackQuery.From.ID)
	if err != nil {
		log.Warn(err)
		return
	}

	fight, err := GetFight(user.ActiveFight.ID)
	if err != nil {
		log.Warn(err)
		return
	}

	status := GetRoundStatus(fight.UsersID, fight)

	// Start new round
	if status.NewRoundWaiting {
		var otherUserID int
		if fight.UsersID[0] == update.CallbackQuery.From.ID {
			otherUserID = fight.UsersID[1]
		} else {
			otherUserID = fight.UsersID[0]
		}

		if data[0] == "hit" {
			fight.History = append(fight.History, Round{
				Time: time.Now(),
				Players: map[int]*Move{
					update.CallbackQuery.From.ID: {
						Hit: data[1],
					},
					otherUserID: {},
				},
			})
		}

		if data[0] == "block" {
			fight.History = append(fight.History, Round{
				Time: time.Now(),
				Players: map[int]*Move{
					update.CallbackQuery.From.ID: {
						Block: data[1],
					},
					otherUserID: {},
				},
			})
		}
	}

	if status, ok := status.UserWaiting[update.CallbackQuery.From.ID]; ok {
		if status.HitWaiting && data[0] == "hit" {
			if user, ok := fight.History[len(fight.History)-1].Players[update.CallbackQuery.From.ID]; ok {
				user.Hit = data[1]
			}
		}
	}

	if status, ok := status.UserWaiting[update.CallbackQuery.From.ID]; ok {
		if status.BlockWaiting && data[0] == "block" {
			if user, ok := fight.History[len(fight.History)-1].Players[update.CallbackQuery.From.ID]; ok {
				user.Block = data[1]
			}
		}
	}

	// Update fight in DB
	_, err = UpdateFight(fight)
	if err != nil {
		log.Warn(err)
		return
	}

	// Recheck status and update message
	status = GetRoundStatus(fight.UsersID, fight)
	if status.NewRoundWaiting || status.FightEnded {
		msg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			MakeSummary(fight, status),
		)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = &DuelKeyboard
		bot.Send(msg)
	}

	bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Done!",
	})
}

func GetRoundStatus(ids []int, fight Fight) RoundStatus {
	var status RoundStatus

	if len(fight.History) == 0 {
		status.NewRoundWaiting = true
	} else {
		moves := []*Move{
			fight.History[len(fight.History)-1].Players[ids[0]],
			fight.History[len(fight.History)-1].Players[ids[1]],
		}

		status.UserWaiting = map[int]*Waiting{
			ids[0]: {},
			ids[1]: {},
		}

		for i, move := range moves {
			if move == nil {
				status.UserWaiting = map[int]*Waiting{
					ids[i]: {
						HitWaiting:   true,
						BlockWaiting: true,
					},
				}
			} else {
				if move.Hit == "" {
					status.UserWaiting[ids[i]].HitWaiting = true
				}

				if move.Block == "" {
					status.UserWaiting[ids[i]].BlockWaiting = true
				}
			}
		}

		if !status.UserWaiting[ids[0]].HitWaiting &&
			!status.UserWaiting[ids[0]].BlockWaiting &&
			!status.UserWaiting[ids[1]].HitWaiting &&
			!status.UserWaiting[ids[1]].BlockWaiting {
			if len(fight.History) < 5 {
				status.NewRoundWaiting = true
			} else {
				status.FightEnded = true
			}
		}
	}

	return status
}

func MakeSummary(fight Fight, status RoundStatus) string {
	var text string

	if status.FightEnded {
		text = "<b>Дуэль окончена!</b>\n\n"
	} else {
		text = "<b>Дуэль в прогрессе...</b>\n\n"
	}

	for _, round := range fight.History {
		text += round.Time.Format("<i>03:04:05 / 02.01.2006</i>\n")
		for _, id := range fight.UsersID {
			text += fmt.Sprintf("%d: Hit: %s, Block: %s\n", id, round.Players[id].Hit, round.Players[id].Block)
		}
		text += "\n"
	}

	return text
}
