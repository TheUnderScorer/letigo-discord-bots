package trivia

import (
	"app/messages"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"lib/aws"
	"lib/discord"
	"lib/util/arrayutil"
	"strings"
)

var ErrNoMoreQuestions = errors.New("no more questions")

type State struct {
	players         []*discordgo.User
	currentPlayer   *discordgo.User
	previousPlayer  *discordgo.User
	questions       []Question
	currentQuestion *Question
	points          map[string]int
}

func NewState(s3 *aws.S3, players []*discordgo.User) *State {
	return &State{
		players:       players,
		questions:     GetQuestions(s3),
		currentPlayer: arrayutil.RandomElement(players),
		points:        make(map[string]int),
	}
}

func (s *State) GetCurrentPlayerMemberDefinition() *discord.MemberDefinition {
	friend, ok := discord.Friends[s.currentPlayer.ID]
	if !ok {
		return nil
	}
	return &friend
}

func (s *State) GetWinners() []*discordgo.User {
	var highScore int
	var winners []*discordgo.User

	for _, point := range s.points {
		if point > highScore {
			highScore = point
		}
	}

	for _, player := range s.players {
		if s.points[player.ID] == highScore {
			winners = append(winners, player)
		}
	}

	return winners
}

func (s *State) HasMoreQuestions() bool {
	return len(s.questions) > 0
}

func (s *State) GetNextQuestion() (*Question, error) {
	if len(s.questions) == 0 {
		return nil, ErrNoMoreQuestions
	}

	question := s.questions[0]
	s.questions = s.questions[1:]

	s.currentQuestion = &question
	return &question, nil
}

func (s *State) AddPointToCurrentPlayer() {
	s.AddPoint(s.currentPlayer.ID)
}

func (s *State) AddPoint(pid string) {
	state, ok := s.points[pid]
	if !ok {
		state = 0
	}
	s.points[pid] = state + 1
}

func (s *State) ChangePlayerToRandom() {
	if len(s.players) <= 1 {
		return
	}

	var player *discordgo.User

	for {
		player = arrayutil.RandomElement(s.players)
		if player.ID != s.currentPlayer.ID {
			s.ChangePlayer(player)
			break
		}
	}
}

func (s *State) SetStartingPlayer(player *discordgo.User) {
	s.currentPlayer = player
}

func (s *State) ChangePlayer(player *discordgo.User) bool {
	s.previousPlayer = s.currentPlayer
	s.currentPlayer = player

	return true
}

func (s *State) GetPointsMessageContents() string {
	var msg []string
	msg = append(msg, messages.Messages.Trivia.PointsSummary)

	for pid, points := range s.points {
		player, ok := discord.Friends[pid]
		if !ok {
			continue
		}

		msg = append(msg, fmt.Sprintf("- %s: %d", player.Mention(), points))
	}

	return strings.Join(msg, "\n")
}
