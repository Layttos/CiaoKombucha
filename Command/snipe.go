package Command

import (
	"time"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

type Snipe struct{}

func (c *Snipe) Name() string {
	return "snipe"
}

func (c *Snipe) Description() string {
	return "Affiche le dernier message supprimé du channel courant."
}

func (c *Snipe) Permissions() *int64 {
	return nil
}

func (c *Snipe) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *Snipe) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	channel := i.ChannelID

	var content string
	var author_id string

	query := `SELECT * FROM deleted_messages WHERE channel_id = ?;`
	err := Utils.DB.QueryRow(query, channel).Scan(&channel, &content, &author_id)
	err_msg := `Aucun message supprimé n'a été trouvé.`
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &err_msg,
		})
		return false
	}

	user, _ := s.User(author_id)
	if user == nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &err_msg,
		})
		return false
	}

	name := user.Username + " (@" + user.Username + ")"
	if user.GlobalName != "" {
		name = user.GlobalName
	}

	avatarURL := user.AvatarURL("")

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: avatarURL,
			Name:    name,
		},
		Title: "<:sniper:1533008954257309797> Charlie Kirk",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Auteur",
				Value:  user.GlobalName,
				Inline: true,
			},
			{
				Name:   "Contenu",
				Value:  "`" + content + "`",
				Inline: true,
			},
		},
		Color: 0xFFFFFF,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	no_attachment_msg := "Le message d'origine est susceptible de contenir des pièces jointes, qui ne sont pas révélées."

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:  &[]*discordgo.MessageEmbed{embed},
		Content: &no_attachment_msg,
	})

	return true
}
