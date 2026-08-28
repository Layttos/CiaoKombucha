package Listener

import (
	"fmt"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

// ChannelCreated se déclenche à la création d'un salon (textuel, vocal, catégorie...).
func ChannelCreated(s *discordgo.Session, e *discordgo.ChannelCreate) {
	if e.Channel == nil {
		return
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionChannelCreate), e.ID)

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerCreate,
		":new: Salon créé",
		fmt.Sprintf("<#%s> (`%s`) vient d'être créé.", e.ID, e.Name),
		[]*discordgo.MessageEmbedField{
			{Name: "Type", Value: Utils.ChannelTypeLabel(e.Type), Inline: true},
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// ChannelDeleted se déclenche à la suppression d'un salon.
func ChannelDeleted(s *discordgo.Session, e *discordgo.ChannelDelete) {
	if e.Channel == nil {
		return
	}

	name := e.Name
	if e.BeforeDelete != nil && e.BeforeDelete.Name != "" {
		name = e.BeforeDelete.Name
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionChannelDelete), e.ID)

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerDelete,
		":wastebasket: Salon supprimé",
		fmt.Sprintf("Le salon `%s` vient d'être supprimé.", name),
		[]*discordgo.MessageEmbedField{
			{Name: "Type", Value: Utils.ChannelTypeLabel(e.Type), Inline: true},
			{Name: "Identifiant", Value: "`" + e.ID + "`", Inline: true},
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// ChannelUpdated se déclenche à la modification d'un salon : nom, sujet, NSFW,
// mode lent, limite d'utilisateurs, bitrate, catégorie parente.
func ChannelUpdated(s *discordgo.Session, e *discordgo.ChannelUpdate) {
	if e.Channel == nil || e.BeforeUpdate == nil {
		return
	}

	before := e.BeforeUpdate
	after := e.Channel

	var changes []string

	if before.Name != after.Name {
		changes = append(changes, fmt.Sprintf("**Nom** : `%s` → `%s`", before.Name, after.Name))
	}
	if before.Topic != after.Topic {
		changes = append(changes, fmt.Sprintf("**Sujet** : `%s` → `%s`", orNone(before.Topic), orNone(after.Topic)))
	}
	if before.NSFW != after.NSFW {
		changes = append(changes, fmt.Sprintf("**NSFW** : `%t` → `%t`", before.NSFW, after.NSFW))
	}
	if before.RateLimitPerUser != after.RateLimitPerUser {
		changes = append(changes, fmt.Sprintf("**Mode lent** : `%ds` → `%ds`", before.RateLimitPerUser, after.RateLimitPerUser))
	}
	if before.Bitrate != after.Bitrate {
		changes = append(changes, fmt.Sprintf("**Bitrate** : `%d` → `%d`", before.Bitrate, after.Bitrate))
	}
	if before.UserLimit != after.UserLimit {
		changes = append(changes, fmt.Sprintf("**Limite d'utilisateurs** : `%d` → `%d`", before.UserLimit, after.UserLimit))
	}
	if before.ParentID != after.ParentID {
		changes = append(changes, fmt.Sprintf("**Catégorie** : %s → %s", parentMention(before.ParentID), parentMention(after.ParentID)))
	}
	if len(before.PermissionOverwrites) != len(after.PermissionOverwrites) || !sameOverwrites(before.PermissionOverwrites, after.PermissionOverwrites) {
		changes = append(changes, "**Permissions du salon** modifiées")
	}

	if len(changes) == 0 {
		return
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionChannelUpdate), after.ID)

	description := fmt.Sprintf("<#%s>\n", after.ID)
	for _, c := range changes {
		description += "\n" + c
	}

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerUpdate,
		":pencil2: Salon modifié",
		description,
		[]*discordgo.MessageEmbedField{
			{Name: "Type", Value: Utils.ChannelTypeLabel(after.Type), Inline: true},
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// ThreadCreated se déclenche à la création d'un fil de discussion.
func ThreadCreated(s *discordgo.Session, e *discordgo.ThreadCreate) {
	if e.Channel == nil || !e.NewlyCreated {
		return
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionThreadCreate), e.ID)

	parent := "un salon inconnu"
	if e.ParentID != "" {
		parent = "<#" + e.ParentID + ">"
	}

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerCreate,
		":thread: Fil créé",
		fmt.Sprintf("Le fil `%s` vient d'être créé dans %s.", e.Name, parent),
		[]*discordgo.MessageEmbedField{
			{Name: "Fil", Value: "<#" + e.ID + ">", Inline: true},
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// ThreadDeleted se déclenche à la suppression d'un fil de discussion.
func ThreadDeleted(s *discordgo.Session, e *discordgo.ThreadDelete) {
	if e.Channel == nil {
		return
	}

	name := e.Name
	if e.BeforeDelete != nil && e.BeforeDelete.Name != "" {
		name = e.BeforeDelete.Name
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionThreadDelete), e.ID)

	parent := "un salon inconnu"
	if e.ParentID != "" {
		parent = "<#" + e.ParentID + ">"
	}

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerDelete,
		":wastebasket: Fil supprimé",
		fmt.Sprintf("Le fil `%s` de %s vient d'être supprimé.", name, parent),
		[]*discordgo.MessageEmbedField{
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

func orNone(v string) string {
	if v == "" {
		return "(vide)"
	}
	return v
}

func parentMention(id string) string {
	if id == "" {
		return "`aucune`"
	}
	return "<#" + id + ">"
}

// sameOverwrites compare deux jeux d'overrides de permissions salon par salon.
func sameOverwrites(a, b []*discordgo.PermissionOverwrite) bool {
	if len(a) != len(b) {
		return false
	}
	index := make(map[string]*discordgo.PermissionOverwrite, len(a))
	for _, o := range a {
		if o != nil {
			index[o.ID] = o
		}
	}
	for _, o := range b {
		if o == nil {
			continue
		}
		prev, ok := index[o.ID]
		if !ok || prev.Allow != o.Allow || prev.Deny != o.Deny || prev.Type != o.Type {
			return false
		}
	}
	return true
}
