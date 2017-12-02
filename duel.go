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

	player, err := GetUser(update.CallbackQuery.From.ID)
	if err != nil {
		log.Warn(err)
		return
	}

	fight, err := GetFight(player.ActiveFight.ID)
	if err != nil {
		log.Warn(err)
		return
	}

	var opponentUserID int
	if fight.UsersID[0] == update.CallbackQuery.From.ID {
		opponentUserID = fight.UsersID[1]
	} else {
		opponentUserID = fight.UsersID[0]
	}

	opponent, err := GetUser(opponentUserID)
	if err != nil {
		log.Warn(err)
		return
	}

	status := GetRoundStatus(fight.UsersID, fight)

	// Start new round
	if status.NewRoundWaiting {
		if data[0] == "hit" {
			fight.History = append(fight.History, Round{
				Time: time.Now(),
				Players: map[int]*Move{
					player.ID: {
						Hit: data[1],
					},
					opponent.ID: {},
				},
			})
		}

		if data[0] == "block" {
			fight.History = append(fight.History, Round{
				Time: time.Now(),
				Players: map[int]*Move{
					player.ID: {
						Block: data[1],
					},
					opponent.ID: {},
				},
			})
		}
	}

	if status, ok := status.UserWaiting[player.ID]; ok {
		if status.HitWaiting && data[0] == "hit" {
			if user, ok := fight.History[len(fight.History)-1].Players[player.ID]; ok {
				user.Hit = data[1]
			}
		}
	}

	if status, ok := status.UserWaiting[player.ID]; ok {
		if status.BlockWaiting && data[0] == "block" {
			if user, ok := fight.History[len(fight.History)-1].Players[player.ID]; ok {
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
		summary := MakeSummary(fight, player, opponent, status)

		msg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			summary,
		)
		msg.ParseMode = "HTML"

		if !status.FightEnded {
			msg.ReplyMarkup = &DuelKeyboard
		}

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

func MakeSummary(fight Fight, player User, opponent User, status RoundStatus) string {
	var text string

	if status.FightEnded {
		text = "<b>Дуэль окончена!</b>\n"
	} else {
		text = "<b>Дуэль в прогрессе...</b>\n"
	}

	text += fmt.Sprintf("@%s VS @%s\n\n",
		strings.ToUpper(player.Username),
		strings.ToUpper(opponent.Username))

	translate := func(target string) string {
		if target == "head" {
			return "по голове"
		}

		if target == "body" {
			return "по корпусу"
		}

		if target == "legs" {
			return "по ногам"
		}

		return ""
	}

	playerTotalScore := 0
	opponentTotalScore := 0

	for _, round := range fight.History {
		playerRound := round.Players[player.ID]
		opponentRound := round.Players[opponent.ID]

		playerRoundScore := 0
		opponentRoundScore := 0

		text += round.Time.Format("<i>03:04:05 / 02.01.2006</i>\n")

		if opponentRound.Block == playerRound.Hit {
			text += fmt.Sprintf("<code>@%s</code> 🛡 нанёс удар <code>%s</code>, но попал в блок.\n", player.Username, translate(playerRound.Hit))
			opponentTotalScore++
			opponentRoundScore++
		} else {
			text += fmt.Sprintf("<code>@%s</code> 🗡 нанёс удар <code>%s</code>\n", player.Username, translate(playerRound.Hit))
			playerTotalScore++
			playerRoundScore++
		}

		if playerRound.Block == opponentRound.Hit {
			text += fmt.Sprintf("<code>@%s</code> 🛡 нанёс удар <code>%s</code>, но попал в блок.\n", opponent.Username, translate(opponentRound.Hit))
			playerTotalScore++
			playerRoundScore++
		} else {
			text += fmt.Sprintf("<code>@%s</code> 🗡 нанёс удар <code>%s</code>\n", opponent.Username, translate(opponentRound.Hit))
			opponentTotalScore++
			opponentRoundScore++
		}

		text += fmt.Sprintf("<b>%d : %d</b>\n\n", playerRoundScore, opponentRoundScore)
	}

	if status.FightEnded {
		if playerTotalScore > opponentTotalScore {
			text += fmt.Sprintf("<code>@%s</code> одержал победу над <code>@%s</code>\n<b>Итог: %d : %d</b>",
				player.Username,
				opponent.Username,
				playerTotalScore,
				opponentTotalScore)
		}

		if opponentTotalScore > playerTotalScore {
			text += fmt.Sprintf("<code>@%s</code> одержал победу над <code>@%s</code>\n<b>Итог: %d : %d</b>",
				opponent.Username,
				player.Username,
				opponentTotalScore,
				playerTotalScore)
		}

		if playerTotalScore == opponentTotalScore {
			text += fmt.Sprintf("Похоже на ничью!\n<b>Итог: %d : %d</b>",
				playerTotalScore,
				opponentTotalScore)
		}
	} else {
		if playerTotalScore > opponentTotalScore {
			text += fmt.Sprintf("<code>@%s</code> одерживает победу над <code>@%s</code>\n<b>Промежуточный итог: %d : %d</b>",
				player.Username,
				opponent.Username,
				playerTotalScore,
				opponentTotalScore)
		}

		if opponentTotalScore > playerTotalScore {
			text += fmt.Sprintf("<code>@%s</code> одерживает победу над <code>@%s</code>\n<b>Промежуточный итог: %d : %d</b>",
				opponent.Username,
				player.Username,
				opponentTotalScore,
				playerTotalScore)
		}

		if playerTotalScore == opponentTotalScore {
			text += fmt.Sprintf("Пока это похоже на ничью!\n<b>Промежуточный итог: %d : %d</b>",
				playerTotalScore,
				opponentTotalScore)
		}
	}

	return text
}
