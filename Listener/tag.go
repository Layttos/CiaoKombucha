package Listener

import (
	"fmt"
	"slices"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

func MemberUpdateTag(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {

	if m.User.ID == s.State.User.ID {
		return
	}

	if m.User.PrimaryGuild.IdentityGuildID == m.GuildID {
		if !slices.Contains(m.Member.Roles, "1483630365611393177") {
			err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, "1483630365611393177")
			Utils.AlertChannelMembers(s, m.Member.User.ID, ":rocket: Personne goatesque", m.Mention()+" a mis le tag RATP.")
			if err != nil {
				fmt.Println("An error occured while attemping to add the \"RATP Enjoyer\" role to the user ID", m.User.ID)
				return
			}
		}
	} else {
		if slices.Contains(m.Member.Roles, "1483630365611393177") {
			err := s.GuildMemberRoleRemove(m.GuildID, m.User.ID, "1483630365611393177")
			Utils.AlertChannelMembers(s, m.Member.User.ID, ":cry: Aïe :/", m.Mention()+" a retiré le tag RATP ...")
			if err != nil {
				fmt.Println("An error occured while attemping to add the \"RATP Enjoyer\" role to the user ID", m.User.ID)
				return
			}
		}
	}

}
