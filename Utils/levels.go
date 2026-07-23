package Utils

import (
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
