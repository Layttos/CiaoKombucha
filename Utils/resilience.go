package Utils

import (
	"log"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"
)

// gatewayHeartbeatGrace : Discord attend un heartbeat toutes les ~41 secondes et
// discordgo remet LastHeartbeatAck à jour à chaque accusé de réception. Passé
// deux minutes sans ACK, la connexion websocket est morte — discordgo l'a même
// déjà fermée et est en train de retenter.
const gatewayHeartbeatGrace = 2 * time.Minute

// GatewayConnected indique si la session possède une connexion gateway vivante.
//
// À appeler avant tout appel qui écrit directement sur le websocket
// (ChannelVoiceJoinManual en particulier) : discordgo n'y vérifie pas que la
// connexion existe toujours et déréférence un pointeur nil quand elle est
// tombée, ce qui fait paniquer — donc mourir — le processus entier.
func GatewayConnected(s *discordgo.Session) bool {
	if s == nil {
		return false
	}

	s.RLock()
	lastAck := s.LastHeartbeatAck
	s.RUnlock()

	return !lastAck.IsZero() && time.Since(lastAck) < gatewayHeartbeatGrace
}

// RecoverPanic rattrape un panic et le journalise au lieu de laisser le
// processus mourir. À placer en defer en tête de goroutine ou de handler :
// discordgo exécute chaque listener dans sa propre goroutine sans recover, donc
// le moindre déréférencement nil dans un listener emporte tout le bot.
func RecoverPanic(context string) {
	if r := recover(); r != nil {
		log.Printf("Panic rattrapé dans %s : %v\n%s", context, r, debug.Stack())
	}
}

// SafeGo lance fn dans une goroutine protégée par RecoverPanic.
func SafeGo(context string, fn func()) {
	go func() {
		defer RecoverPanic(context)
		fn()
	}()
}
