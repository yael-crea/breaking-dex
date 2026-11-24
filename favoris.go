package main

// FT4 : les favoris sont enregistrés dans un seul fichier JSON côté serveur.
// La liste est commune à tous les visiteurs (pas de compte utilisateur).

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

const fichierFavoris = "favoris.json"

// Favori : on garde une copie des infos utiles de l'épisode,
// pour pouvoir afficher les favoris même si l'API est hors ligne.
type Favori struct {
	ID       int       `json:"id"`
	Nom      string    `json:"nom"`
	Code     string    `json:"code"`
	Image    string    `json:"image"`
	AjouteLe time.Time `json:"ajoute_le"`
}

var verrouFavoris sync.Mutex

// lireFavoris lit le fichier. S'il n'existe pas encore, la liste est vide.
func lireFavoris() ([]Favori, error) {
	contenu, err := os.ReadFile(fichierFavoris)
	if errors.Is(err, os.ErrNotExist) {
		return []Favori{}, nil
	}
	if err != nil {
		return nil, err
	}
	var favoris []Favori
	if len(contenu) == 0 {
		return []Favori{}, nil
	}
	if err := json.Unmarshal(contenu, &favoris); err != nil {
		return nil, err
	}
	return favoris, nil
}

// ecrireFavoris écrit d'abord dans un fichier temporaire puis le renomme,
// pour ne jamais laisser un fichier à moitié écrit.
func ecrireFavoris(favoris []Favori) error {
	contenu, err := json.MarshalIndent(favoris, "", "  ")
	if err != nil {
		return err
	}
	tmp := fichierFavoris + ".tmp"
	if err := os.WriteFile(tmp, contenu, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, fichierFavoris)
}

// listeFavoris renvoie les favoris (accès protégé par le verrou).
func listeFavoris() ([]Favori, error) {
	verrouFavoris.Lock()
	defer verrouFavoris.Unlock()
	return lireFavoris()
}

// idsFavoris renvoie un ensemble des identifiants en favoris, pour afficher l'étoile sur les cartes.
func idsFavoris() map[int]bool {
	ids := map[int]bool{}
	favoris, err := listeFavoris()
	if err != nil {
		return ids
	}
	for _, f := range favoris {
		ids[f.ID] = true
	}
	return ids
}

// ajouterFavori ajoute un épisode s'il n'y est pas déjà.
func ajouterFavori(e Episode) error {
	verrouFavoris.Lock()
	defer verrouFavoris.Unlock()
	favoris, err := lireFavoris()
	if err != nil {
		return err
	}
	for _, f := range favoris {
		if f.ID == e.ID {
			return nil
		}
	}
	favoris = append(favoris, Favori{ID: e.ID, Nom: e.Name, Code: e.Code(), Image: e.Vignette(), AjouteLe: time.Now()})
	return ecrireFavoris(favoris)
}

// supprimerFavori retire un épisode de la liste.
func supprimerFavori(id int) error {
	verrouFavoris.Lock()
	defer verrouFavoris.Unlock()
	favoris, err := lireFavoris()
	if err != nil {
		return err
	}
	restants := []Favori{}
	for _, f := range favoris {
		if f.ID != id {
			restants = append(restants, f)
		}
	}
	return ecrireFavoris(restants)
}

// viderFavoris supprime tous les favoris (bonus).
func viderFavoris() error {
	verrouFavoris.Lock()
	defer verrouFavoris.Unlock()
	return ecrireFavoris([]Favori{})
}
