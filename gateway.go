package main

import (
	"errors"
	"log"
	"os"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"bot.ciaokombucha.tv/Radio"
	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

const (
	// Nombre de tentatives d'ouverture de la session au démarrage. Avec le
	// backoff ci-dessous, cela laisse une dizaine de minutes à Discord (ou au
	// réseau de la machine) pour revenir.
	startupOpenAttempts = 10

	// Fréquence à laquelle le chien de garde inspecte la connexion.
	watchdogInterval = 30 * time.Second

	// Délai laissé à discordgo pour se reconnecter tout seul avant qu'on force
	// une session neuve. Son backoff interne (1, 2, 4, 8... secondes) règle la
	// grande majorité des coupures : inutile de lui couper l'herbe sous le pied.
	watchdogGraceBeforeForcing = 3 * time.Minute

	// Intervalle entre deux reconnexions forcées. Il double à chaque échec
	// jusqu'au plafond : une reconnexion complète consomme un IDENTIFY, et
	// Discord n'en accorde qu'un millier par jour. Le chien de garde n'abandonne
	// jamais — le bot tourne dans un screen, sans superviseur pour le relancer,
	// donc s'arrêter ne ferait que transformer « hors ligne » en « mort ».
	watchdogRetryMin = 2 * time.Minute
	watchdogRetryMax = 15 * time.Minute
)

// safeHandler enveloppe un listener discordgo dans un recover.
//
// discordgo appelle chaque handler dans sa propre goroutine et ne rattrape
// rien : un simple déréférencement nil (utilisateur introuvable, événement
// partiel envoyé par Discord, base de données momentanément indisponible...)
// suffit à tuer le processus entier. Ici, l'incident se limite à une ligne de
// log et le bot continue de tourner.
func safeHandler[E any](handler func(*discordgo.Session, E)) func(*discordgo.Session, E) {
	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	return func(s *discordgo.Session, e E) {
		defer Utils.RecoverPanic(name)
		handler(s, e)
	}
}

// openSession ouvre la session en réessayant. Au démarrage — redémarrage de la
// machine, coupure réseau, incident côté Discord — un échec ne doit pas laisser
// le bot mort jusqu'à une intervention manuelle.
func openSession(s *discordgo.Session) error {
	wait := 5 * time.Second

	for attempt := 1; ; attempt++ {
		err := s.Open()
		if err == nil {
			return nil
		}
		if attempt >= startupOpenAttempts {
			return err
		}

		log.Printf("Connexion à Discord impossible (tentative %d/%d) : %v — nouvel essai dans %s.", attempt, startupOpenAttempts, err, wait)
		time.Sleep(wait)
		if wait < 2*time.Minute {
			wait *= 2
		}
	}
}

// startGatewayWatchdog surveille la connexion gateway et la reconstruit quand
// discordgo n'arrive plus à la rétablir de lui-même.
func startGatewayWatchdog(s *discordgo.Session) {
	Utils.SafeGo("le chien de garde de la gateway", func() {
		ticker := time.NewTicker(watchdogInterval)
		defer ticker.Stop()

		var downSince, nextAttempt time.Time
		retryIn := watchdogRetryMin

		for range ticker.C {
			if Utils.GatewayConnected(s) {
				if !downSince.IsZero() {
					log.Printf("Gateway : connexion rétablie après %s.", time.Since(downSince).Round(time.Second))
					downSince, nextAttempt = time.Time{}, time.Time{}
					retryIn = watchdogRetryMin
				}
				continue
			}

			now := time.Now()

			if downSince.IsZero() {
				downSince = now
				nextAttempt = now.Add(watchdogGraceBeforeForcing)
				log.Println("Gateway : plus de réponse, on laisse discordgo se reconnecter seul.")
				continue
			}

			if now.Before(nextAttempt) {
				continue
			}
			nextAttempt = now.Add(retryIn)

			log.Printf("Gateway : hors ligne depuis %s, reconnexion forcée sur une session neuve (prochain essai dans %s en cas d'échec).",
				now.Sub(downSince).Round(time.Second), retryIn)
			forceReconnect(s)

			retryIn = min(retryIn*2, watchdogRetryMax)
		}
	})
}

// forceReconnect repart d'une session gateway neuve : on jette l'état de reprise
// (périmé, c'est justement ce qui bloque), on ferme ce qui traîne, puis on
// rouvre.
//
// Effet de bord utile : si la boucle de reconnexion interne de discordgo tourne
// encore, son prochain Open() renverra ErrWSAlreadyOpen et elle s'arrêtera
// proprement au lieu de retenter indéfiniment dans le vide.
func forceReconnect(s *discordgo.Session) {
	defer Utils.RecoverPanic("la reconnexion forcée")

	// La boucle interne a pu réussir entre-temps.
	if Utils.GatewayConnected(s) {
		return
	}

	if !resetGatewayResumeState(s) {
		log.Println("Gateway : impossible de réinitialiser l'état de reprise de discordgo (champs internes renommés ?). La reconnexion a peu de chances d'aboutir.")
	}

	// La connexion est déjà morte : l'erreur éventuelle n'a pas d'intérêt, seul
	// compte le fait que discordgo relâche son websocket pour qu'Open() reparte.
	_ = s.Close()

	switch err := s.Open(); {
	case err == nil:
		log.Println("Gateway : reconnexion forcée réussie.")
		onGatewayReconnected(s)
	case errors.Is(err, discordgo.ErrWSAlreadyOpen):
		// La boucle de reconnexion interne de discordgo a repris la main juste
		// avant nous : il n'y a rien à faire.
	default:
		log.Println("Gateway : la reconnexion forcée a échoué :", err)
	}
}

// onGatewayReconnected remet en route ce qui ne survit pas à une session neuve.
// Une reconnexion complète invalide la connexion vocale : il faut rejoindre le
// salon et relancer la lecture.
func onGatewayReconnected(s *discordgo.Session) {
	if os.Getenv("ENABLE_RADIO") != "true" {
		return
	}

	Utils.SafeGo("la reprise de la radio", func() {
		if err := Radio.ConnectToRadioChannel(s); err != nil {
			log.Println("Radio : reprise impossible après la reconnexion :", err)
		}
	})
}

// resetGatewayResumeState efface les informations de reprise de session mises en
// cache par discordgo : sessionID, sequence et resumeGatewayURL.
//
// Pourquoi c'est nécessaire — c'est la cause des « websocket: bad handshake »
// qui reviennent toutes les dix minutes sans jamais se résoudre :
//
// Session.Open() se reconnecte sur resumeGatewayURL dès que ces trois champs
// sont renseignés. Or cette URL n'est valable que tant que la session gateway
// l'est ; une fois celle-ci expirée côté Discord, elle répond une erreur HTTP et
// le handshake websocket échoue. Et discordgo ne remet ces champs à zéro que
// lorsqu'il reçoit un op 9 (Invalid Session) *sur une connexion établie* — ce
// qui ne peut plus jamais arriver puisque la connexion ne s'établit plus. Le bot
// reste donc définitivement hors ligne, à retenter la même URL morte toutes les
// dix minutes (le backoff de discordgo est plafonné à 600 secondes).
//
// Ces champs ne sont pas exportés, d'où reflect + unsafe. Si une version future
// de discordgo les renomme, la fonction renvoie false et l'appelant se rabat sur
// l'arrêt du processus au bout de watchdogGiveUpAfter.
func resetGatewayResumeState(s *discordgo.Session) bool {
	session := reflect.ValueOf(s).Elem()

	sequence := session.FieldByName("sequence")
	sessionID := session.FieldByName("sessionID")
	resumeURL := session.FieldByName("resumeGatewayURL")

	if !sequence.IsValid() || sequence.Type() != reflect.TypeOf((*int64)(nil)) ||
		!sessionID.IsValid() || sessionID.Kind() != reflect.String ||
		!resumeURL.IsValid() || resumeURL.Kind() != reflect.String {
		return false
	}

	s.Lock()
	defer s.Unlock()

	if counter, ok := reflect.NewAt(sequence.Type(), unsafe.Pointer(sequence.UnsafeAddr())).Elem().Interface().(*int64); ok && counter != nil {
		atomic.StoreInt64(counter, 0)
	}
	reflect.NewAt(sessionID.Type(), unsafe.Pointer(sessionID.UnsafeAddr())).Elem().SetString("")
	reflect.NewAt(resumeURL.Type(), unsafe.Pointer(resumeURL.UnsafeAddr())).Elem().SetString("")

	return true
}
