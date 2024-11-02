package dailyreport

import (
	"env"
	"github.com/bwmarrin/discordgo"
)

func Reply(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID != env.Cfg.DailyReportChannelId ||
		m.Author.ID != env.Cfg.DailyReportTargetUserId {
		return
	}

}
