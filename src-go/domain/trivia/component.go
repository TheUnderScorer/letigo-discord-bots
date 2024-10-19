package trivia

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"src-go/messages"
	"src-go/util"
	"strconv"
	"strings"
)

const TrueButtonId = "trivia-true"
const FalseButtonId = "trivia-false"
const Option = "trivia-option"
const NextParticipant = "trivia-next-participant"

func HandleInteraction(context context.Context, cid string, i *discordgo.InteractionCreate) error {
	manager := context.Value(ManagerContextKey).(*Manager)
	if manager == nil {
		return errors.New("trivia manager is nil")
	}

	trivia, ok := manager.Get(cid)
	if !ok {
		return errors.New("trivia is nil")
	}

	data := i.MessageComponentData()

	if trivia.state.currentPlayer == nil || i.Member == nil || i.Member.User == nil || i.Member.User.ID != trivia.state.currentPlayer.ID {
		// Silently fail when different user invoked interaction, or if there is no current player
		return nil
	}

	if strings.HasPrefix(data.CustomID, Option) {
		index, err := strconv.Atoi(strings.Split(data.CustomID, "-")[2])
		if err != nil {
			return err
		}

		return trivia.HandleAnswer(index)
	}

	switch data.CustomID {
	case NextParticipant:
		var user *discordgo.User
		for _, u := range data.Resolved.Users {
			user = u
			break
		}

		if user != nil {
			trivia.PlayerNominated <- user
		}

	}

	return nil
}

func GetQuestionNominationComponent() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					MenuType: discordgo.UserSelectMenu,
					CustomID: NextParticipant,
					Disabled: false,
				},
			},
		},
	}
}

func GetQuestionComponent(q *Question) []discordgo.MessageComponent {
	var action discordgo.MessageComponent

	switch q.Type {
	case TrueFalse:
		action = discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.SuccessButton,
					Label:    messages.Messages.Trivia.True,
					CustomID: TrueButtonId,
					Disabled: false,
				},
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    messages.Messages.Trivia.False,
					CustomID: FalseButtonId,
					Disabled: false,
				},
			},
		}

	case MultipleChoice:
		var choices []discordgo.MessageComponent

		for i, option := range q.Options() {
			choices = append(choices, discordgo.Button{
				Label:    option,
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("%s-%d", Option, i),
				Disabled: false,
			})
		}
		choices = util.Shuffle(choices)

		action = discordgo.ActionsRow{
			Components: choices,
		}
	}

	return []discordgo.MessageComponent{
		action,
	}
}
