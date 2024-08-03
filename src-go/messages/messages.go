package messages

import (
	_ "embed"
	"encoding/json"
	"go.uber.org/zap"
	"src-go/logging"
)

type DailyReportReminder struct {
	Afternoon []string `json:"afternoon"`
	Night     []string `json:"night"`
}

type TwinTailsReminder struct {
	Night []string `json:"night"`
}

type Player struct {
	NoMoreSongs        string   `json:"noMoreSongs"`
	ClearedQueue       string   `json:"clearedQueue"`
	AlreadyQueued      string   `json:"alreadyQueued"`
	Ended              []string `json:"ended"`
	AddedToQueue       []string `json:"addedToQueue"`
	AddedToQueueAsNext string   `json:"addedToQueueAsNext"`
	NowPlaying         []string `json:"nowPlaying"`
	AvailableCommands  string   `json:"availableCommands"`
	FailedToQueue      string   `json:"failedToQueue"`
}

type DailyReportReplies struct {
	MentalComments struct {
		Medium []string `json:"medium"`
		High   []string `json:"high"`
		Low    []string `json:"low"`
	} `json:"mentalComments"`
	TimeSpentComments struct {
		Medium []string `json:"medium"`
		High   []string `json:"high"`
		Low    []string `json:"low"`
	} `json:"timeSpentComments"`
	Greeting []string `json:"greeting"`
	Song     []string `json:"song"`
	Skipped  []string `json:"skipped"`
}

type messages struct {
	NotAQuestion         string              `json:"notAQuestion"`
	WhatAreYouSaying     string              `json:"whatAreYouSaying"`
	MustBeInVoiceChannel string              `json:"mustBeInVoiceChannel"`
	UnknownCommand       string              `json:"unknownCommand"`
	UnknownError         string              `json:"unknownError"`
	DailyReportReminder  DailyReportReminder `json:"dailyReportReminder"`
	TwinTailsReminder    TwinTailsReminder   `json:"twinTailsReminder"`
	InvalidUrl           string              `json:"invalidUrl"`
	Player               Player              `json:"player"`
	Answers              []string            `json:"answers"`
	Insults              []string            `json:"insults"`
	WhatsUpReplies       []string            `json:"whatsUpReplies"`
	Greetings            [][]string          `json:"greetings"`
	DailyReportReplies   DailyReportReplies  `json:"dailyReportReplies"`
}

var Messages messages

//go:embed messages.json
var f []byte

var log = logging.Get().Named("messages")

func Init() {
	err := json.Unmarshal(f, &Messages)
	if err != nil {
		log.Fatal("failed to unmarshal messages", zap.Error(err))
	}
}
