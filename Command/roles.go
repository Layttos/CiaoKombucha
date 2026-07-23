package Command

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Role struct{}

func (c *Role) Name() string {
	return "roles"
}

func (c *Role) Description() string {
	return "Affiche le message gérant les rôles dans le channel courant pour les membres."
}

func (c *Role) Permissions() *int64 {
	var permissions = int64(discordgo.PermissionAdministrator) // Administrator permission
	return &permissions
}

func (c *Role) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool {

	s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Kyo Kombucha",
			IconURL: "https://www.kyokombucha.com/cdn/shop/files/kyo-logo-site-60px-blanc_1afe2068-cc51-4a8b-8473-51ef2a836dc8.png?v=1634890684&width=130",
		},
		Title:       "Modifier ses propres rôles",
		Description: "Pour modifier vos rôles, cliquez simplement sur l'une des réactions correspondantes aux rôles que vous souhaitez obtenir ou retirer. En retirant une réaction, vous retirerez le rôle correspondant et vice-versa. Vous pouvez également cliquer sur plusieurs réactions pour obtenir plusieurs rôles à la fois.\n\n**Rôles disponibles :**\n1. <:squeezie:1516591580977303552> → <@&1516563897610404044>\n2. <:cyprien:1516602309398761674> → <@&1522333169762828420>",
		Color:       0xC7A49D,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
	})

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Le message gérant les rôles a été envoyé dans le channel courant.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Erreur Deferred: %v", err)
		return false
	}
	return true
}
