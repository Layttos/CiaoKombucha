package Listener

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func EmailBotJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m.User.Bot {
		return
	}

	if m.User.ID == "905084617265152072" {
		time.Sleep(500 * time.Millisecond)
		s.GuildMemberDeleteWithReason(m.GuildID, m.User.ID, "Exclusion automatique : Bot d'Email")
	}
}
