package Utils

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type LeaderboardUser struct {
	UserID     string
	Experience int
	Level      int
}

func GetLeaderboard(max int, s *discordgo.Session) ([]LeaderboardUser, error) {
	query := `SELECT user_id, experience, level FROM levels ORDER BY experience DESC LIMIT ?`
	rows, err := DB.Query(query, max)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []LeaderboardUser
	for rows.Next() {
		var u LeaderboardUser
		err := rows.Scan(&u.UserID, &u.Experience, &u.Level)
		if err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return leaderboard, nil
}

func GetUserLevel(userID string) (int, int) {
	query := `SELECT experience, level FROM levels WHERE user_id = ?`
	var experience, level int
	err := DB.QueryRow(query, userID).Scan(&experience, &level)
	if err != nil {
		return 0, 0
	}
	return experience, level
}

func GetRequiredExperienceForLevel(level int) int {
	if level <= 10 {
		return level * 1000
	}
	if level <= 15 {
		delta := level - 10
		return 10000 + (delta*delta)*2000
	}
	return 60000 + (level-15)*3000
}

func GetRequiredExperienceForNextLevel(user_id string) int {
	var current_level, current_experience int
	query := `SELECT level, experience FROM levels WHERE user_id=?`
	err := DB.QueryRow(query, user_id).Scan(&current_level, &current_experience)
	if err != nil {
		return 0
	}

	return GetRequiredExperienceForLevel(current_level+1) - current_experience
}

func AddExperienceToUser(user_id string, s *discordgo.Session) error {
	_, level := GetUserLevel(user_id)
	var added_experience int
	if level < 5 {
		added_experience = 20
	} else if level < 10 {
		added_experience = 50
	} else {
		added_experience = 100
	}

	update_query := `UPDATE levels SET experience=experience+? WHERE user_id=?;`
	stmt, _ := DB.Prepare(update_query)
	defer stmt.Close()
	_, err := stmt.Exec(added_experience, user_id)
	if err != nil {
		return err
	}

	if GetRequiredExperienceForNextLevel(user_id) <= 0 {
		update_query := `UPDATE levels SET level=level+1 WHERE user_id=?;`
		_, current_level := GetUserLevel(user_id)
		stmt, _ := DB.Prepare(update_query)
		defer stmt.Close()
		_, err = stmt.Exec(user_id)
		if err != nil {
			return err
		}
		if strings.Compare(user_id, "386468470788980738") == 0 || strings.Compare(user_id, "550412509719298049") == 0 || strings.Compare(user_id, "584752863457050624") == 0 {
			AlertLevelsChannel(s, user_id, ":tada: Toutes mes félicitations !", fmt.Sprintf("Le dictateur <@%s> a atteint le niveau %d", user_id, current_level+1))
		} else {
			AlertLevelsChannel(s, user_id, ":tada: Félicitations !", fmt.Sprintf("<@%s> a atteint le niveau %d !", user_id, current_level+1))
		}
	}

	return nil
}

/*required_experience := Utils.GetRequiredExperienceForNextLevel(user_id)

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
}*/
