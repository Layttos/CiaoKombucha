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

	if m == nil || m.Member == nil || m.Member.User == nil || m.BeforeUpdate == nil || m.BeforeUpdate.User == nil {
		return
	}

	getDisplayName := func(nick string, user *discordgo.User) string {
		if nick != "" {
			return nick
		}
		if user.GlobalName != "" {
			return user.GlobalName
		}
		return user.Username
	}

	if m.BeforeUpdate == nil {
		if strings.Compare(m.Member.Nick, m.Member.User.GlobalName) != 0 {
			Utils.AlertChannelMembers(s, m.Member.User.ID, ":arrow_up: Nouveau nickname", m.User.GlobalName+" a décidé de s'appeler `"+m.Member.Nick+"` sur le serveur")
		}
		return
	}

	if strings.Compare(m.BeforeUpdate.Nick, m.Member.Nick) != 0 {
		previousNick := getDisplayName(m.BeforeUpdate.Nick, m.User)
		newNickname := getDisplayName(m.Member.Nick, m.User)

		if strings.Compare(previousNick, newNickname) != 0 {
			if strings.Compare(m.Member.Nick, "") == 0 {
				Utils.AlertChannelMembers(s, m.Member.User.ID,
					"<:cyprien:1516602309398761674> Réinitialisation du nickname",
					fmt.Sprintf("%s a décidé de retirer son nickname sur le serveur\n`%s` -> `%s`", m.User.GlobalName, previousNick, newNickname))
			} else {
				Utils.AlertChannelMembers(s, m.Member.User.ID,
					":arrow_up: Changement de nickname",
					fmt.Sprintf("%s a changé de nickname sur le serveur\n`%s` -> `%s`", m.User.GlobalName, previousNick, newNickname))
			}
		}

	}

	if m.BeforeUpdate.User != nil && strings.Compare(m.BeforeUpdate.User.GlobalName, m.Member.User.GlobalName) != 0 {
		Utils.AlertChannelMembers(s, m.Member.User.ID, ":pencil2: Changement de pseudo", m.User.GlobalName+" a changé de pseudo\n `"+m.BeforeUpdate.User.GlobalName+"` -> `"+m.User.GlobalName+"`")
	}

	for _, roleID := range m.BeforeUpdate.Roles {
		if !slices.Contains(m.Member.Roles, roleID) {
			role, err := s.GuildRole(m.GuildID, roleID)
			if err == nil && role != nil {
				Utils.AlertChannelMembers(s, m.Member.User.ID,
					":8ball: Perte de rôle",
					fmt.Sprintf("Le membre %s vient de perdre le rôle %s", getDisplayName(m.Member.Nick, m.User), role.Mention()))
			}
		}
	}

	for _, roleID := range m.Member.Roles {
		if !slices.Contains(m.BeforeUpdate.Roles, roleID) {
			role, err := s.GuildRole(m.GuildID, roleID)
			if err == nil && role != nil {
				Utils.AlertChannelMembers(s, m.Member.User.ID,
					":balloon: Obtention de rôle",
					fmt.Sprintf("Le membre %s vient d'obtenir le rôle %s", getDisplayName(m.Member.Nick, m.User), role.Mention()))
			}
		}
	}

	previousAvatar := m.BeforeUpdate.User.Avatar
	newAvatar := m.Member.User.Avatar

	if previousAvatar != newAvatar {
		Utils.AlertChannelMembers(s, m.Member.User.ID,
			":frame_photo: Changement d'avatar",
			fmt.Sprintf("%s a changé d'avatar sur le serveur", getDisplayName(m.Member.Nick, m.User)))
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
