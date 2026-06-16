package Utils

import (
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

func AlertChannelMembers(s *discordgo.Session, memberID string, title string, content string) {
	user, _ := s.User(memberID)
	avatarURL := user.AvatarURL("")
	s.ChannelMessageSendEmbed(os.Getenv("USER_CHANNEL_ID"), &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: avatarURL,
			Name: user.GlobalName + " (@" + user.Username + ")",
		},
		Title: title,
		Color: 0xC7A49D,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func AlertChannelMembersComplex(s *discordgo.Session, embed *discordgo.MessageEmbed) {
	s.ChannelMessageSendEmbed(os.Getenv("USER_CHANNEL_ID"), embed)
}

func AlertChannelMessages(s *discordgo.Session, memberID string, title string, content string) {
	//member, _ := s.State.Member(os.Getenv("GUILD_ID"), memberID)
	user, _ := s.User(memberID)
	avatarURL := user.AvatarURL("")
	s.ChannelMessageSendEmbed(os.Getenv("MESSAGES_CHANNEL_ID"), &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: avatarURL,
			Name: user.GlobalName + " (@" + user.Username + ")",
		},
		Title: title,
		Color: 0x4A5B85,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func AlertChannelMessagesComplex(s *discordgo.Session, embed *discordgo.MessageEmbed) {
	//member, _ := s.State.Member(os.Getenv("GUILD_ID"), memberID)
	s.ChannelMessageSendEmbed(os.Getenv("MESSAGES_CHANNEL_ID"), embed)
}

func AlertChannelModeration(s *discordgo.Session, memberID string, title string, content string) {
	//member, _ := s.State.Member(os.Getenv("GUILD_ID"), memberID)
	user, _ := s.User(memberID)
	avatarURL := user.AvatarURL("")
	s.ChannelMessageSendEmbed(os.Getenv("MODERATION_CHANNEL_ID"), &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: avatarURL,
			Name: user.GlobalName + " (@" + user.Username + ")",
		},
		Title: title,
		Color: 0x7B53A3,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func AlertChannelModerationComplex(s *discordgo.Session, embed *discordgo.MessageEmbed) {
	//member, _ := s.State.Member(os.Getenv("GUILD_ID"), memberID)
	s.ChannelMessageSendEmbed(os.Getenv("MODERATION_CHANNEL_ID"), embed)
}