package Listener

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

func MemberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if !slices.Contains(m.Member.Roles, "1443912368005320704") {
		err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, "1443912368005320704")
		Utils.AlertChannelMembers(s, m.Member.User.ID, ":+1: Nouveau membre", m.User.GlobalName+" a rejoint le serveur")
		if err != nil {
			fmt.Println("An error occured while attemping to add the \"RATP Enjoyer\" role to the user ID", m.User.ID)
			return
		}
	}
}

func MemberQuit(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	Utils.AlertChannelMembers(s, m.Member.User.ID, ":-1: Un membre a quitté", m.User.GlobalName+" a quitté le serveur")
}

func MemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {

	if m.BeforeUpdate == nil {
		if strings.Compare(m.Member.Nick, m.Member.User.GlobalName) != 0 {
			Utils.AlertChannelMembers(s, m.Member.User.ID, ":arrow_up: Nouveau nickname", m.User.GlobalName+" a décidé de s'appeler `"+m.Member.Nick+"` sur le serveur")
		}
		return
	}

	if strings.Compare(m.BeforeUpdate.Nick, m.Member.Nick) != 0 {
		previousNick := m.BeforeUpdate.Nick
		if strings.Compare(previousNick, "") == 0 {
			previousNick = m.User.GlobalName
		}

		newNickname := m.Member.Nick

		if strings.Compare(newNickname, "") == 0 {
			newNickname = m.User.GlobalName
		}

		if strings.Compare(newNickname, m.Member.User.GlobalName) == 0 {
			Utils.AlertChannelMembers(s, m.Member.User.ID, "<:cyprien:1516602309398761674> Réinitialisation du nickname", m.User.GlobalName+" a décidé de retirer son nickname sur le serveur\n `"+previousNick+"` -> `"+newNickname+"`")
		} else {
			Utils.AlertChannelMembers(s, m.Member.User.ID, ":arrow_up: Changement de nickname", m.User.GlobalName+" a changé de nickname sur le serveur\n `"+previousNick+"` -> `"+newNickname+"`")
		}

	}

	if m.BeforeUpdate.User.GlobalName != m.User.GlobalName {
		Utils.AlertChannelMembers(s, m.Member.User.ID, ":pencil2: Changement de pseudo", m.User.GlobalName+" a changé de pseudo\n `"+m.BeforeUpdate.User.GlobalName+"` -> `"+m.User.GlobalName+"`")
	}

	hasLostRole := false
	for _, role := range m.BeforeUpdate.Roles {
		if !slices.Contains(m.Member.Roles, role) {
			hasLostRole = true
			break
		}
	}

	hasGainedRole := false
	for _, role := range m.Member.Roles {
		if !slices.Contains(m.BeforeUpdate.Roles, role) {
			hasGainedRole = true
			break
		}
	}

	if hasLostRole {
		for _, role := range m.BeforeUpdate.Roles {
			role1, err := s.GuildRole(m.GuildID, role)
			if err == nil && role1 != nil && !slices.Contains(m.Member.Roles, role) {
				Utils.AlertChannelMembers(s, m.Member.User.ID, ":8ball: Perte de rôle", "Le membre "+m.Member.User.GlobalName+" vient de perdre le rôle "+role1.Mention())
				break
			}
		}
	}

	if hasGainedRole {
		for _, role := range m.Member.Roles {
			role1, err := s.GuildRole(m.GuildID, role)
			if err == nil && role1 != nil && !slices.Contains(m.BeforeUpdate.Roles, role) {
				Utils.AlertChannelMembers(s, m.Member.User.ID, ":balloon: Obtention de rôle", "Le membre "+m.Member.User.GlobalName+" vient d'obtenir le rôle "+role1.Mention())
				break
			}
		}
	}

}

func MemberBanned(s *discordgo.Session, m *discordgo.GuildBanAdd) {
	time.Sleep(500 * time.Millisecond)
	user, _ := s.User(m.User.ID)
	avatarURL := user.AvatarURL("")

	auditLog, err := s.GuildAuditLog(m.GuildID, "", "", 22 /* (c.f discordgo.AuditLogActionMemberBanAdd)*/, 1)

	if err != nil {
		fmt.Println("An error occured after a member was banned,", err)
	}

	moderator_id := "Non Spécifié(e)"
	reason := "Non Spécifié(e)"
	if len(auditLog.AuditLogEntries) > 0 {
		entry := auditLog.AuditLogEntries[0]
		if entry.TargetID == m.User.ID {
			moderator_id = entry.UserID
		}
		if entry.Reason != "" {
			reason = entry.Reason
		}
	}

	var mod string

	if strings.Compare("Non Spécifié(e)", moderator_id) != 0 {
		mod = "<@" + moderator_id + ">"
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: avatarURL,
			Name:    user.GlobalName + " (@" + user.Username + ")",
		},
		Title: ":tools: Membre banni(e)",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Modérateur",
				Value:  mod,
				Inline: true,
			},
			{
				Name:   "Banni(e)",
				Value:  m.User.Mention(),
				Inline: true,
			},
			{
				Name:   "Raison",
				Value:  reason,
				Inline: true,
			},
		},
		Color: 0x7B53A3,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	Utils.AlertChannelMembersComplex(s, embed)
}

func MemberKicked(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	time.Sleep(500 * time.Millisecond)
	user, _ := s.User(m.User.ID)
	avatarURL := user.AvatarURL("")

	auditLog, err := s.GuildAuditLog(m.GuildID, "", "", 20 /* (c.f discordgo.AuditLogActionMemberKick)*/, 1)

	if err != nil {
		fmt.Println("An error occured after a member was banned,", err)
	}

	moderator_id := "Non Spécifié(e)"
	reason := "Non Spécifié(e)"
	was_kicked := false
	if len(auditLog.AuditLogEntries) > 0 {
		entry := auditLog.AuditLogEntries[0]
		if entry.TargetID == m.User.ID {
			was_kicked = true
			moderator_id = entry.UserID
		}
		if entry.Reason != "" {
			reason = entry.Reason
		}
	}

	if was_kicked {

		var mod string

		if strings.Compare("Non Spécifié(e)", moderator_id) != 0 {
			fmt.Println("Member kicked #8")
			mod = "<@" + moderator_id + ">"
		}

		embed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				IconURL: avatarURL,
				Name:    user.GlobalName + " (@" + user.Username + ")",
			},
			Title: ":tools: Membre expulsé(e)",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Modérateur",
					Value:  mod,
					Inline: true,
				},
				{
					Name:   "Expulsé(e)",
					Value:  m.User.Mention(),
					Inline: true,
				},
				{
					Name:   "Raison",
					Value:  reason,
					Inline: true,
				},
			},
			Color: 0x7B53A3,
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "Layttos Industries© - Tous droits réservés.",
				IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		Utils.AlertChannelMembersComplex(s, embed)
	}
}
