package main

import (
	"net/http"

	"gorm.io/gorm"
)

type ServerConfig struct {
	// Auth
	username             string
	password             string
	jwtSecret            string
	tokenDurationSeconds int

	// Database
	db            *gorm.DB
	dbAutoMigrate bool
}

func NewServer(serverConfig ServerConfig) (*http.ServeMux, error) {

	if serverConfig.dbAutoMigrate {
		err := serverConfig.db.AutoMigrate(&his{}, &rec{})
		if err != nil {
			return nil, err
		}
	}

	// Register paths
	server := http.NewServeMux()
	authController := authController{
		jwtSecret:            serverConfig.jwtSecret,
		tokenDurationSeconds: serverConfig.tokenDurationSeconds,
		username:             serverConfig.username,
		password:             serverConfig.password,
	}
	server.HandleFunc("GET /auth/token", authController.getAuthToken)

	tokenAuth := http.NewServeMux()

	hisController := hisController{db: serverConfig.db}
	tokenAuth.HandleFunc("GET /his/{pointId}", hisController.getHis)
	tokenAuth.HandleFunc("POST /his/{pointId}", hisController.postHis)
	tokenAuth.HandleFunc("DELETE /his/{pointId}", hisController.deleteHis)
	server.Handle("/his/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	recController := recController{db: serverConfig.db}
	tokenAuth.HandleFunc("GET /recs", recController.getRecs)
	tokenAuth.HandleFunc("POST /recs", recController.postRecs)
	server.Handle("/recs", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	tokenAuth.HandleFunc("GET /recs/tag/{tag}", recController.getRecsByTag)
	tokenAuth.HandleFunc("GET /recs/{id}", recController.getRec)
	tokenAuth.HandleFunc("PUT /recs/{id}", recController.putRec)
	tokenAuth.HandleFunc("DELETE /recs/{id}", recController.deleteRec)
	server.Handle("/recs/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	return server, nil
}