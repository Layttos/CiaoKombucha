package Radio

import "sync"

var (
	skipVotes   = map[string]bool{}
	skipVotesMu sync.Mutex
)

// resetSkipVotes efface les votes en cours. Appelé à chaque changement de
// morceau (vote abouti, /radiosearch admin, ou passage automatique).
func resetSkipVotes() {
	skipVotesMu.Lock()
	skipVotes = map[string]bool{}
	skipVotesMu.Unlock()
}

// SkipVotesNeeded renvoie le nombre de votes requis pour passer le morceau :
// la moitié des auditeurs présents, arrondie au supérieur.
//
//	1 auditeur  -> 1   (seul avec le bot : passage direct)
//	2 auditeurs -> 1
//	3 auditeurs -> 2
//	4 auditeurs -> 2
//	5 auditeurs -> 3
func SkipVotesNeeded(listeners int) int {
	if listeners < 1 {
		listeners = 1
	}
	return (listeners + 1) / 2
}

// AddSkipVote enregistre le vote de userID et indique si le seuil est atteint.
// listenerIDs est la liste des auditeurs humains actuellement dans le salon :
// les votes des personnes qui ont quitté sont automatiquement retirés.
func AddSkipVote(userID string, listenerIDs []string) (reached bool, tally int, needed int, already bool) {
	skipVotesMu.Lock()
	defer skipVotesMu.Unlock()

	already = skipVotes[userID]
	skipVotes[userID] = true

	present := make(map[string]bool, len(listenerIDs))
	for _, id := range listenerIDs {
		present[id] = true
	}
	for id := range skipVotes {
		if !present[id] {
			delete(skipVotes, id)
		}
	}

	tally = len(skipVotes)
	needed = SkipVotesNeeded(len(listenerIDs))
	return tally >= needed, tally, needed, already
}
