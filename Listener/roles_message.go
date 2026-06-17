package Listener

import (
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func RolesReactionsAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if strings.Compare(r.MessageID, os.Getenv("ROLES_MESSAGE_ID")) == 0 {
		if strings.Compare(r.Emoji.ID, "1516591580977303552") == 0 {
			s.GuildMemberRoleAdd(r.GuildID, r.UserID, "1516563897610404044")
		}
	}
}

func RolesReactionsRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if strings.Compare(r.MessageID, os.Getenv("ROLES_MESSAGE_ID")) == 0 {
		if strings.Compare(r.Emoji.ID, "1516591580977303552") == 0 {
			s.GuildMemberRoleRemove(r.GuildID, r.UserID, "1516563897610404044")
		}
	}
}
