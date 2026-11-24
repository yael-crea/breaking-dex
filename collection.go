package main

// Ce fichier contient la recherche, les filtres, le tri et la pagination.
// Tout est fait à la main en Go : on parcourt nous-mêmes la collection reçue de l'API,
// sans utiliser les paramètres de recherche / filtre / pagination de l'API.

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
)

/* ---------- FT1 : recherche ---------- */

// Remplacement des lettres accentuées pour que "épisode" trouve "episode".
var sansAccents = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a", "á", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"î", "i", "ï", "i", "í", "i",
	"ô", "o", "ö", "o", "ó", "o",
	"ù", "u", "û", "u", "ü", "u", "ú", "u",
	"ç", "c", "ñ", "n",
)

// normaliser met le texte en minuscules, sans accents et sans espaces autour.
func normaliser(s string) string {
	return sansAccents.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// contient dit si "texte" contient "recherche", sans tenir compte de la casse ni des accents.
func contient(texte, recherche string) bool {
	return strings.Contains(normaliser(texte), normaliser(recherche))
}

// rechercherEpisodes garde les épisodes dont le titre OU le résumé contient la recherche.
func rechercherEpisodes(episodes []Episode, q string) []Episode {
	var resultats []Episode
	for _, e := range episodes {
		if contient(e.Name, q) || contient(e.Resume(), q) {
			resultats = append(resultats, e)
		}
	}
	return resultats
}

// rechercherPersonnages garde les membres du casting dont le nom du personnage OU de l'acteur contient la recherche.
func rechercherPersonnages(casting []MembreCasting, q string) []MembreCasting {
	var resultats []MembreCasting
	for _, m := range casting {
		if contient(m.Character.Name, q) || contient(m.Person.Name, q) {
			resultats = append(resultats, m)
		}
	}
	return resultats
}

/* ---------- FT2 : filtres cumulables ---------- */

// Filtres choisis par l'utilisateur. La valeur 0 veut dire "pas de filtre".
type Filtres struct {
	Saison  int     // numéro de saison
	Annee   int     // année de diffusion
	NoteMin float64 // note minimale sur 10
	Duree   int     // durée maximale en minutes
	Tri     string  // "diffusion", "note" ou "titre"
}

// filtrer garde uniquement les épisodes qui respectent TOUS les filtres actifs.
func filtrer(episodes []Episode, f Filtres) []Episode {
	var resultats []Episode
	for _, e := range episodes {
		if f.Saison != 0 && e.Season != f.Saison {
			continue
		}
		if f.Annee != 0 && e.Annee() != f.Annee {
			continue
		}
		if f.NoteMin != 0 && e.Rating.Average < f.NoteMin {
			continue
		}
		if f.Duree != 0 && e.Runtime > f.Duree {
			continue
		}
		resultats = append(resultats, e)
	}
	return resultats
}

// trier range les épisodes (bonus). On travaille sur une copie pour ne pas modifier le cache.
func trier(episodes []Episode, tri string) []Episode {
	copie := append([]Episode(nil), episodes...)
	switch tri {
	case "note":
		sort.SliceStable(copie, func(i, j int) bool { return copie[i].Rating.Average > copie[j].Rating.Average })
	case "titre":
		sort.SliceStable(copie, func(i, j int) bool { return normaliser(copie[i].Name) < normaliser(copie[j].Name) })
	}
	return copie
}

// Choix proposés dans les listes déroulantes.
func anneesDisponibles(episodes []Episode) []int {
	var annees []int
	for _, e := range episodes {
		a := e.Annee()
		if a != 0 && (len(annees) == 0 || annees[len(annees)-1] != a) {
			annees = append(annees, a)
		}
	}
	return annees
}

var notesMin = []float64{7, 7.5, 8, 8.5, 9, 9.5}
var dureesMax = []int{45, 50, 55, 60}

/* ---------- FT3 : pagination ---------- */

// Nombres d'éléments par page que l'utilisateur peut choisir.
var choixParPage = []int{10, 20, 30}

type LienPage struct {
	Numero int
	URL    string
	Actif  bool
}

type Pagination struct {
	Page       int
	Pages      int
	ParPage    int
	Total      int
	Premier    int // numéro du premier élément affiché
	Dernier    int // numéro du dernier élément affiché
	Liens      []LienPage
	Precedente string
	Suivante   string
}

// paginer calcule la pagination et renvoie les indices [debut, fin) des éléments à afficher.
// "params" sert à garder la recherche et les filtres dans les liens des pages.
func paginer(total, page, parPage int, chemin string, params url.Values) (Pagination, int, int) {
	pages := (total + parPage - 1) / parPage
	if pages == 0 {
		pages = 1
	}
	page = max(1, min(page, pages)) // une page trop grande ou négative est ramenée dans les bornes

	lien := func(n int) string {
		p := url.Values{}
		for cle, valeurs := range params {
			p[cle] = valeurs
		}
		p.Set("page", strconv.Itoa(n))
		p.Set("par_page", strconv.Itoa(parPage))
		return chemin + "?" + p.Encode()
	}

	debut := (page - 1) * parPage
	fin := min(debut+parPage, total)
	pg := Pagination{Page: page, Pages: pages, ParPage: parPage, Total: total}
	if total > 0 {
		pg.Premier, pg.Dernier = debut+1, fin
	}
	for n := 1; n <= pages; n++ {
		pg.Liens = append(pg.Liens, LienPage{Numero: n, URL: lien(n), Actif: n == page})
	}
	if page > 1 {
		pg.Precedente = lien(page - 1)
	}
	if page < pages {
		pg.Suivante = lien(page + 1)
	}
	return pg, debut, fin
}
