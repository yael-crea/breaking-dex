package main

// Page favoris : GET /favoris, POST /favoris/ajouter, POST /favoris/supprimer, POST /favoris/vider

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type PageFavoris struct {
	Base
	Favoris []Favori
}

func favoris(w http.ResponseWriter, r *http.Request) {
	liste, err := listeFavoris()
	if err != nil {
		fmt.Println("Erreur favoris :", err)
		erreur(w, http.StatusInternalServerError, "Impossible de lire le fichier des favoris.")
		return
	}
	slices.Reverse(liste) // les plus récents en premier
	afficher(w, http.StatusOK, "favoris", PageFavoris{
		Base:    nouvelleBase("Favoris", "favoris"),
		Favoris: liste,
	})
}

// retour renvoie l'adresse où rediriger après un formulaire (uniquement une adresse du site).
func retour(r *http.Request) string {
	adresse := r.FormValue("retour")
	if strings.HasPrefix(adresse, "/") && !strings.HasPrefix(adresse, "//") && !strings.HasPrefix(adresse, "/\\") {
		return adresse
	}
	return "/favoris"
}

func idFormulaire(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("identifiant d'épisode invalide")
	}
	return id, nil
}

func ajouterAuxFavoris(w http.ResponseWriter, r *http.Request) {
	id, err := idFormulaire(r)
	if err != nil {
		erreur(w, http.StatusBadRequest, err.Error())
		return
	}
	_, _, episodes, _, err := donnees()
	if err != nil {
		erreurAPI(w, err)
		return
	}
	position := slices.IndexFunc(episodes, func(e Episode) bool { return e.ID == id })
	if position == -1 {
		erreur(w, http.StatusNotFound, "Cet épisode n'existe pas dans Breaking Bad.")
		return
	}
	if err := ajouterFavori(episodes[position]); err != nil {
		fmt.Println("Erreur favoris :", err)
		erreur(w, http.StatusInternalServerError, "Impossible d'enregistrer le favori.")
		return
	}
	http.Redirect(w, r, retour(r), http.StatusSeeOther)
}

func supprimerDesFavoris(w http.ResponseWriter, r *http.Request) {
	id, err := idFormulaire(r)
	if err != nil {
		erreur(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := supprimerFavori(id); err != nil {
		fmt.Println("Erreur favoris :", err)
		erreur(w, http.StatusInternalServerError, "Impossible de supprimer le favori.")
		return
	}
	http.Redirect(w, r, retour(r), http.StatusSeeOther)
}

func viderLesFavoris(w http.ResponseWriter, r *http.Request) {
	if err := viderFavoris(); err != nil {
		fmt.Println("Erreur favoris :", err)
		erreur(w, http.StatusInternalServerError, "Impossible de vider les favoris.")
		return
	}
	http.Redirect(w, r, "/favoris", http.StatusSeeOther)
}
