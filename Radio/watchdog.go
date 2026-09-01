package Radio

import (
	"fmt"
	"os"
	"sync"
	"time"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

const (
	// Fréquence d'inspection de la lecture en cours.
	watchdogInterval = 30 * time.Second

	// Lavalink envoie l'état du lecteur toutes les cinq secondes environ. Passé
	// ce délai sans nouvelle, c'est le nœud lui-même qui ne répond plus.
	playerUpdateGrace = time.Minute

	// Nombre de relevés consécutifs avec une position figée avant de relancer.
	// Deux relevés espacés de 30 s évitent de réagir à un simple à-coup réseau.
	stalledChecksBeforeRestart = 2

	// Fenêtre pendant laquelle les demandes de passage au morceau suivant sont
	// fusionnées : Lavalink signale souvent un même incident par deux
	// évènements (exception puis fin de morceau), et sans ce filtre le bot
	// sauterait deux morceaux d'un coup.
	advanceDebounce = 3 * time.Second

	// Même principe pour la reconnexion vocale, en plus large : elle est
	// nettement plus coûteuse et Discord peut fermer la connexion en rafale.
	voiceRecoveryDebounce = 15 * time.Second
)

var (
	recoveryMu        sync.Mutex
	lastAdvance       time.Time
	lastVoiceRecovery time.Time
)

// LavalinkEventHandler construit le listener passé à disgolink.
//
// L'ancienne version ne réagissait qu'à une fin de morceau normale
// (TrackEndReasonFinished). Tous les autres évènements étaient ignorés : flux
// illisible, morceau bloqué, connexion vocale fermée par Discord. Le bot restait
// alors dans le salon, se croyait en train de jouer, et plus aucun son ne
// sortait — sans la moindre ligne de log pour l'expliquer.
func LavalinkEventHandler(s *discordgo.Session) func(disgolink.Player, lavalink.Event) {
	return func(player disgolink.Player, event lavalink.Event) {
		switch e := event.(type) {
		case lavalink.TrackEndEvent:
			// MayStartNext() couvre "finished" mais aussi "loadFailed" : un flux
			// Navidrome momentanément illisible ne doit pas arrêter la radio.
			// "replaced", "stopped" et "cleanup" sont au contraire des arrêts
			// voulus — enchaîner dessus provoquerait une boucle, puisque chaque
			// changement de morceau termine le précédent en "replaced".
			if e.Reason.MayStartNext() {
				requestAdvance(s, "fin de morceau ("+string(e.Reason)+")")
			}

		case lavalink.TrackStuckEvent:
			// Lavalink signale que le pourvoyeur de trames audio est bloqué mais
			// ne termine pas le morceau : sans ça, le bot reste muet
			// indéfiniment.
			requestAdvance(s, "morceau bloqué depuis "+e.Threshold.String())

		case lavalink.TrackExceptionEvent:
			requestAdvance(s, "erreur de lecture ("+e.Exception.Error()+")")

		case lavalink.WebSocketClosedEvent:
			// Discord a fermé la connexion vocale : 4006 session invalide, 4009
			// session expirée, 4014 déconnecté, 4015 serveur vocal tombé.
			// Lavalink continue de « jouer » dans le vide tant qu'on ne refait
			// pas la connexion.
			requestVoiceRecovery(s, fmt.Sprintf("connexion vocale fermée par Discord (code %d, %s)", e.Code, e.Reason))
		}
	}
}

// StartWatchdog surveille la lecture et la relance quand le son s'arrête sans
// que Lavalink n'émette le moindre évènement : nœud redémarré, connexion vocale
// morte côté Discord, flux figé. C'est le filet pour tout ce que
// LavalinkEventHandler ne peut pas voir passer.
func StartWatchdog(s *discordgo.Session) {
	Utils.SafeGo("le chien de garde de la radio", func() {
		ticker := time.NewTicker(watchdogInterval)
		defer ticker.Stop()

		var lastPosition lavalink.Duration
		var stalled int

		for range ticker.C {
			player := radioPlayer(s)
			if player == nil {
				stalled = 0
				continue
			}

			// Attention : Player.Position() extrapole à partir de l'horloge
			// locale et continue donc d'avancer même quand plus rien n'arrive de
			// Lavalink. C'est la position brute du dernier état reçu qui dit la
			// vérité sur ce qui sort réellement.
			state := player.State()

			switch {
			case !state.Connected:
				stalled = 0
				requestVoiceRecovery(s, "Lavalink n'est plus connecté au serveur vocal")

			case state.Time.IsZero() || time.Since(state.Time.Time) > playerUpdateGrace:
				stalled = 0
				requestVoiceRecovery(s, "plus aucun état de lecture reçu de Lavalink")

			case player.Track() == nil:
				stalled = 0
				requestAdvance(s, "plus aucun morceau chargé")

			case player.Paused():
				stalled = 0
				requestAdvance(s, "lecteur en pause")

			case state.Position != lastPosition:
				lastPosition = state.Position
				stalled = 0

			default:
				stalled++
				if stalled >= stalledChecksBeforeRestart {
					stalled = 0
					requestAdvance(s, "position de lecture figée")
				}
			}
		}
	})
}

// radioPlayer renvoie le lecteur de la guilde, ou nil quand il n'y a rien à
// surveiller : radio non initialisée, gateway coupée, ou lecteur inexistant.
func radioPlayer(s *discordgo.Session) disgolink.Player {
	if Link == nil || !Utils.GatewayConnected(s) {
		return nil
	}

	guildID, err := snowflake.Parse(os.Getenv("GUILD_ID"))
	if err != nil {
		return nil
	}

	return Link.ExistingPlayer(guildID)
}

// allowRecovery indique si une action de récupération peut être déclenchée et,
// si oui, note l'instant de la tentative. Sans ce filtre, un même incident
// signalé par deux évènements successifs ferait sauter deux morceaux d'un coup.
func allowRecovery(last *time.Time, debounce time.Duration) bool {
	recoveryMu.Lock()
	defer recoveryMu.Unlock()

	if !last.IsZero() && time.Since(*last) < debounce {
		return false
	}
	*last = time.Now()
	return true
}

// requestAdvance passe au morceau suivant en fusionnant les demandes trop
// rapprochées.
func requestAdvance(s *discordgo.Session, reason string) {
	if !allowRecovery(&lastAdvance, advanceDebounce) {
		return
	}

	fmt.Println("Radio :", reason, "— passage au morceau suivant.")
	Utils.SafeGo("le passage au morceau suivant", func() {
		AdvanceTrack(s)
	})
}

// requestVoiceRecovery refait la connexion au salon vocal et relance la lecture.
func requestVoiceRecovery(s *discordgo.Session, reason string) {
	if !allowRecovery(&lastVoiceRecovery, voiceRecoveryDebounce) {
		return
	}

	fmt.Println("Radio :", reason, "— reconnexion au salon vocal.")
	Utils.SafeGo("la reconnexion vocale de la radio", func() {
		if err := ConnectToRadioChannel(s); err != nil {
			fmt.Println("Radio : reconnexion impossible :", err)
		}
	})
}
