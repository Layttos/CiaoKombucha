package Command

import (
	"fmt"
	"os"

	"bot.ciaokombucha.tv/Radio"
	"github.com/bwmarrin/discordgo"
)

type Play struct{}

func (c *Play) Name() string {
	return "play"
}

func (c *Play) Description() string {
	return "Choisir une musique à jouer dans la radio (la radio reprend ensuite)."
}

func (c *Play) Permissions() *int64 {
	return nil // accessible à tout le monde
}

func (c *Play) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "musique",
			Description: "Le nom de la musique à jouer.",
			Required:    true,
		},
	}
}

func (c *Play) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	radioChannelID := os.Getenv("RADIO_CHANNEL_ID")
	if radioChannelID == "" {
		radioRespond(s, i, "La radio n'est pas configurée.", true)
		return false
	}

	if i.Member == nil || i.Member.User == nil {
		radioRespond(s, i, "Commande utilisable uniquement sur le serveur.", true)
		return false
	}

	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		radioRespond(s, i, "Impossible de récupérer les informations du serveur.", true)
		return false
	}

	// L'utilisateur doit être dans le salon de la radio.
	userInChannel := false
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == radioChannelID && vs.UserID == i.Member.User.ID {
			userInChannel = true
			break
		}
	}
	if !userInChannel {
		radioRespond(s, i, "Tu dois être dans le salon de la radio pour choisir une musique.", true)
		return false
	}

	query := i.ApplicationCommandData().Options[0].StringValue()

	track, position, err := Radio.EnqueueRequest(query, i.Member.User.ID)
	if err != nil {
		radioRespond(s, i, "Impossible d'ajouter la musique : "+err.Error(), true)
		return false
	}

	when := fmt.Sprintf("Position dans la file : **%d**.", position)
	if position <= 1 {
		when = "Elle sera jouée juste après la musique en cours."
	}

	radioRespond(s, i, fmt.Sprintf("🎵 **%s** — %s ajouté à la file par %s. %s",
		track.Title, track.Artist, i.Member.User.Mention(), when), false)
	return true
}
