package main

// Page recherche : GET /recherche?q=...

import (
	"net/http"
	"net/url"
	"strings"
)

type PageRecherche struct {
	Base
	Q            string
	Episodes     []Episode
	Personnages  []MembreCasting
	Favoris      map[int]bool
	ChoixParPage []int
	Pagination   Pagination
	URLActuelle  string
}

func recherche(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	page, parPage, err := lirePagination(r)
	if err != nil {
		erreur(w, http.StatusBadRequest, err.Error())
		return
	}
	if len([]rune(q)) > 100 {
		erreur(w, http.StatusBadRequest, "La recherche ne doit pas dépasser 100 caractères.")
		return
	}

	donneesPage := PageRecherche{
		Base:         nouvelleBase("Recherche", "recherche"),
		Q:            q,
		Favoris:      map[int]bool{},
		ChoixParPage: choixParPage,
		Pagination:   Pagination{ParPage: parPage},
		URLActuelle:  r.URL.RequestURI(),
	}
	donneesPage.Recherche = q

	if q != "" {
		_, _, episodes, casting, err := donnees()
		if err != nil {
			erreurAPI(w, err)
			return
		}
		resultats := rechercherEpisodes(episodes, q)
		pg, debut, fin := paginer(len(resultats), page, parPage, "/recherche", url.Values{"q": {q}})
		donneesPage.Episodes = resultats[debut:fin]
		donneesPage.Pagination = pg
		donneesPage.Personnages = rechercherPersonnages(casting, q)
	}
	afficher(w, http.StatusOK, "recherche", donneesPage)
}
