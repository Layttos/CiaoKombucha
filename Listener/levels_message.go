package Listener

import (
	"fmt"
	"strings"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

func LevelsMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	user_id := m.Author.ID
	current_level, current_experience := 0, 0
	query := `SELECT level, experience FROM levels WHERE user_id=?`
	err := Utils.DB.QueryRow(query, user_id).Scan(&current_level, &current_experience)
	if err != nil {
		insert_query := `INSERT INTO levels (user_id, experience, level) VALUES(?, ?, ?);`
		stmt, _ := Utils.DB.Prepare(insert_query)
		defer stmt.Close()
		_, err = stmt.Exec(user_id, 0, 0)
		if err != nil {
			return
		}
	}

	required_experience := Utils.GetRequiredExperienceForNextLevel(user_id)

	if required_experience > 0 {
		update_query := `UPDATE levels SET experience=experience+20 WHERE user_id=?;`
		stmt, _ := Utils.DB.Prepare(update_query)
		defer stmt.Close()
		_, err = stmt.Exec(user_id)
		if err != nil {
			return
		}
		if current_experience+20 >= Utils.GetRequiredExperienceForLevel(current_level+1) {
			update_query := `UPDATE levels SET level=level+1 WHERE user_id=?;`
			stmt, _ := Utils.DB.Prepare(update_query)
			defer stmt.Close()
			_, err = stmt.Exec(user_id)
			if err != nil {
				return
			}
			if strings.Compare(user_id, "386468470788980738") == 0 || strings.Compare(user_id, "550412509719298049") == 0 || strings.Compare(user_id, "584752863457050624") == 0 {
				Utils.AlertLevelsChannel(s, user_id, ":tada: Toutes mes félicitations !", fmt.Sprintf("Le dictateur <@%s> a atteint le niveau %d", user_id, current_level+1))
			} else {
				Utils.AlertLevelsChannel(s, user_id, ":tada: Félicitations !", fmt.Sprintf("<@%s> a atteint le niveau %d !", user_id, current_level+1))
			}
		}
	}
}
