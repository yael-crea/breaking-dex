package main

// Ce fichier regroupe tout ce qui concerne l'API TVmaze :
// les structures des données reçues et les fonctions qui font les appels HTTP.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// URL de base de l'API TVmaze (gratuite, sans clé, format JSON).
const baseAPI = "https://api.tvmaze.com"

// Identifiant de la série Breaking Bad sur TVmaze.
const idSerie = 169

// Durée pendant laquelle les données gardées en mémoire restent valables.
const dureeCache = 10 * time.Minute

// Client HTTP avec un délai maximum : si l'API ne répond pas, on n'attend pas indéfiniment.
var client = &http.Client{Timeout: 10 * time.Second}

// ErrIntrouvable est renvoyée quand l'API répond 404.
var ErrIntrouvable = errors.New("ressource introuvable")

// ErrAPIIndisponible est renvoyée quand l'API est hors ligne ou répond une erreur.
var ErrAPIIndisponible = errors.New("l'API TVmaze est indisponible pour le moment")

/* ---------- Structures des données reçues ---------- */

type Image struct {
	Medium   string `json:"medium"`
	Original string `json:"original"`
}

type Note struct {
	Average float64 `json:"average"`
}

type Pays struct {
	Name string `json:"name"`
}

type Chaine struct {
	Name    string `json:"name"`
	Country *Pays  `json:"country"`
}

// Serie : réponse de GET /shows/169
type Serie struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Genres    []string `json:"genres"`
	Status    string   `json:"status"`
	Premiered string   `json:"premiered"`
	Ended     string   `json:"ended"`
	Rating    Note     `json:"rating"`
	Network   *Chaine  `json:"network"`
	Image     *Image   `json:"image"`
	Summary   string   `json:"summary"`
}

// Saison : élément de GET /shows/169/seasons
type Saison struct {
	ID           int    `json:"id"`
	Number       int    `json:"number"`
	EpisodeOrder int    `json:"episodeOrder"`
	PremiereDate string `json:"premiereDate"`
	EndDate      string `json:"endDate"`
	Image        *Image `json:"image"`
}

// Personne : un acteur (dans le casting ou via GET /people/{id})
type Personne struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Birthday string `json:"birthday"`
	Deathday string `json:"deathday"`
	Gender   string `json:"gender"`
	Country  *Pays  `json:"country"`
	Image    *Image `json:"image"`
}

// Role : le personnage joué par un acteur
type Role struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Image *Image `json:"image"`
}

// MembreCasting : élément de GET /shows/169/cast
type MembreCasting struct {
	Person    Personne `json:"person"`
	Character Role     `json:"character"`
}

// Episode : élément de GET /shows/169/episodes ou réponse de GET /episodes/{id}
type Episode struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Season   int    `json:"season"`
	Number   int    `json:"number"`
	Airdate  string `json:"airdate"`
	Runtime  int    `json:"runtime"`
	Rating   Note   `json:"rating"`
	Image    *Image `json:"image"`
	Summary  string `json:"summary"`
	Embedded *struct {
		Guestcast []MembreCasting `json:"guestcast"`
	} `json:"_embedded"`
}

/* ---------- Méthodes utilisées dans les templates ---------- */

var balisesHTML = regexp.MustCompile(`<[^>]*>`)

// texteBrut enlève les balises HTML renvoyées par l'API dans les résumés.
func texteBrut(s string) string {
	return strings.TrimSpace(balisesHTML.ReplaceAllString(s, ""))
}

// dateFR transforme "2008-01-20" en "20/01/2008".
func dateFR(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "inconnue"
	}
	return t.Format("02/01/2006")
}

// Resume renvoie le résumé en français s'il est traduit, sinon celui de l'API.
func (s Serie) Resume() string {
	if traductions.Serie != "" {
		return traductions.Serie
	}
	return texteBrut(s.Summary)
}

// GenresFR renvoie les genres traduits en français.
func (s Serie) GenresFR() []string {
	var genres []string
	for _, g := range s.Genres {
		genres = append(genres, traduire(genresFR, g))
	}
	return genres
}

func (s Saison) Debut() string { return dateFR(s.PremiereDate) }
func (s Saison) Fin() string   { return dateFR(s.EndDate) }

// Resume renvoie le résumé en français s'il est traduit, sinon celui de l'API.
func (e Episode) Resume() string {
	if r, ok := traductions.Episodes[strconv.Itoa(e.ID)]; ok {
		return r
	}
	return texteBrut(e.Summary)
}
func (e Episode) Date() string   { return dateFR(e.Airdate) }

// Code renvoie le code de l'épisode, ex. "S01E05".
func (e Episode) Code() string { return fmt.Sprintf("S%02dE%02d", e.Season, e.Number) }

// Annee renvoie l'année de diffusion (0 si inconnue).
func (e Episode) Annee() int {
	t, err := time.Parse("2006-01-02", e.Airdate)
	if err != nil {
		return 0
	}
	return t.Year()
}

func (e Episode) Vignette() string {
	if e.Image == nil {
		return ""
	}
	return e.Image.Medium
}

func (e Episode) GrandeImage() string {
	if e.Image == nil {
		return ""
	}
	return e.Image.Original
}

func (e Episode) InvitesListe() []MembreCasting {
	if e.Embedded == nil {
		return nil
	}
	return e.Embedded.Guestcast
}

// Portrait renvoie l'image du personnage, ou à défaut celle de l'acteur.
func (m MembreCasting) Portrait() string {
	if m.Character.Image != nil {
		return m.Character.Image.Medium
	}
	if m.Person.Image != nil {
		return m.Person.Image.Medium
	}
	return ""
}

func (p Personne) Naissance() string { return dateFR(p.Birthday) }
func (p Personne) Deces() string     { return dateFR(p.Deathday) }

func (p Personne) Genre() string {
	switch p.Gender {
	case "Male":
		return "Homme"
	case "Female":
		return "Femme"
	default:
		return "Non renseigné"
	}
}

func (p Personne) NomPays() string {
	if p.Country == nil {
		return "Non renseigné"
	}
	return traduire(paysFR, p.Country.Name)
}

func (p Personne) Photo() string {
	if p.Image == nil {
		return ""
	}
	return p.Image.Medium
}

/* ---------- Appels à l'API ---------- */

// appelAPI fait un GET sur l'API et décode le JSON dans "cible".
// Il fait 2 essais en cas de panne réseau ou d'erreur 5xx / 429.
func appelAPI(chemin string, cible any) error {
	var derniereErreur error
	for essai := 1; essai <= 2; essai++ {
		resp, err := client.Get(baseAPI + chemin)
		if err != nil {
			derniereErreur = err
		} else {
			switch {
			case resp.StatusCode == http.StatusOK:
				err = json.NewDecoder(resp.Body).Decode(cible)
				resp.Body.Close()
				if err != nil {
					return fmt.Errorf("%w (réponse JSON illisible)", ErrAPIIndisponible)
				}
				return nil
			case resp.StatusCode == http.StatusNotFound:
				resp.Body.Close()
				return ErrIntrouvable
			default:
				resp.Body.Close()
				derniereErreur = fmt.Errorf("code HTTP %d", resp.StatusCode)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Printf("Erreur API sur %s : %v\n", chemin, derniereErreur)
	return ErrAPIIndisponible
}

// Les données de la série changent rarement : on les garde en mémoire quelques minutes
// pour ne pas rappeler l'API à chaque page affichée.
type cacheDonnees struct {
	verrou   sync.Mutex
	charge   time.Time
	serie    Serie
	saisons  []Saison
	episodes []Episode
	casting  []MembreCasting
}

var cache cacheDonnees

// donnees renvoie la série, ses saisons, ses épisodes et son casting
// (endpoints /shows/169, /shows/169/seasons, /shows/169/episodes, /shows/169/cast).
func donnees() (Serie, []Saison, []Episode, []MembreCasting, error) {
	cache.verrou.Lock()
	defer cache.verrou.Unlock()

	if time.Since(cache.charge) < dureeCache {
		return cache.serie, cache.saisons, cache.episodes, cache.casting, nil
	}

	var serie Serie
	var saisons []Saison
	var episodes []Episode
	var casting []MembreCasting
	base := fmt.Sprintf("/shows/%d", idSerie)
	err := appelAPI(base, &serie)
	if err == nil {
		err = appelAPI(base+"/seasons", &saisons)
	}
	if err == nil {
		err = appelAPI(base+"/episodes", &episodes)
	}
	if err == nil {
		err = appelAPI(base+"/cast", &casting)
	}
	if err != nil {
		// Si l'API est tombée mais qu'on a d'anciennes données, on les réutilise.
		if !cache.charge.IsZero() {
			return cache.serie, cache.saisons, cache.episodes, cache.casting, nil
		}
		return Serie{}, nil, nil, nil, err
	}

	cache.serie, cache.saisons, cache.episodes, cache.casting = serie, saisons, episodes, casting
	cache.charge = time.Now()
	return serie, saisons, episodes, casting, nil
}

// detailEpisode appelle GET /episodes/{id}?embed=guestcast (épisode + acteurs invités).
func detailEpisode(id int) (Episode, error) {
	var e Episode
	err := appelAPI(fmt.Sprintf("/episodes/%d?embed=guestcast", id), &e)
	return e, err
}

// detailPersonne appelle GET /people/{id}.
func detailPersonne(id int) (Personne, error) {
	var p Personne
	err := appelAPI(fmt.Sprintf("/people/%d", id), &p)
	return p, err
}
