package main

// Ce fichier contient ce qui est partagé par toutes les pages :
// données communes, affichage des templates, pages d'erreur et lecture des paramètres.

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var templates *template.Template

// Base : données communes à toutes les pages (titre, menu actif, nombre de favoris).
type Base struct {
	Titre     string
	Actif     string
	NbFavoris int
	Recherche string
}

func nouvelleBase(titre, actif string) Base {
	favoris, _ := listeFavoris()
	return Base{Titre: titre, Actif: actif, NbFavoris: len(favoris)}
}

// afficher exécute le template dans un tampon avant de l'envoyer :
// si le template plante, on peut encore renvoyer une vraie page d'erreur 500.
func afficher(w http.ResponseWriter, statut int, nom string, donnees any) {
	var tampon bytes.Buffer
	if err := templates.ExecuteTemplate(&tampon, nom, donnees); err != nil {
		fmt.Println("Erreur de template :", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statut)
	tampon.WriteTo(w)
}

// CarteEpisode : données d'une carte d'épisode (template "carte-episode").
type CarteEpisode struct {
	Episode Episode
	Favori  bool
	Retour  string
}

type PageErreur struct {
	Base
	Code    int
	Message string
}

// erreur affiche la page d'erreur avec le bon code HTTP.
func erreur(w http.ResponseWriter, code int, message string) {
	afficher(w, code, "erreur", PageErreur{
		Base:    nouvelleBase(fmt.Sprintf("Erreur %d", code), ""),
		Code:    code,
		Message: message,
	})
}

// erreurAPI choisit la page d'erreur adaptée à une erreur renvoyée par l'API.
func erreurAPI(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrIntrouvable) {
		erreur(w, http.StatusNotFound, "Cette ressource n'existe pas.")
		return
	}
	erreur(w, http.StatusServiceUnavailable, "L'API TVmaze ne répond pas pour le moment. Réessaie dans quelques instants.")
}

// entier lit un paramètre numérique de l'URL. Paramètre absent = 0.
func entier(r *http.Request, nom string) (int, error) {
	v := strings.TrimSpace(r.URL.Query().Get(nom))
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("le paramètre « %s » doit être un nombre positif (reçu : %q)", nom, v)
	}
	return n, nil
}

// pageIntrouvable répond à toutes les routes qui n'existent pas.
func pageIntrouvable(w http.ResponseWriter, r *http.Request) {
	erreur(w, http.StatusNotFound, "La page demandée n'existe pas.")
}
