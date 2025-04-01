package discord

import (
	"app/env"
	"app/util/arrayutil"
	"github.com/bwmarrin/discordgo"
)

func ListVoiceChannelMembers(s *discordgo.Session, cid string) ([]*discordgo.Member, error) {
	guild, err := s.State.Guild(env.Env.GuildId)
	if err != nil {
		return nil, err
	}

	var members []*discordgo.Member
	var ids []string
	for _, member := range guild.VoiceStates {
		if member.ChannelID != cid {
			continue
		}

		ids = append(ids, member.UserID)
	}

	if len(ids) > 0 {
		fetchedMembers, err := s.GuildMembers(guild.ID, "", 1000)
		if err != nil {
			return nil, err
		}

		for _, member := range fetchedMembers {
			if member.User != nil && arrayutil.Includes(ids, member.User.ID) {
				members = append(members, member)
			}
		}
	}

	return members, nil
}
