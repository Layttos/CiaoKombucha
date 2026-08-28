package Listener

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

// GuildUpdated se déclenche à la modification des paramètres du serveur
// (nom, icône, salon AFK, niveau de vérification, salon système...).
func GuildUpdated(s *discordgo.Session, e *discordgo.GuildUpdate) {
	if e.Guild == nil {
		return
	}

	entry := Utils.LatestAuditEntry(s, e.ID, int(discordgo.AuditLogActionGuildUpdate), e.ID)
	changes := Utils.FormatAuditChanges(entry)
	if len(changes) == 0 {
		return
	}

	actorID := ""
	if entry != nil {
		actorID = entry.UserID
	}

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerGuild,
		":gear: Paramètres du serveur modifiés",
		strings.Join(changes, "\n"),
		[]*discordgo.MessageEmbedField{
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// emojiCache conserve le dernier état connu des emojis par serveur afin de
// détecter les ajouts, suppressions et renommages.
var (
	emojiCache   = map[string]map[string]string{}
	emojiCacheMu sync.Mutex
)

// GuildEmojisUpdated se déclenche à chaque modification de la liste des emojis.
func GuildEmojisUpdated(s *discordgo.Session, e *discordgo.GuildEmojisUpdate) {
	current := map[string]string{}
	for _, emoji := range e.Emojis {
		if emoji != nil {
			current[emoji.ID] = emoji.Name
		}
	}

	emojiCacheMu.Lock()
	previous, seen := emojiCache[e.GuildID]
	emojiCache[e.GuildID] = current
	emojiCacheMu.Unlock()

	if !seen {
		// Premier évènement reçu après le démarrage : on initialise sans logguer.
		return
	}

	var added, removed, renamed []string
	for id, name := range current {
		if oldName, ok := previous[id]; !ok {
			added = append(added, "`:"+name+":`")
		} else if oldName != name {
			renamed = append(renamed, "`:"+oldName+":` → `:"+name+":`")
		}
	}
	for id, name := range previous {
		if _, ok := current[id]; !ok {
			removed = append(removed, "`:"+name+":`")
		}
	}

	if len(added) == 0 && len(removed) == 0 && len(renamed) == 0 {
		return
	}

	actorID, _ := Utils.ResolveAuditActorAny(s, e.GuildID, "",
		int(discordgo.AuditLogActionEmojiCreate),
		int(discordgo.AuditLogActionEmojiUpdate),
		int(discordgo.AuditLogActionEmojiDelete),
	)

	fields := []*discordgo.MessageEmbedField{
		{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
	}
	if len(added) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Ajouté(s)", Value: strings.Join(added, ", "), Inline: false})
	}
	if len(removed) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Supprimé(s)", Value: strings.Join(removed, ", "), Inline: false})
	}
	if len(renamed) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Renommé(s)", Value: strings.Join(renamed, ", "), Inline: false})
	}

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerExpression, ":smiley: Emojis du serveur modifiés", "", fields)
}

// MemberUnbanned se déclenche lorsqu'un membre est débanni. Le log est envoyé
// dans le même channel que les bannissements (USER_CHANNEL_ID).
func MemberUnbanned(s *discordgo.Session, e *discordgo.GuildBanRemove) {
	time.Sleep(500 * time.Millisecond)

	actorID, reason := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionMemberBanRemove), e.User.ID)

	user, err := s.User(e.User.ID)
	if err != nil || user == nil {
		user = e.User
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: user.AvatarURL(""),
			Name:    user.GlobalName + " (@" + user.Username + ")",
		},
		Title: ":unlock: Membre débanni(e)",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Modérateur", Value: Utils.ActorMention(actorID), Inline: true},
			{Name: "Débanni(e)", Value: user.Mention(), Inline: true},
			{Name: "Raison", Value: Utils.ReasonOrDefault(reason), Inline: true},
		},
		Color: Utils.ColorModerationLift,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Layttos Industries© - Tous droits réservés.",
			IconURL: "https://cdn.discordapp.com/avatars/727939986175033346/3ef68283b237e83f6cb4b6815b96ab0f.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	Utils.AlertChannelMembersComplex(s, embed)
}

// WebhooksUpdated se déclenche lorsqu'un webhook d'un salon est créé, modifié
// ou supprimé.
func WebhooksUpdated(s *discordgo.Session, e *discordgo.WebhooksUpdate) {
	title := ":robot: Webhook modifié"
	action := int(discordgo.AuditLogActionWebhookUpdate)
	color := Utils.ColorServerUpdate

	if entry := Utils.LatestAuditEntry(s, e.GuildID, int(discordgo.AuditLogActionWebhookCreate), ""); entry != nil {
		title, action, color = ":robot: Webhook créé", int(discordgo.AuditLogActionWebhookCreate), Utils.ColorServerCreate
	} else if entry := Utils.LatestAuditEntry(s, e.GuildID, int(discordgo.AuditLogActionWebhookDelete), ""); entry != nil {
		title, action, color = ":robot: Webhook supprimé", int(discordgo.AuditLogActionWebhookDelete), Utils.ColorServerDelete
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, action, "")

	Utils.AlertChannelServer(s, actorID, color,
		title,
		fmt.Sprintf("Un webhook du salon <#%s> a été modifié.", e.ChannelID),
		[]*discordgo.MessageEmbedField{
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// InviteCreated se déclenche à la création d'une invitation.
func InviteCreated(s *discordgo.Session, e *discordgo.InviteCreate) {
	actorID := ""
	if e.Inviter != nil {
		actorID = e.Inviter.ID
	}

	maxUses := "illimité"
	if e.MaxUses > 0 {
		maxUses = fmt.Sprintf("%d", e.MaxUses)
	}

	expiration := "jamais"
	if e.MaxAge > 0 {
		expiration = (time.Duration(e.MaxAge) * time.Second).String()
	}

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerCreate,
		":envelope_with_arrow: Invitation créée",
		fmt.Sprintf("Code `%s` pour le salon <#%s>.", e.Code, e.ChannelID),
		[]*discordgo.MessageEmbedField{
			{Name: "Créée par", Value: Utils.ActorMention(actorID), Inline: true},
			{Name: "Utilisations max", Value: maxUses, Inline: true},
			{Name: "Expire dans", Value: expiration, Inline: true},
			{Name: "Temporaire", Value: fmt.Sprintf("%t", e.Temporary), Inline: true},
		},
	)
}

// InviteDeleted se déclenche à la suppression (ou expiration) d'une invitation.
func InviteDeleted(s *discordgo.Session, e *discordgo.InviteDelete) {
	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionInviteDelete), "")

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerDelete,
		":wastebasket: Invitation supprimée",
		fmt.Sprintf("Le code `%s` du salon <#%s> n'est plus valide.", e.Code, e.ChannelID),
		[]*discordgo.MessageEmbedField{
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}
