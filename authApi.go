package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func registerAuth(jwtSecret string, username string, password string) {

	// GET /auth/token
	// Requires basic auth
	http.HandleFunc("GET /auth/token", func(w http.ResponseWriter, r *http.Request) {
		reqUsername, reqPassword, _ := r.BasicAuth()
		if reqUsername != username || reqPassword != password {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"subject":  "go",
			"username": username,
		})
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			log.Printf("Unable to sign JWT: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		clientToken := clientToken{
			Token: tokenString,
		}

		httpJson, err := json.Marshal(clientToken)
		if err != nil {
			log.Printf("Cannot encode response JSON")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(httpJson)
	})
}
