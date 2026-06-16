package Listener

import (
	"fmt"
	"strings"
	"time"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.Compare("727939986175033346", m.Author.ID) > 0 {
		query := `INSERT OR IGNORE INTO messages (id, channel_id, content, author_id) VALUES(?, ?, ?, ?);`
		stmt, _ := Utils.DB.Prepare(query)
		defer stmt.Close()
		_, err := stmt.Exec(m.Message.ID, m.Message.ChannelID, m.Content, m.Author.ID)
		if err != nil {
			fmt.Println("An error occured when a user sent a message:", err)
		}
	}
}

func MessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
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
	stmt, _ := Utils.DB.Prepare(update_query)
	defer stmt.Close()
	_, err = stmt.Exec(m.Message.Content, m.Message.ID)
	if err != nil {
		fmt.Println("An error occured when the program attempted to update the message in the database:", err)
	}


	user, _ := s.User(m.Author.ID)
	avatarURL := m.Message.Author.AvatarURL("")
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: avatarURL,
			Name: user.GlobalName + " (@" + user.Username + ")",
		},
		Title: ":droplet: Modification de message",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Avant",
				Value: "`" + previous_content + "`",
				Inline: true,
			},
			{
				Name: "Après",
				Value: "`" + m.Content + "`",
				Inline: true,
			},
		},
		Color: 0x4A5B85,
		Description: "\n[Aller voir le message](https://discord.com/channels/" + m.GuildID + "/" + m.ChannelID + "/" + m.ID + "/)",
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	Utils.AlertChannelMessagesComplex(s, embed)

	//Utils.AlertChannelMessages(s, ":droplet: Modification de message", m.Author.Mention() + " a modifié un message\nAvant : `JE RETIRE 2S LE PREVIOUS CONTENT`\nAprès : `" + m.Message.Content + "`\n[Aller voir le message](https://discord.com/channels/" + m.GuildID + "/" + m.ChannelID + "/" + m.ID + "/)")


}

func MessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {

	deleted_by := "Non Spécifié(e)"
	auditLog, err := s.GuildAuditLog(m.GuildID, "", "", 72 /* cf -> discordgo.AuditLogActionMessageDelete */, 1)
	if err != nil {
		fmt.Println("An error occured while trying to fetch some data information on a message suppression")
	}

	if len(auditLog.AuditLogEntries) > 0 {
		entry := auditLog.AuditLogEntries[0]
		if entry.UserID != "" {
			usr, _ := s.User(entry.UserID)
			deleted_by = usr.GlobalName
		}
	}

	var author_id, content string

	query := `SELECT author_id, content FROM messages WHERE id = ?;`
	
	stmt, err := Utils.DB.Prepare(query)
	defer stmt.Close()
	if err != nil {
		fmt.Println("An error occured while attemping to delete the message from the database")
	}	
	err = stmt.QueryRow(m.Message.ID).Scan(&author_id, &content)



	delete_query := `DELETE FROM messages WHERE id = ?;`

	_, err = Utils.DB.Exec(delete_query, m.Message.ID)
	if err != nil {
		fmt.Println("An error occured while attemping to delete the message from the database")
	}

	user, _ := s.User(author_id)
	avatarURL := user.AvatarURL("")
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: avatarURL,
			Name: user.GlobalName + " (@" + user.Username + ")",
		},
		Title: ":speaking_head: Message supprimé",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Auteur",
				Value: user.GlobalName,
				Inline: true,
			},
			{
				Name: "Contenu",
				Value: "`" + content + "`",
				Inline: true,
			},
			{
				Name: "Channel",
				Value: "<#" + m.ChannelID + ">",
				Inline: true,
			},
			{
				Name: "Par",
				Value: deleted_by,
				Inline: true,
			},
		},
		Color: 0x4A5B85,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	Utils.AlertChannelMessagesComplex(s, embed)

}