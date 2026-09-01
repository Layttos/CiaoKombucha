package Listener

import (
	"fmt"
	"strings"
	"time"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	query := `INSERT OR IGNORE INTO messages (id, channel_id, content, author_id) VALUES(?, ?, ?, ?);`
	stmt, err := Utils.DB.Prepare(query)
	if err != nil {
		fmt.Println("An error occured while attemping to prepare the message insertion:", err)
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(m.Message.ID, m.Message.ChannelID, m.Content, m.Author.ID); err != nil {
		fmt.Println("An error occured when a user sent a message:", err)
	}
}

func MessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	// Discord émet aussi un MESSAGE_UPDATE partiel, sans auteur, lorsqu'il
	// ajoute de lui-même un embed à un message (aperçu de lien par exemple).
	if m.Author == nil || s.State == nil || s.State.User == nil {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	previous_content := "Undefined"

	query := `SELECT content FROM messages WHERE id=?`

	err := Utils.DB.QueryRow(query, m.Message.ID).Scan(&previous_content)
	if err != nil {
		fmt.Println("An error occured when a user edited a message:", err)
		return
	}

	if strings.Compare(previous_content, "Undefined") == 0 {
		return
	}

	update_query := `UPDATE messages SET content=? WHERE id=?`
	stmt, err := Utils.DB.Prepare(update_query)
	if err != nil {
		fmt.Println("An error occured while attemping to prepare the message update:", err)
		return
	}
	defer stmt.Close()
	if _, err := stmt.Exec(m.Message.Content, m.Message.ID); err != nil {
		fmt.Println("An error occured when the program attempted to update the message in the database:", err)
		return
	}

	if strings.Compare(previous_content, m.Message.Content) == 0 {
		return
	}

	embed := &discordgo.MessageEmbed{
		Author: Utils.ActorEmbedAuthor(s, m.Author.ID),
		Title:  ":droplet: Modification de message",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Avant",
				Value:  "`" + previous_content + "`",
				Inline: true,
			},
			{
				Name:   "Après",
				Value:  "`" + m.Content + "`",
				Inline: true,
			},
		},
		Color:       0x4A5B85,
		Description: "\n[Aller voir le message](https://discord.com/channels/" + m.GuildID + "/" + m.ChannelID + "/" + m.ID + "/)",
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	Utils.AlertChannelMessagesComplex(s, embed)

	//Utils.AlertChannelMessages(s, ":droplet: Modification de message", m.Author.Mention() + " a modifié un message\nAvant : `JE RETIRE 2S LE PREVIOUS CONTENT`\nAprès : `" + m.Message.Content + "`\n[Aller voir le message](https://discord.com/channels/" + m.GuildID + "/" + m.ChannelID + "/" + m.ID + "/)")

}

func MessageDeleteBulk(s *discordgo.Session, m *discordgo.MessageDeleteBulk) {
	actorID, reason := Utils.ResolveAuditActor(s, m.GuildID, 73 /* cf -> discordgo.AuditLogActionMessageBulkDelete */, "")

	// Nettoyage des messages concernés dans la base locale.
	for _, id := range m.Messages {
		if _, err := Utils.DB.Exec(`DELETE FROM messages WHERE id = ?;`, id); err != nil {
			fmt.Println("An error occured while attempting to delete a bulk-removed message from the database:", err)
		}
	}

	embed := &discordgo.MessageEmbed{
		Author: Utils.ActorEmbedAuthor(s, actorID),
		Title:  ":wastebasket: Suppression groupée de messages",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Nombre", Value: fmt.Sprintf("%d messages", len(m.Messages)), Inline: true},
			{Name: "Channel", Value: "<#" + m.ChannelID + ">", Inline: true},
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
			{Name: "Raison", Value: Utils.ReasonOrDefault(reason), Inline: true},
		},
		Color: 0x4A5B85,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	Utils.AlertChannelMessagesComplex(s, embed)
}

func MessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {

	var author_id, content string

	query := `SELECT author_id, content FROM messages WHERE id = ?;`

	stmt, err := Utils.DB.Prepare(query)
	if err != nil {
		fmt.Println("An error occured while attemping to delete the message from the database")
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(m.Message.ID).Scan(&author_id, &content)

	time.Sleep(500 * time.Millisecond)

	var deleted_by string

	auditLog, err := s.GuildAuditLog(m.GuildID, "", "", 72 /* cf -> discordgo.AuditLogActionMessageDelete */, 1)
	if err != nil {
		fmt.Println("An error occured while trying to fetch audit logs:", err)
	}

	var is_self_delete bool = true

	if auditLog != nil && len(auditLog.AuditLogEntries) > 0 {
		entry := auditLog.AuditLogEntries[0]

		if entry.ActionType != nil && *entry.ActionType == discordgo.AuditLogActionMessageDelete {

			if entry.TargetID == author_id && entry.Options != nil && entry.Options.ChannelID == m.ChannelID {
				logTime, _ := discordgo.SnowflakeTimestamp(entry.ID)

				if time.Since(logTime) < 5*time.Second {
					is_self_delete = false
					deleted_by = "<@" + entry.UserID + ">"
				}
			}
		}
	}

	if is_self_delete {
		deleted_by = "<@" + author_id + ">"
	}

	delete_query := `DELETE FROM messages WHERE id = ?;`

	_, err = Utils.DB.Exec(delete_query, m.Message.ID)
	if err != nil {
		fmt.Println("An error occured while attemping to delete the message from the database")
	}

	user, err := s.User(author_id)
	if err != nil {
		fmt.Println("An error occured while attemping to fetch the user information")
		return
	}

	name := "@" + user.Username
	if user.GlobalName != "" {
		name = user.GlobalName + " (@" + user.Username + ")"
	}

	avatarURL := user.AvatarURL("")
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: avatarURL,
			Name:    name,
		},
		Title: ":speaking_head: Message supprimé",
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
			{
				Name:   "Channel",
				Value:  "<#" + m.ChannelID + ">",
				Inline: true,
			},
			{
				Name:   "Par",
				Value:  deleted_by,
				Inline: true,
			},
		},
		Color: 0x4A5B85,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	Utils.AlertChannelMessagesComplex(s, embed)

	/* Ajout dans la liste des messages supprimés APRÈS l'envoi du message */

	add_query := `INSERT INTO deleted_messages (channel_id, message, author_id) VALUES(?, ?, ?) ON CONFLICT DO UPDATE SET message = EXCLUDED.message, author_id = EXCLUDED.author_id;`
	_, err = Utils.DB.Exec(add_query, m.ChannelID, content, author_id)
	if err != nil {
		fmt.Println("An error occured while attemping to add the message to the deleted messages list:", err)
	}

}
