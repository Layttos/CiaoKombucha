package Listener

import (
	"fmt"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

// RoleCreated se déclenche lorsqu'un rôle est créé sur le serveur.
func RoleCreated(s *discordgo.Session, e *discordgo.GuildRoleCreate) {
	if e.Role == nil {
		return
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionRoleCreate), e.Role.ID)

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerCreate,
		":new: Rôle créé",
		fmt.Sprintf("Le rôle `%s` vient d'être créé.", e.Role.Name),
		[]*discordgo.MessageEmbedField{
			{Name: "Rôle", Value: "<@&" + e.Role.ID + ">", Inline: true},
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// RoleDeleted se déclenche lorsqu'un rôle est supprimé du serveur.
func RoleDeleted(s *discordgo.Session, e *discordgo.GuildRoleDelete) {
	name := "Rôle inconnu"
	if e.BeforeDelete != nil {
		name = e.BeforeDelete.Name
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionRoleDelete), e.RoleID)

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerDelete,
		":wastebasket: Rôle supprimé",
		fmt.Sprintf("Le rôle `%s` vient d'être supprimé.", name),
		[]*discordgo.MessageEmbedField{
			{Name: "Identifiant", Value: "`" + e.RoleID + "`", Inline: true},
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// RoleUpdated se déclenche lorsqu'un rôle existant est modifié (nom, couleur,
// permissions, affichage séparé, mention autorisée...).
func RoleUpdated(s *discordgo.Session, e *discordgo.GuildRoleUpdate) {
	if e.Role == nil || e.BeforeUpdate == nil {
		return
	}

	before := e.BeforeUpdate
	after := e.Role

	var changes []string

	if before.Name != after.Name {
		changes = append(changes, fmt.Sprintf("**Nom** : `%s` → `%s`", before.Name, after.Name))
	}
	if before.Color != after.Color {
		changes = append(changes, fmt.Sprintf("**Couleur** : `#%06X` → `#%06X`", before.Color, after.Color))
	}
	if before.Hoist != after.Hoist {
		changes = append(changes, fmt.Sprintf("**Affiché séparément** : `%t` → `%t`", before.Hoist, after.Hoist))
	}
	if before.Mentionable != after.Mentionable {
		changes = append(changes, fmt.Sprintf("**Mentionnable** : `%t` → `%t`", before.Mentionable, after.Mentionable))
	}
	if before.Permissions != after.Permissions {
		added, removed := diffPermissions(before.Permissions, after.Permissions)
		if added != "" {
			changes = append(changes, "**Permissions ajoutées** : "+added)
		}
		if removed != "" {
			changes = append(changes, "**Permissions retirées** : "+removed)
		}
	}
	// La position n'est volontairement pas suivie : réorganiser les rôles émet un
	// évènement par rôle décalé et provoquerait un flot de logs sans intérêt.

	if len(changes) == 0 {
		return
	}

	actorID, _ := Utils.ResolveAuditActor(s, e.GuildID, int(discordgo.AuditLogActionRoleUpdate), after.ID)

	description := "<@&" + after.ID + ">\n"
	for _, c := range changes {
		description += "\n" + c
	}

	Utils.AlertChannelServer(s, actorID, Utils.ColorServerUpdate,
		":pencil2: Rôle modifié",
		description,
		[]*discordgo.MessageEmbedField{
			{Name: "Par", Value: Utils.ActorMention(actorID), Inline: true},
		},
	)
}

// permissionNames associe chaque bit de permission Discord à un libellé lisible.
var permissionNames = []struct {
	bit  int64
	name string
}{
	{discordgo.PermissionAdministrator, "Administrateur"},
	{discordgo.PermissionManageServer, "Gérer le serveur"},
	{discordgo.PermissionManageRoles, "Gérer les rôles"},
	{discordgo.PermissionManageChannels, "Gérer les salons"},
	{discordgo.PermissionManageWebhooks, "Gérer les webhooks"},
	{discordgo.PermissionManageEmojis, "Gérer les emojis et stickers"},
	{discordgo.PermissionManageMessages, "Gérer les messages"},
	{discordgo.PermissionManageThreads, "Gérer les fils"},
	{discordgo.PermissionManageEvents, "Gérer les évènements"},
	{discordgo.PermissionViewAuditLogs, "Voir les logs d'audit"},
	{discordgo.PermissionKickMembers, "Expulser des membres"},
	{discordgo.PermissionBanMembers, "Bannir des membres"},
	{discordgo.PermissionModerateMembers, "Exclure temporairement des membres"},
	{discordgo.PermissionMentionEveryone, "Mentionner @everyone"},
	{discordgo.PermissionManageNicknames, "Gérer les pseudos"},
	{discordgo.PermissionChangeNickname, "Changer de pseudo"},
	{discordgo.PermissionVoiceMoveMembers, "Déplacer des membres"},
	{discordgo.PermissionVoiceMuteMembers, "Rendre muet en vocal"},
	{discordgo.PermissionVoiceDeafenMembers, "Mettre en sourdine en vocal"},
	{discordgo.PermissionVoicePrioritySpeaker, "Priorité d'orateur"},
}

// diffPermissions renvoie deux listes lisibles : les permissions ajoutées et
// celles retirées entre deux masques de bits.
func diffPermissions(before, after int64) (added string, removed string) {
	for _, p := range permissionNames {
		hadBefore := before&p.bit == p.bit
		hasAfter := after&p.bit == p.bit
		if !hadBefore && hasAfter {
			if added != "" {
				added += ", "
			}
			added += p.name
		}
		if hadBefore && !hasAfter {
			if removed != "" {
				removed += ", "
			}
			removed += p.name
		}
	}
	return
}
