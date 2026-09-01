package Utils

import (
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

func AlertChannelMembers(s *discordgo.Session, memberID string, title string, content string) {
	s.ChannelMessageSendEmbed(os.Getenv("USER_CHANNEL_ID"), &discordgo.MessageEmbed{
		Author:      ActorEmbedAuthor(s, memberID),
		Title:       title,
		Color:       0xC7A49D,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func AlertLevelsChannel(s *discordgo.Session, memberID string, title string, content string) {
	s.ChannelMessageSendEmbed(os.Getenv("LEVELS_CHANNEL_ID"), &discordgo.MessageEmbed{
		Author:      ActorEmbedAuthor(s, memberID),
		Title:       title,
		Color:       0x2F75A1,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func AlertChannelLevelsComplex(s *discordgo.Session, embed *discordgo.MessageEmbed) {
	s.ChannelMessageSendEmbed(os.Getenv("LEVELS_CHANNEL_ID"), embed)
}

func AlertChannelMembersComplex(s *discordgo.Session, embed *discordgo.MessageEmbed) {
	s.ChannelMessageSendEmbed(os.Getenv("USER_CHANNEL_ID"), embed)
}

func AlertChannelMessages(s *discordgo.Session, memberID string, title string, content string) {
	s.ChannelMessageSendEmbed(os.Getenv("MESSAGES_CHANNEL_ID"), &discordgo.MessageEmbed{
		Author:      ActorEmbedAuthor(s, memberID),
		Title:       title,
		Color:       0x4A5B85,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func AlertChannelMessagesComplex(s *discordgo.Session, embed *discordgo.MessageEmbed) {
	s.ChannelMessageSendEmbed(os.Getenv("MESSAGES_CHANNEL_ID"), embed)
}

func AlertChannelModeration(s *discordgo.Session, memberID string, title string, content string) {
	s.ChannelMessageSendEmbed(os.Getenv("MODERATION_CHANNEL_ID"), &discordgo.MessageEmbed{
		Author:      ActorEmbedAuthor(s, memberID),
		Title:       title,
		Color:       0x7B53A3,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func AlertChannelModerationComplex(s *discordgo.Session, embed *discordgo.MessageEmbed) {
	s.ChannelMessageSendEmbed(os.Getenv("MODERATION_CHANNEL_ID"), embed)
}

// Couleurs des embeds, différenciées selon le type d'action.
const (
	ColorServerCreate      = 0x57F287 // vert    — création (rôle, salon, fil, invitation...)
	ColorServerDelete      = 0xED4245 // rouge   — suppression
	ColorServerUpdate      = 0xFEE75C // jaune   — modification
	ColorServerGuild       = 0x5865F2 // blurple — paramètres du serveur
	ColorServerExpression  = 0xEB459E // fuchsia — emojis / stickers
	ColorServerWebhook     = 0x9B59B6 // violet  — webhooks
	ColorModerationTimeout = 0xE67E22 // orange  — exclusion temporaire
	ColorModerationLift    = 0x57F287 // vert    — levée d'une sanction
)

// AlertChannelServer construit un embed standardisé pour les logs de structure du
// serveur (rôles, salons, paramètres, webhooks, invitations, emojis...) et
// l'envoie dans le channel de modération. actorID est l'utilisateur à l'origine
// de l'action : il est utilisé comme auteur de l'embed (avatar + pseudo).
func AlertChannelServer(s *discordgo.Session, actorID string, color int, title string, description string, fields []*discordgo.MessageEmbedField) {
	embed := &discordgo.MessageEmbed{
		Author:      ActorEmbedAuthor(s, actorID),
		Title:       title,
		Description: description,
		Fields:      fields,
		Color:       color,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	s.ChannelMessageSendEmbed(os.Getenv("MODERATION_CHANNEL_ID"), embed)
}

// ActorEmbedAuthor renvoie un bloc auteur d'embed pour l'utilisateur donné, ou
// nil si l'ID est vide ou l'utilisateur introuvable.
func ActorEmbedAuthor(s *discordgo.Session, actorID string) *discordgo.MessageEmbedAuthor {
	if actorID == "" {
		return nil
	}
	user, err := s.User(actorID)
	if err != nil || user == nil {
		return nil
	}
	name := "@" + user.Username
	if user.GlobalName != "" {
		name = user.GlobalName + " (@" + user.Username + ")"
	}
	return &discordgo.MessageEmbedAuthor{
		IconURL: user.AvatarURL(""),
		Name:    name,
	}
}

// ActorMention renvoie une mention exploitable pour un champ d'embed, ou une
// valeur par défaut si l'ID est vide.
func ActorMention(actorID string) string {
	if actorID == "" {
		return "Non spécifié(e)"
	}
	return "<@" + actorID + ">"
}

// ReasonOrDefault normalise une raison d'audit log.
func ReasonOrDefault(reason string) string {
	if reason == "" {
		return "Non spécifiée"
	}
	return reason
}

// LatestAuditEntry renvoie l'entrée d'audit log la plus récente pour le type
// d'action donné. Si targetID est renseigné, seule une entrée ciblant cet ID est
// retournée. Les entrées de plus de 10 secondes sont ignorées afin de ne pas
// attribuer une action passée sans rapport avec l'évènement courant.
func LatestAuditEntry(s *discordgo.Session, guildID string, action int, targetID string) *discordgo.AuditLogEntry {
	auditLog, err := s.GuildAuditLog(guildID, "", "", action, 5)
	if err != nil || auditLog == nil {
		return nil
	}

	for _, entry := range auditLog.AuditLogEntries {
		if entry == nil {
			continue
		}
		if targetID != "" && entry.TargetID != targetID {
			continue
		}
		if ts, err := discordgo.SnowflakeTimestamp(entry.ID); err == nil {
			if time.Since(ts) > 10*time.Second {
				continue
			}
		}
		return entry
	}

	return nil
}

// ResolveAuditActor renvoie l'ID de l'utilisateur ayant effectué la dernière
// action du type donné ainsi que la raison éventuellement fournie ("" si absente).
func ResolveAuditActor(s *discordgo.Session, guildID string, action int, targetID string) (actorID string, reason string) {
	entry := LatestAuditEntry(s, guildID, action, targetID)
	if entry == nil {
		return "", ""
	}
	return entry.UserID, entry.Reason
}

// ResolveAuditActorAny teste plusieurs types d'action et retient l'entrée la plus
// récente. Utile quand un seul évènement gateway (emojis, webhooks) peut
// correspondre à une création, une modification ou une suppression.
func ResolveAuditActorAny(s *discordgo.Session, guildID string, targetID string, actions ...int) (actorID string, reason string) {
	var newest *discordgo.AuditLogEntry
	var newestAt time.Time
	for _, action := range actions {
		entry := LatestAuditEntry(s, guildID, action, targetID)
		if entry == nil {
			continue
		}
		ts, err := discordgo.SnowflakeTimestamp(entry.ID)
		if err != nil {
			continue
		}
		if newest == nil || ts.After(newestAt) {
			newest, newestAt = entry, ts
		}
	}
	if newest == nil {
		return "", ""
	}
	return newest.UserID, newest.Reason
}

// auditChangeLabels traduit les clés d'audit log Discord les plus courantes.
var auditChangeLabels = map[string]string{
	"name":                          "Nom",
	"icon_hash":                     "Icône",
	"splash_hash":                   "Image d'invitation",
	"discovery_splash_hash":         "Image de découverte",
	"banner_hash":                   "Bannière",
	"owner_id":                      "Propriétaire",
	"region":                        "Région vocale",
	"afk_channel_id":                "Salon AFK",
	"afk_timeout":                   "Délai AFK",
	"rules_channel_id":              "Salon des règles",
	"public_updates_channel_id":     "Salon des annonces communautaires",
	"system_channel_id":             "Salon système",
	"widget_enabled":                "Widget activé",
	"verification_level":            "Niveau de vérification",
	"default_message_notifications": "Notifications par défaut",
	"explicit_content_filter":       "Filtre de contenu explicite",
	"mfa_level":                     "Double authentification requise",
	"vanity_url_code":               "URL personnalisée",
	"preferred_locale":              "Langue principale",
	"description":                   "Description",
	"premium_progress_bar_enabled":  "Barre de progression des boosts",
}

// FormatAuditChanges met en forme la liste des changements d'une entrée d'audit
// log sous forme de lignes lisibles « **Clé** : `avant` → `après` ».
func FormatAuditChanges(entry *discordgo.AuditLogEntry) []string {
	if entry == nil {
		return nil
	}
	var lines []string
	for _, change := range entry.Changes {
		if change == nil || change.Key == nil {
			continue
		}
		key := string(*change.Key)
		label, ok := auditChangeLabels[key]
		if !ok {
			label = key
		}
		before := stringifyChange(change.OldValue)
		after := stringifyChange(change.NewValue)
		if before == after {
			continue
		}
		lines = append(lines, fmt.Sprintf("**%s** : `%s` → `%s`", label, before, after))
	}
	return lines
}

func stringifyChange(v interface{}) string {
	if v == nil {
		return "(vide)"
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return "(vide)"
		}
		return t
	case bool:
		return fmt.Sprintf("%t", t)
	case float64:
		return fmt.Sprintf("%g", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// ChannelTypeLabel renvoie un libellé lisible pour un type de salon.
func ChannelTypeLabel(t discordgo.ChannelType) string {
	switch t {
	case discordgo.ChannelTypeGuildText:
		return "Salon textuel"
	case discordgo.ChannelTypeGuildVoice:
		return "Salon vocal"
	case discordgo.ChannelTypeGuildCategory:
		return "Catégorie"
	case discordgo.ChannelTypeGuildNews:
		return "Salon d'annonces"
	case discordgo.ChannelTypeGuildNewsThread, discordgo.ChannelTypeGuildPublicThread, discordgo.ChannelTypeGuildPrivateThread:
		return "Fil de discussion"
	case discordgo.ChannelTypeGuildStageVoice:
		return "Salon de conférence"
	case discordgo.ChannelTypeGuildForum:
		return "Forum"
	default:
		return "Salon"
	}
}
