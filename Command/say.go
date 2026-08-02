package Command

import "github.com/bwmarrin/discordgo"

type Say struct{}

func (c *Say) Name() string {
	return "say"
}

func (c *Say) Description() string {
	return "Faire dire un message au bot"
}

func (c *Say) Permissions() *int64 {
	permissions := int64(discordgo.PermissionManageMessages)
	return &permissions
}

func (c *Say) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "Le message à faire dire au bot.",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "origine",
			Description: "Si pour répondre à un message",
			Required:    false,
		},
	}
}

func (c *Say) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool {

	content := i.ApplicationCommandData().Options[0].StringValue()
	response := ""

	if len(i.ApplicationCommandData().Options) > 1 {
		response = i.ApplicationCommandData().Options[1].StringValue()
	}

	if response != "" {
		s.ChannelMessageSendReply(i.ChannelID, content, &discordgo.MessageReference{
			MessageID: response,
			ChannelID: i.ChannelID,
			GuildID:   i.GuildID,
		})
	} else {
		s.ChannelMessageSend(i.ChannelID, content)
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Le message a été envoyé.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		return false
	}

	return true
}
