package Command

import (
	"bytes"
	"fmt"
	"log"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
	"github.com/fogleman/gg"
)

type Leaderboard struct{}

func (c *Leaderboard) Name() string {
	return "leaderboard"
}

func (c *Leaderboard) Description() string {
	return "Affiche le classement des utilisateurs."
}

func (c *Leaderboard) Permissions() *int64 {
	return nil
}

// LeaderboardEntry représente les données nécessaires pour un joueur du classement
type LeaderboardEntry struct {
	User  *discordgo.User
	Level int
}

func GenerateLeaderboardCard(entries []LeaderboardEntry) ([]byte, error) {
	const width = 750
	const height = 700

	dc := gg.NewContext(width, height)

	// 💡 FIX POUR LES TROUS DANS LES LETTRES :
	// Force le remplissage complet des contours de polices (Non-Zero Winding)
	dc.SetFillRule(gg.FillRuleWinding)

	// 1. Fond principal
	dc.SetHexColor("#1E1F22")
	dc.DrawRoundedRectangle(0, 0, width, height, 15)
	dc.Fill()

	// 2. Titre
	if err := dc.LoadFontFace("fonts/Roboto-Bold.ttf", 22); err != nil {
		log.Printf("Erreur police titre: %v", err)
	}
	dc.SetHexColor("#FFFFFF")
	dc.DrawStringAnchored("CLASSEMENT DU SERVEUR", 30, 45, 0.0, 0.5)

	// Ligne de séparation
	dc.SetHexColor("#2B2D31")
	dc.DrawRectangle(30, 65, width-60, 2)
	dc.Fill()

	// 3. Police pour le contenu
	if err := dc.LoadFontFace("fonts/Roboto-Bold.ttf", 16); err != nil {
		log.Printf("Erreur police texte: %v", err)
	}

	startY := 85.0
	rowHeight := 58.0

	for i := 0; i < 10; i++ {
		rowY := startY + float64(i)*rowHeight
		centerY := rowY + 24.0

		// Fond de la ligne
		dc.SetHexColor("#2B2D31")
		dc.DrawRoundedRectangle(30, rowY, width-60, 48, 8)
		dc.Fill()

		// Rang
		dc.SetHexColor("#B5BAC1")
		if i == 0 {
			dc.SetHexColor("#FEE75C")
		} else if i == 1 {
			dc.SetHexColor("#C0C0C0")
		} else if i == 2 {
			dc.SetHexColor("#CD7F32")
		}
		dc.DrawStringAnchored(fmt.Sprintf("%d.", i+1), 45, centerY, 0.0, 0.5)

		if i < len(entries) && entries[i].User != nil {
			entry := entries[i]
			user := entry.User

			// Avatar
			avatarURL := user.AvatarURL("64")
			avatarImg, err := DownloadImg(avatarURL)
			if err == nil && avatarImg != nil {
				dc.Push()
				dc.DrawCircle(110, centerY, 18)
				dc.Clip()
				dc.DrawImageAnchored(avatarImg, 110, int(centerY), 0.5, 0.5)
				dc.ResetClip()
				dc.Pop()
			}

			// Nom
			displayName := user.GlobalName
			if displayName == "" {
				displayName = user.Username
			}
			fullName := fmt.Sprintf("%s (@%s)", displayName, user.Username)

			dc.SetHexColor("#FFFFFF")
			dc.DrawStringAnchored(fullName, 145, centerY, 0.0, 0.5)

			// Niveau
			levelText := fmt.Sprintf("Niveau %d", entry.Level)
			dc.SetHexColor("#5865F2")
			dc.DrawStringAnchored(levelText, width-50, centerY, 1.0, 0.5)

		} else {
			// EMPLACEMENT LIBRE
			dc.SetHexColor("#5C5E66")
			dc.DrawStringAnchored("Emplacement libre", 145, centerY, 0.0, 0.5)
		}
	}

	var buf bytes.Buffer
	err := dc.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Leaderboard) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool {

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	userIDs, err := Utils.GetLeaderboard(10, s)
	if err != nil {
		errMsg := "Impossible de récupérer le classement."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errMsg})
		return false
	}

	var entries []LeaderboardEntry
	for _, entry := range userIDs {
		user, err := s.User(entry.UserID)
		if err != nil {
			continue
		}
		_, level := Utils.GetUserLevel(entry.UserID)
		entries = append(entries, LeaderboardEntry{User: user, Level: level})
	}

	img, err := GenerateLeaderboardCard(entries)
	if err != nil {
		errMsg := "Impossible de générer la carte du classement."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errMsg})
		return false
	}

	file := &discordgo.File{
		Name:        "leaderboard.png",
		ContentType: "image/png",
		Reader:      bytes.NewReader(img),
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Files: []*discordgo.File{file},
	})

	return true
}
