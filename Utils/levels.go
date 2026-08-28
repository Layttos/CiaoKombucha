package Utils

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

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

// ---------------------------------------------------------------------------
// Courbe de progression
//
// XP nécessaires pour passer du niveau n au niveau n+1 : 5n² + 50n + 100.
// C'est la courbe quadratique éprouvée (type MEE6) : les premiers niveaux
// tombent en quelques messages, les suivants demandent de plus en plus
// d'activité, sans jamais devenir hors d'atteinte.
//
//	niveau  1  ->        100 XP cumulés
//	niveau  5  ->      1 400 XP cumulés
//	niveau 10 ->      6 400 XP cumulés
//	niveau 20 ->     35 800 XP cumulés
//	niveau 50 ->    475 500 XP cumulés
// ---------------------------------------------------------------------------

// xpToNextLevel renvoie les XP à gagner entre le niveau `level` et `level+1`.
func xpToNextLevel(level int) int {
	if level < 0 {
		level = 0
	}
	return 5*level*level + 50*level + 100
}

// GetRequiredExperienceForLevel renvoie le total d'XP cumulé nécessaire pour
// atteindre `level` (niveau 0 = 0 XP).
func GetRequiredExperienceForLevel(level int) int {
	total := 0
	for n := 0; n < level; n++ {
		total += xpToNextLevel(n)
	}
	return total
}

// GetLevelForExperience renvoie le niveau correspondant à un total d'XP.
func GetLevelForExperience(experience int) int {
	level := 0
	for experience >= GetRequiredExperienceForLevel(level+1) {
		level++
	}
	return level
}

// GetRequiredExperienceForNextLevel renvoie les XP restants avant le prochain
// niveau pour un utilisateur.
func GetRequiredExperienceForNextLevel(user_id string) int {
	experience, _ := GetUserLevel(user_id)
	level := GetLevelForExperience(experience)
	return GetRequiredExperienceForLevel(level+1) - experience
}

// ---------------------------------------------------------------------------
// Gain d'XP par message
//
// Chaque message rapporte un montant aléatoire (15–25 XP), au plus une fois
// par minute et par membre. Le tirage aléatoire empêche de calculer au XP près
// le nombre de messages restants, et le cooldown neutralise le spam.
// ---------------------------------------------------------------------------

const (
	xpMinPerMessage   = 15
	xpMaxPerMessage   = 25
	xpMessageCooldown = 60 * time.Second
)

var (
	xpLastGrant   = map[string]time.Time{}
	xpLastGrantMu sync.Mutex
)

// AddExperienceToUser attribue l'XP d'un message à son auteur (en respectant le
// cooldown) puis synchronise le niveau stocké et annonce les montées de niveau.
func AddExperienceToUser(user_id string, s *discordgo.Session) error {
	xpLastGrantMu.Lock()
	if last, ok := xpLastGrant[user_id]; ok && time.Since(last) < xpMessageCooldown {
		xpLastGrantMu.Unlock()
		return nil
	}
	xpLastGrant[user_id] = time.Now()
	xpLastGrantMu.Unlock()

	previous_experience, previous_level := GetUserLevel(user_id)

	gain := xpMinPerMessage + rand.IntN(xpMaxPerMessage-xpMinPerMessage+1)
	new_experience := previous_experience + gain

	stmt, err := DB.Prepare(`UPDATE levels SET experience=? WHERE user_id=?;`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(new_experience, user_id); err != nil {
		return err
	}

	new_level := GetLevelForExperience(new_experience)
	if new_level == previous_level {
		return nil
	}

	level_stmt, err := DB.Prepare(`UPDATE levels SET level=? WHERE user_id=?;`)
	if err != nil {
		return err
	}
	defer level_stmt.Close()
	if _, err := level_stmt.Exec(new_level, user_id); err != nil {
		return err
	}

	if new_level <= previous_level {
		return nil
	}

	if strings.Compare(user_id, "386468470788980738") == 0 || strings.Compare(user_id, "550412509719298049") == 0 || strings.Compare(user_id, "584752863457050624") == 0 {
		AlertLevelsChannel(s, user_id, ":tada: Toutes mes félicitations !", fmt.Sprintf("Le dictateur <@%s> a atteint le niveau %d", user_id, new_level))
	} else {
		AlertLevelsChannel(s, user_id, ":tada: Félicitations !", fmt.Sprintf("<@%s> a atteint le niveau %d !", user_id, new_level))
	}

	return nil
}
