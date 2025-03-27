package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func HasUser(users []*discordgo.User, userID string) bool {
	for _, user := range users {
		if user.ID == userID {
			return true
		}
	}

	return false
}

func Spoiler(contents string) string {
	return fmt.Sprintf("<||%s||>", contents)
}
