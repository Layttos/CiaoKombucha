package Command

import (
	"bytes"
	"fmt"
	"image"
	"net/http"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"

	"github.com/fogleman/gg"
)

type Levels struct{}

func (c *Levels) Name() string {
	return "levels"
}

func (c *Levels) Description() string {
	return "Affiche le niveau d'un utilisateur."
}

func (c *Levels) Permissions() *int64 {
	return nil
}

func (c *Levels) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	user := i.Member.User
	avatarURL := user.AvatarURL("")
	displayName := user.GlobalName
	if displayName == "" {
		displayName = user.Username
	}

	currentXP, currentLevel := Utils.GetUserLevel(user.ID)
	maxXP := Utils.GetRequiredExperienceForLevel(currentLevel + 1)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	img, err := GenerateRankCard(displayName, user.Username, avatarURL, currentLevel, currentXP, maxXP)
	if err != nil {
		fmt.Println("An error occurred while generating the rank card:", err)
		errMsg := "Une erreur est survenue lors de la génération de la carte de niveau."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &errMsg,
		})
		return false
	}

	file := &discordgo.File{
		Name:        "rank_card.png",
		ContentType: "image/png",
		Reader:      bytes.NewReader(img),
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Files: []*discordgo.File{file},
	})

	return true

}

func GenerateRankCard(displayName string, username string, avatarURL string, level int, currentXP int, maxXP int) ([]byte, error) {
	const width = 600
	const height = 200

	dc := gg.NewContext(width, height)
	if err := dc.LoadFontFace("fonts/Arial Bold.ttf", 20); err != nil {
		return nil, err
	}

	// --- 1. Fond de la carte ---
	dc.SetHexColor("#1E1F22") // Arrière-plan sombre style Discord
	dc.DrawRoundedRectangle(0, 0, width, height, 15)
	dc.Fill()

	// --- 2. Chargement & Découpage de l'Avatar ---
	avatarImg, err := DownloadImg(avatarURL)
	if err == nil {
		dc.Push() // Sauvegarde le contexte graphique (remplace Save)
		dc.NewSubPath()
		dc.DrawCircle(80, 100, 50)
		dc.Clip()
		dc.DrawImageAnchored(avatarImg, 80, 100, 0.5, 0.5)
		dc.Pop() // Restaure le contexte graphique (remplace Restore)
		dc.ResetClip()
	}

	// --- 3. Textes (Pseudo, Level, XP) ---
	dc.SetHexColor("#FFFFFF")

	nameLine := fmt.Sprintf("%s (@%s)", displayName, username)
	dc.DrawString(nameLine, 160, 80)

	// Niveau & XP
	levelStr := fmt.Sprintf("Niveau %d", level)
	xpStr := fmt.Sprintf("%d / %d XP", currentXP, maxXP)

	if err := dc.LoadFontFace("/System/Library/Fonts/Supplemental/Arial.ttf", 14); err != nil {
		return nil, err
	}
	dc.SetHexColor("#B9BBBE")
	dc.DrawString(levelStr, 160, 109)
	// Alignement à droite pour garder les chiffres lisibles même quand la barre est pleine.
	dc.DrawStringAnchored(xpStr, 540, 109, 1.0, 0.5)

	// --- 4. Barre de progression (Fond sombre) ---
	barX := 160.0
	barY := 132.0
	barWidth := 380.0
	barHeight := 14.0
	barRadius := barHeight / 2
	progress := 0.0
	if maxXP > 0 {
		progress = float64(currentXP) / float64(maxXP)
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1.0 {
		progress = 1.0
	}
	filledWidth := barWidth * progress

	dc.SetHexColor("#2B2D31")
	dc.DrawRoundedRectangle(barX, barY, barWidth, barHeight, barRadius)
	dc.Fill()

	if filledWidth > 0 {
		dc.Push()
		dc.DrawRoundedRectangle(barX, barY, barWidth, barHeight, barRadius)
		dc.Clip()
		dc.SetHexColor("#5865F2") // Couleur Discord Blurple
		dc.DrawRectangle(barX, barY, filledWidth, barHeight)
		dc.Fill()
		dc.Pop()
		dc.ResetClip()
	}

	// --- 6. Exportation PNG en mémoire ---
	var buf bytes.Buffer
	err = dc.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Fonction utilitaire pour télécharger l'avatar Discord
func DownloadImg(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}
