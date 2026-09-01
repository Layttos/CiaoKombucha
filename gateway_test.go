package main

import (
	"bytes"
	"log"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/bwmarrin/discordgo"
)

// setUnexportedString écrit dans un champ non exporté de la session, comme le
// fait discordgo lui-même à la réception d'un paquet READY.
func setUnexportedString(t *testing.T, session reflect.Value, name string, value string) {
	t.Helper()

	field := session.FieldByName(name)
	if !field.IsValid() {
		t.Fatalf("le champ %s a disparu de discordgo.Session", name)
	}
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().SetString(value)
}

// TestResetGatewayResumeState verrouille l'accès aux champs internes de
// discordgo : si une mise à jour de la bibliothèque les renomme, ce test casse
// avant que le bot ne se retrouve à nouveau bloqué hors ligne en production.
func TestResetGatewayResumeState(t *testing.T) {
	s, err := discordgo.New("Bot jeton-de-test")
	if err != nil {
		t.Fatal(err)
	}

	session := reflect.ValueOf(s).Elem()
	setUnexportedString(t, session, "sessionID", "abcdef0123456789")
	setUnexportedString(t, session, "resumeGatewayURL", "wss://url-de-reprise-perimee.discord.gg")

	sequenceField := session.FieldByName("sequence")
	sequence, ok := reflect.NewAt(sequenceField.Type(), unsafe.Pointer(sequenceField.UnsafeAddr())).Elem().Interface().(*int64)
	if !ok || sequence == nil {
		t.Fatal("le champ sequence n'est plus un *int64 dans discordgo.Session")
	}
	atomic.StoreInt64(sequence, 42)

	if !resetGatewayResumeState(s) {
		t.Fatal("resetGatewayResumeState a échoué : les champs internes de discordgo ont changé")
	}

	if got := session.FieldByName("sessionID").String(); got != "" {
		t.Errorf("sessionID vaut %q, attendu vide", got)
	}
	if got := session.FieldByName("resumeGatewayURL").String(); got != "" {
		t.Errorf("resumeGatewayURL vaut %q, attendu vide", got)
	}
	if got := atomic.LoadInt64(sequence); got != 0 {
		t.Errorf("sequence vaut %d, attendu 0", got)
	}
}

// TestSafeHandlerContientLesPanics vérifie qu'un listener défaillant ne remonte
// plus jusqu'à la goroutine de discordgo, qui n'a aucun recover et emporterait
// le processus entier.
func TestSafeHandlerContientLesPanics(t *testing.T) {
	appele := false

	handler := safeHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		appele = true
		var user *discordgo.User
		_ = user.AvatarURL("") // déréférencement nil volontaire
	})

	handler(nil, &discordgo.MessageCreate{})

	if !appele {
		t.Fatal("le handler enveloppé n'a pas été appelé")
	}
}

// TestSafeHandlerResteReconnuParDiscordgo vérifie que l'enveloppe générique
// produit toujours une signature que le type switch de discordgo sait router.
// Sinon, AddHandler se contenterait d'écrire « Invalid handler type » dans les
// logs et *tous* les listeners du bot deviendraient muets, sans autre symptôme.
func TestSafeHandlerResteReconnuParDiscordgo(t *testing.T) {
	s, err := discordgo.New("Bot jeton-de-test")
	if err != nil {
		t.Fatal(err)
	}

	var journal bytes.Buffer
	precedent := log.Writer()
	log.SetOutput(&journal)
	defer log.SetOutput(precedent)

	s.AddHandler(safeHandler(func(*discordgo.Session, *discordgo.MessageCreate) {}))
	s.AddHandler(safeHandler(func(*discordgo.Session, *discordgo.MessageUpdate) {}))
	s.AddHandler(safeHandler(func(*discordgo.Session, *discordgo.GuildMemberUpdate) {}))
	s.AddHandler(safeHandler(func(*discordgo.Session, *discordgo.InteractionCreate) {}))
	s.AddHandler(safeHandler(func(*discordgo.Session, *discordgo.VoiceStateUpdate) {}))
	s.AddHandler(safeHandler(func(*discordgo.Session, *discordgo.VoiceServerUpdate) {}))

	if strings.Contains(journal.String(), "Invalid handler type") {
		t.Fatalf("discordgo ne reconnaît plus la signature enveloppée :\n%s", journal.String())
	}
}
