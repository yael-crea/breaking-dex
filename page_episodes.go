package main

// Page collection (GET /episodes) et page ressource (GET /episodes/{id}).

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
)

// lirePagination lit "page" et "par_page" (10, 20 ou 30, 10 par défaut).
func lirePagination(r *http.Request) (page, parPage int, err error) {
	page, err = entier(r, "page")
	if err != nil {
		return
	}
	parPage, err = entier(r, "par_page")
	if err != nil {
		return
	}
	if parPage == 0 {
		parPage = choixParPage[0]
	}
	if !slices.Contains(choixParPage, parPage) {
		err = fmt.Errorf("le nombre d'éléments par page doit être 10, 20 ou 30 (reçu : %d)", parPage)
	}
	return
}

// lireFiltres lit les filtres et le tri dans l'URL et refuse les valeurs invalides.
func lireFiltres(r *http.Request) (Filtres, error) {
	var f Filtres
	var err error
	if f.Saison, err = entier(r, "saison"); err != nil {
		return f, err
	}
	if f.Annee, err = entier(r, "annee"); err != nil {
		return f, err
	}
	if f.Duree, err = entier(r, "duree"); err != nil {
		return f, err
	}
	if v := r.URL.Query().Get("note"); v != "" {
		f.NoteMin, err = strconv.ParseFloat(v, 64)
		if err != nil || f.NoteMin < 0 || f.NoteMin > 10 {
			return f, fmt.Errorf("la note minimale doit être un nombre entre 0 et 10 (reçu : %q)", v)
		}
	}
	f.Tri = r.URL.Query().Get("tri")
	if f.Tri != "" && f.Tri != "diffusion" && f.Tri != "note" && f.Tri != "titre" {
		return f, fmt.Errorf("tri inconnu : %q", f.Tri)
	}
	return f, nil
}

type PageEpisodes struct {
	Base
	Episodes     []Episode
	Favoris      map[int]bool
	Filtres      Filtres
	Saisons      []Saison
	Annees       []int
	NotesMin     []float64
	DureesMax    []int
	ChoixParPage []int
	Pagination   Pagination
	URLActuelle  string
}

func collection(w http.ResponseWriter, r *http.Request) {
	page, parPage, err := lirePagination(r)
	if err != nil {
		erreur(w, http.StatusBadRequest, err.Error())
		return
	}
	f, err := lireFiltres(r)
	if err != nil {
		erreur(w, http.StatusBadRequest, err.Error())
		return
	}

	_, saisons, episodes, _, err := donnees()
	if err != nil {
		erreurAPI(w, err)
		return
	}

	resultats := trier(filtrer(episodes, f), f.Tri)

	// On garde les filtres dans les liens de pagination.
	params := url.Values{}
	for _, cle := range []string{"saison", "annee", "note", "duree", "tri"} {
		if v := r.URL.Query().Get(cle); v != "" {
			params.Set(cle, v)
		}
	}
	pg, debut, fin := paginer(len(resultats), page, parPage, "/episodes", params)

	afficher(w, http.StatusOK, "episodes", PageEpisodes{
		Base:         nouvelleBase("Épisodes", "episodes"),
		Episodes:     resultats[debut:fin],
		Favoris:      map[int]bool{},
		Filtres:      f,
		Saisons:      saisons,
		Annees:       anneesDisponibles(episodes),
		NotesMin:     notesMin,
		DureesMax:    dureesMax,
		ChoixParPage: choixParPage,
		Pagination:   pg,
		URLActuelle:  r.URL.RequestURI(),
	})
}

type PageEpisode struct {
	Base
	Episode     Episode
	Favori      bool
	Precedent   *Episode
	Suivant     *Episode
	URLActuelle string
}

func detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		erreur(w, http.StatusBadRequest, "L'identifiant de l'épisode doit être un nombre positif.")
		return
	}
	_, _, episodes, _, err := donnees()
	if err != nil {
		erreurAPI(w, err)
		return
	}
	// On vérifie que l'épisode appartient bien à Breaking Bad.
	position := slices.IndexFunc(episodes, func(e Episode) bool { return e.ID == id })
	if position == -1 {
		erreur(w, http.StatusNotFound, "Cet épisode n'existe pas dans Breaking Bad.")
		return
	}
	episode, err := detailEpisode(id)
	if err != nil {
		erreurAPI(w, err)
		return
	}
	donneesPage := PageEpisode{
		Base:        nouvelleBase(episode.Name, "episodes"),
		Episode:     episode,
		Favori:      false,
		URLActuelle: r.URL.RequestURI(),
	}
	if position > 0 {
		donneesPage.Precedent = &episodes[position-1]
	}
	if position < len(episodes)-1 {
		donneesPage.Suivant = &episodes[position+1]
	}
	afficher(w, http.StatusOK, "episode", donneesPage)
}
