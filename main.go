package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	// Fonctions utilisables dans les templates.
	fonctions := template.FuncMap{
	}
	// Les templates sont chargés une seule fois au démarrage : une erreur est détectée tout de suite.
	var err error
	templates, err = template.New("").Funcs(fonctions).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Impossible de charger les templates : ", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", pageIntrouvable)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Serveur lancé sur http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
