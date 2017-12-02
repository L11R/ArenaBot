package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

var DuelKeyboard tgbotapi.InlineKeyboardMarkup

type User struct {
	ID             int     `gorethink:"id"`
	Username       string  `gorethink:"username"`
	FirstName      string  `gorethink:"firstName"`
	LastName       string  `gorethink:"lastName"`
	ActiveFight    Fight   `gorethink:"activeFight"`
	PreviousFights []Fight `gorethink:"previousFights"`
}

type Fight struct {
	ID      string    `gorethink:"id"`
	Time    time.Time `gorethink:"time"`
	UsersID []int     `gorethink:"usersId"`
	History []Round   `gorethink:"history"`
}

type Round struct {
	Time    time.Time     `gorethink:"time"`
	Players map[int]*Move `gorethink:"players"`
}

type Move struct {
	Hit   string `gorethink:"hit"`
	Block string `gorethink:"block"`
}

type RoundStatus struct {
	UserWaiting     map[int]*Waiting
	FightEnded      bool
	NewRoundWaiting bool
}

type Waiting struct {
	HitWaiting   bool
	BlockWaiting bool
}
