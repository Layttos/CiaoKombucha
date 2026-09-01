package Radio

import (
	"fmt"
	"os"
	"sync"

	"bot.ciaokombucha.tv/Utils"
	"github.com/bwmarrin/discordgo"
)

// maxQueueLength borne la file d'attente des morceaux demandés via /play.
const maxQueueLength = 25

var (
	requestQueue   []RequestedTrack
	requestQueueMu sync.Mutex
)

// EnqueueRequest résout la requête sur Navidrome et l'ajoute à la file
// d'attente. Renvoie le morceau ajouté et sa position dans la file
// (1 = prochain morceau joué, juste après la musique en cours).
func EnqueueRequest(query string, requestedBy string) (*RequestedTrack, int, error) {
	track, err := resolveTrack(query)
	if err != nil {
		return nil, 0, err
	}
	track.RequestedBy = requestedBy

	requestQueueMu.Lock()
	defer requestQueueMu.Unlock()

	if len(requestQueue) >= maxQueueLength {
		return nil, 0, fmt.Errorf("la file d'attente est pleine (%d morceaux)", maxQueueLength)
	}

	requestQueue = append(requestQueue, *track)
	return track, len(requestQueue), nil
}

// dequeueRequest retire et renvoie le prochain morceau demandé, s'il y en a un.
func dequeueRequest() (RequestedTrack, bool) {
	requestQueueMu.Lock()
	defer requestQueueMu.Unlock()

	if len(requestQueue) == 0 {
		return RequestedTrack{}, false
	}

	next := requestQueue[0]
	requestQueue = requestQueue[1:]
	return next, true
}

// QueueLength renvoie le nombre de morceaux en attente.
func QueueLength() int {
	requestQueueMu.Lock()
	defer requestQueueMu.Unlock()
	return len(requestQueue)
}

// AdvanceTrack décide de la suite après la fin d'un morceau (ou un skip) : jouer
// le prochain morceau demandé via /play s'il y en a un, sinon reprendre la radio
// normale (morceau aléatoire).
func AdvanceTrack(s *discordgo.Session) {
	// Un morceau se termine souvent pendant une coupure de la gateway : sans ce
	// garde-fou, le (re)join vocal ci-dessous écrirait sur un websocket fermé et
	// ferait tomber le bot. La radio repart d'elle-même à la reconnexion.
	if !Utils.GatewayConnected(s) {
		fmt.Println("Gateway Discord injoignable : passage au morceau suivant reporté.")
		return
	}

	next, ok := dequeueRequest()
	if !ok {
		if err := ConnectToRadioChannel(s); err != nil {
			fmt.Println("An error occurred while attempting to resume the radio:", err)
		}
		return
	}

	if err := s.ChannelVoiceJoinManual(os.Getenv("GUILD_ID"), os.Getenv("RADIO_CHANNEL_ID"), false, false); err != nil {
		fmt.Println("An error occurred while attempting to (re)join the radio channel:", err)
	}

	if err := playStreamURL(next.StreamURL); err != nil {
		fmt.Println("Queued request failed, falling back to the radio:", err)
		if err := ConnectToRadioChannel(s); err != nil {
			fmt.Println("An error occurred while attempting to resume the radio:", err)
		}
	}
}
