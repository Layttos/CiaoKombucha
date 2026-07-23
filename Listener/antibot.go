package Listener

import (
	"log"
	"os"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

func AntiBotListener(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author == nil || m.Author.Bot {
		return
	}

	antibotChannel := os.Getenv("ANTI_BOT_CHANNEL_ID")

	if antibotChannel != "" && m.ChannelID == antibotChannel {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			guildID = os.Getenv("GUILD_ID")
		}

		if guildID == "" {
			log.Println("Cannot fetch the GUILD ID to ban the compromised user.")
		}

		err := s.GuildBanCreateWithReason(guildID, m.Author.ID, "Utilisateur compromis", 7)
		if err != nil {
			log.Println("An error occurred while trying to ban the compromised user:", err)
		}
		Utils.AlertChannelMembers(s, m.Author.ID, ":no_entry: Anti-bot", "Un bot a été détecté et a été banni du serveur.")
	}

}
