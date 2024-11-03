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
	tokenAuth.HandleFunc("GET /his/{pointId}", hisController.getHis)       // Deprecated
	tokenAuth.HandleFunc("POST /his/{pointId}", hisController.postHis)     // Deprecated
	tokenAuth.HandleFunc("DELETE /his/{pointId}", hisController.deleteHis) // Deprecated
	server.Handle("/his/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	recController := recController{db: serverConfig.db}
	tokenAuth.HandleFunc("GET /recs", recController.getRecs)
	tokenAuth.HandleFunc("POST /recs", recController.postRecs)
	server.Handle("/recs", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	tokenAuth.HandleFunc("GET /recs/tag/siteMeter", recController.getRecsByTag) // Deprecated. Use /recs?tag=""
	tokenAuth.HandleFunc("GET /recs/{id}", recController.getRec)
	tokenAuth.HandleFunc("PUT /recs/{id}", recController.putRec)
	tokenAuth.HandleFunc("DELETE /recs/{id}", recController.deleteRec)

	tokenAuth.HandleFunc("GET /recs/{pointId}/history", hisController.getHis)
	tokenAuth.HandleFunc("POST /recs/{pointId}/history", hisController.postHis)
	tokenAuth.HandleFunc("DELETE /recs/{pointId}/history", hisController.getHis)
	server.Handle("/recs/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	// Catch all others with public files. Not found fallback is app index for browser router.
	server.Handle("/", fileServerWithFallback(http.Dir("./public"), "./public/index.html"))

	return server, nil
}
