package main

// Pages personnages : GET /personnages et GET /personnages/{id}

import (
	"net/http"
	"strconv"
)

type PagePersonnages struct {
	Base
	Casting []MembreCasting
}

func personnages(w http.ResponseWriter, r *http.Request) {
	_, _, _, casting, err := donnees()
	if err != nil {
		erreurAPI(w, err)
		return
	}
	afficher(w, http.StatusOK, "personnages", PagePersonnages{
		Base:    nouvelleBase("Personnages", "personnages"),
		Casting: casting,
	})
}

type PagePersonnage struct {
	Base
	Acteur Personne
	Roles  []Role
}

func personnage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		erreur(w, http.StatusBadRequest, "L'identifiant de l'acteur doit être un nombre positif.")
		return
	}
	_, _, _, casting, err := donnees()
	if err != nil {
		erreurAPI(w, err)
		return
	}
	var roles []Role
	for _, m := range casting {
		if m.Person.ID == id {
			roles = append(roles, m.Character)
		}
	}
	if len(roles) == 0 {
		erreur(w, http.StatusNotFound, "Cet acteur ne fait pas partie du casting principal de Breaking Bad.")
		return
	}
	acteur, err := detailPersonne(id)
	if err != nil {
		erreurAPI(w, err)
		return
	}
	afficher(w, http.StatusOK, "personnage", PagePersonnage{
		Base:   nouvelleBase(acteur.Name, "personnages"),
		Acteur: acteur,
		Roles:  roles,
	})
}
