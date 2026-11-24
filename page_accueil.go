package main

// Page d'accueil : GET /

import (
	"net/http"
	"sort"
)

type PageAccueil struct {
	Base
	Serie        Serie
	Saisons      []Saison
	MeilleursEps []Episode
	Casting      []MembreCasting
	NbEpisodes   int
}

// meilleursEpisodes renvoie les n épisodes les mieux notés (sans modifier la liste d'origine).
func meilleursEpisodes(episodes []Episode, n int) []Episode {
	copie := append([]Episode(nil), episodes...)
	sort.SliceStable(copie, func(i, j int) bool { return copie[i].Rating.Average > copie[j].Rating.Average })
	return copie[:min(n, len(copie))]
}

func accueil(w http.ResponseWriter, r *http.Request) {
	serie, saisons, episodes, casting, err := donnees()
	if err != nil {
		erreurAPI(w, err)
		return
	}
	afficher(w, http.StatusOK, "accueil", PageAccueil{
		Base:         nouvelleBase("Accueil", "accueil"),
		Serie:        serie,
		Saisons:      saisons,
		MeilleursEps: meilleursEpisodes(episodes, 4),
		Casting:      casting[:min(4, len(casting))],
		NbEpisodes:   len(episodes),
	})
}
