package Command

import (
	"bot.ciaokombucha.tv/Radio"
	"github.com/bwmarrin/discordgo"
)

type RadioSearch struct{}

func (c *RadioSearch) Name() string {
	return "radiosearch"
}

func (c *RadioSearch) Description() string {
	return "Change la musique en cours de lecture dans la radio."
}

func (c *RadioSearch) Permissions() *int64 {
	permission := int64(discordgo.PermissionAdministrator)
	return &permission
}

func (c *RadioSearch) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "query",
			Description: "Le nom de la musique à rechercher.",
			Required:    true,
		},
	}
}

func (c *RadioSearch) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	query := i.ApplicationCommandData().Options[0].StringValue()
	err := Radio.ChangeCurrentTrack(query)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Une erreur est survenue lors du changement de la musique : " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return false
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "La musique a été changée avec succès !",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	return true
}
