package Radio

import (
	"sync"
	"testing"
	"time"
)

// TestAllowRecoveryFusionneLesDemandesRapprochees : Lavalink signale un même
// incident par plusieurs évènements (TrackException puis TrackEnd "loadFailed",
// par exemple). Sans fusion, le bot sauterait deux morceaux au lieu d'un.
func TestAllowRecoveryFusionneLesDemandesRapprochees(t *testing.T) {
	var last time.Time

	if !allowRecovery(&last, time.Minute) {
		t.Fatal("la première demande doit toujours passer")
	}
	if allowRecovery(&last, time.Minute) {
		t.Fatal("une seconde demande immédiate doit être fusionnée avec la première")
	}
}

// TestAllowRecoveryRepasseApresLaFenetre vérifie qu'on ne se retrouve pas
// bloqué : passé le délai, une nouvelle récupération doit pouvoir se déclencher.
func TestAllowRecoveryRepasseApresLaFenetre(t *testing.T) {
	var last time.Time

	if !allowRecovery(&last, 10*time.Millisecond) {
		t.Fatal("la première demande doit toujours passer")
	}
	time.Sleep(20 * time.Millisecond)
	if !allowRecovery(&last, 10*time.Millisecond) {
		t.Fatal("la fenêtre est écoulée, la demande doit repasser")
	}
}

// TestAllowRecoveryEstSurACcesConcurrent : les évènements Lavalink et le chien
// de garde tournent dans des goroutines distinctes et peuvent réagir au même
// incident en même temps. Une seule doit l'emporter.
func TestAllowRecoveryEstSurAccesConcurrent(t *testing.T) {
	var last time.Time
	var passees int64
	var compteur sync.Mutex
	var attente sync.WaitGroup

	for range 50 {
		attente.Add(1)
		go func() {
			defer attente.Done()
			if allowRecovery(&last, time.Minute) {
				compteur.Lock()
				passees++
				compteur.Unlock()
			}
		}()
	}
	attente.Wait()

	if passees != 1 {
		t.Fatalf("%d demandes sont passées, une seule était attendue", passees)
	}
}
