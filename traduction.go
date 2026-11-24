package main

// L'API TVmaze ne fournit ses textes qu'en anglais.
// Ce fichier charge les traductions françaises (donnees/traductions_fr.json)
// et contient les petits dictionnaires pour les genres et les pays.

import (
	"encoding/json"
	"fmt"
	"os"
)

const fichierTraductions = "donnees/traductions_fr.json"

// Traductions : résumé de la série et résumés des épisodes (clé = identifiant TVmaze).
type Traductions struct {
	Serie    string            `json:"serie"`
	Episodes map[string]string `json:"episodes"`
}

var traductions Traductions

// chargerTraductions lit le fichier au démarrage.
// S'il manque ou est illisible, le site fonctionne quand même avec les textes de l'API.
func chargerTraductions() {
	contenu, err := os.ReadFile(fichierTraductions)
	if err == nil {
		err = json.Unmarshal(contenu, &traductions)
	}
	if err != nil {
		fmt.Println("Attention : traductions indisponibles, les résumés resteront en anglais :", err)
		return
	}
	fmt.Println(len(traductions.Episodes), "résumés traduits en français chargés")
}

var genresFR = map[string]string{
	"Drama":    "Drame",
	"Crime":    "Policier",
	"Thriller": "Thriller",
}

var paysFR = map[string]string{
	"United States":  "États-Unis",
	"United Kingdom": "Royaume-Uni",
	"Denmark":        "Danemark",
	"Canada":         "Canada",
	"Mexico":         "Mexique",
}

// traduire renvoie la traduction si elle existe, sinon le texte d'origine.
func traduire(dico map[string]string, texte string) string {
	if t, ok := dico[texte]; ok {
		return t
	}
	return texte
}
