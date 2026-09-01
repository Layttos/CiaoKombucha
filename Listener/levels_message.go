package Listener

import (
	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

func LevelsMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	user_id := m.Author.ID
	current_level, current_experience := 0, 0
	query := `SELECT level, experience FROM levels WHERE user_id=?`
	err := Utils.DB.QueryRow(query, user_id).Scan(&current_level, &current_experience)
	if err != nil {
		insert_query := `INSERT INTO levels (user_id, experience, level) VALUES(?, ?, ?);`
		stmt, err := Utils.DB.Prepare(insert_query)
		if err != nil {
			return
		}
		defer stmt.Close()
		if _, err := stmt.Exec(user_id, 0, 0); err != nil {
			return
		}
	}

	Utils.AddExperienceToUser(user_id, s)
}
