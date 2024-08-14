package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authController struct {
	jwtSecret            string
	tokenDurationSeconds int
	username             string
	password             string
}

// GET /auth/token
// Requires basic auth
func (a authController) getAuthToken(w http.ResponseWriter, r *http.Request) {
	reqUsername, reqPassword, _ := r.BasicAuth()
	if reqUsername != a.username || reqPassword != a.password {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"subject":  "go",
		"username": a.username,
		"exp":      time.Now().Unix() + int64(a.tokenDurationSeconds),
	})
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
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
}

func tokenAuthMiddleware(jwtSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		for _, headerValue := range r.Header["Authorization"] {
			if strings.HasPrefix(headerValue, "Bearer ") {
				tokenString, _ = strings.CutPrefix(headerValue, "Bearer ")
			}
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			log.Printf("JWT parsing failed: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("JWT claims failed: %s", err)
		}
		exp, _ := claims.GetExpirationTime()
		if exp.Before(time.Now()) {
			log.Printf("JWT expired at: %s", exp)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Pass to the next handler
		next.ServeHTTP(w, r)
	})
}
