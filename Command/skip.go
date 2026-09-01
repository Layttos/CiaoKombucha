package Command

import (
	"fmt"
	"os"

	"bot.ciaokombucha.tv/Radio"
	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

type Skip struct{}

func (c *Skip) Name() string {
	return "skip"
}

func (c *Skip) Description() string {
	return "Voter pour passer la musique en cours dans la radio."
}

func (c *Skip) Permissions() *int64 {
	return nil // accessible à tout le monde
}

func (c *Skip) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *Skip) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
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

	// Auditeurs humains présents dans le salon de la radio.
	var listenerIDs []string
	userInChannel := false
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID != radioChannelID || vs.UserID == s.State.User.ID {
			continue
		}
		listenerIDs = append(listenerIDs, vs.UserID)
		if vs.UserID == i.Member.User.ID {
			userInChannel = true
		}
	}

	if !userInChannel {
		radioRespond(s, i, "Tu dois être dans le salon de la radio pour voter.", true)
		return false
	}

	reached, tally, needed, already := Radio.AddSkipVote(i.Member.User.ID, listenerIDs)

	if !reached {
		if already {
			radioRespond(s, i, fmt.Sprintf("Tu as déjà voté. **%d/%d** votes pour passer la musique.", tally, needed), true)
		} else {
			radioRespond(s, i, fmt.Sprintf("🗳️ Vote enregistré : **%d/%d** pour passer la musique.", tally, needed), false)
		}
		return true
	}

	// Seuil atteint : on avance dans la file (prochain /play ou radio aléatoire).
	Utils.SafeGo("le passage au morceau suivant demandé par /skip", func() {
		Radio.AdvanceTrack(s)
	})

	radioRespond(s, i, fmt.Sprintf("⏭️ Musique passée (%d/%d votes).", tally, needed), false)
	return true
}

// radioRespond répond à une interaction avec un simple message texte, éphémère
// ou non. Partagé par les commandes radio (/skip, /play).
func radioRespond(s *discordgo.Session, i *discordgo.InteractionCreate, content string, ephemeral bool) {
	data := &discordgo.InteractionResponseData{Content: content}
	if ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
}
