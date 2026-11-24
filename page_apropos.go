package main

// Page à-propos : GET /a-propos

import "net/http"

func aPropos(w http.ResponseWriter, r *http.Request) {
	afficher(w, http.StatusOK, "apropos", struct{ Base }{nouvelleBase("À propos", "apropos")})
}
