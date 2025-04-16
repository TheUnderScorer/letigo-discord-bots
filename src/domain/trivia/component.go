package trivia

import (
	"app/messages"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

const TrueButtonId = "trivia-true"
const FalseButtonId = "trivia-false"
const Option = "trivia-option"
const NextParticipant = "trivia-next-participant"

type ComponentInteractionHandler struct {
	*Manager
}

func (h ComponentInteractionHandler) ShouldHandle(i *discordgo.InteractionCreate) bool {
	customID := i.MessageComponentData().CustomID

	return strings.HasPrefix(customID, "trivia")
}

func (h ComponentInteractionHandler) Handle(ctx context.Context, i *discordgo.InteractionCreate, session *discordgo.Session) error {
	trivia, ok := h.Manager.Get(i.ChannelID)
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

		return trivia.HandleChoice(index)
	}

	switch data.CustomID {
	case TrueButtonId:
		return trivia.HandleBoolean(True)
	case FalseButtonId:
		return trivia.HandleBoolean(False)
	case NextParticipant:
		var user *discordgo.User
		for _, u := range data.Resolved.Users {
			if u != nil {
				user = u
				break
			}
		}

		if user != nil {
			select {
			case trivia.PlayerNominated <- user:

			case <-ctx.Done():
				return ctx.Err()

			default:
				log.Warn("player nomination channel is not listening")
			}
		}
	}

	return nil
}

func HandleInteraction(manager *Manager, cid string, i *discordgo.InteractionCreate) error {
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

		return trivia.HandleChoice(index)
	}

	switch data.CustomID {
	case TrueButtonId:
		return trivia.HandleBoolean(True)
	case FalseButtonId:
		return trivia.HandleBoolean(False)
	case NextParticipant:
		var user *discordgo.User
		for _, u := range data.Resolved.Users {
			if u != nil {
				user = u
				break
			}
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

type QuestionComponentOpts struct {
	SelectedAnswer string
}

func GetQuestionComponent(q *Question, opts *QuestionComponentOpts) []discordgo.MessageComponent {
	var action discordgo.MessageComponent

	switch q.Type {
	case TrueFalse:
		styles := make(map[string]discordgo.ButtonStyle)

		if opts != nil && opts.SelectedAnswer != "" {
			var oppositeAnswer string
			if opts.SelectedAnswer == True {
				oppositeAnswer = False
			} else {
				oppositeAnswer = True
			}

			styles[q.Correct] = discordgo.SuccessButton

			if opts.SelectedAnswer != q.Correct {
				styles[oppositeAnswer] = discordgo.DangerButton
			} else {
				styles[oppositeAnswer] = discordgo.SecondaryButton
			}
		} else {
			styles[True] = discordgo.SecondaryButton
			styles[False] = discordgo.SecondaryButton
		}

		action = discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    styles[True],
					Label:    messages.Messages.Trivia.True,
					CustomID: TrueButtonId,
					Disabled: false,
				},
				discordgo.Button{
					Style:    styles[False],
					Label:    messages.Messages.Trivia.False,
					CustomID: FalseButtonId,
					Disabled: false,
				},
			},
		}

	case MultipleChoice:
		var choices []discordgo.MessageComponent

		for i, option := range q.Options() {
			var style discordgo.ButtonStyle
			if opts != nil && opts.SelectedAnswer != "" {
				switch option {
				case q.Correct:
					style = discordgo.SuccessButton

				case opts.SelectedAnswer:
					style = discordgo.DangerButton
				}
			} else {
				style = discordgo.SecondaryButton
			}

			choices = append(choices, discordgo.Button{
				Label:    option,
				Style:    style,
				CustomID: fmt.Sprintf("%s-%d", Option, i),
				Disabled: false,
			})
		}

		action = discordgo.ActionsRow{
			Components: choices,
		}
	}

	return []discordgo.MessageComponent{
		action,
	}
}
