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
	tokenAuth := http.NewServeMux()

	authController := authController{
		jwtSecret:            serverConfig.jwtSecret,
		tokenDurationSeconds: serverConfig.tokenDurationSeconds,
		username:             serverConfig.username,
		password:             serverConfig.password,
	}
	hisController := hisController{store: newGormHistoryStore(serverConfig.db)}
	recController := recController{store: newGormRecStore(serverConfig.db)}
	currentController := currentController{store: newInMemoryCurrentStore()}

	// All are deprecated. Instead use /api/*
	server.HandleFunc("GET /auth/token", authController.getAuthToken)
	tokenAuth.HandleFunc("GET /his/{pointId}", hisController.getHis)       // Deprecated
	tokenAuth.HandleFunc("POST /his/{pointId}", hisController.postHis)     // Deprecated
	tokenAuth.HandleFunc("DELETE /his/{pointId}", hisController.deleteHis) // Deprecated
	tokenAuth.HandleFunc("GET /recs", recController.getRecs)
	tokenAuth.HandleFunc("POST /recs", recController.postRecs)
	tokenAuth.HandleFunc("GET /recs/tag/siteMeter", recController.getRecsByTag) // Deprecated. Use /recs?tag=""
	tokenAuth.HandleFunc("GET /recs/{id}", recController.getRec)
	tokenAuth.HandleFunc("PUT /recs/{id}", recController.putRec)
	tokenAuth.HandleFunc("DELETE /recs/{id}", recController.deleteRec)
	tokenAuth.HandleFunc("GET /recs/{pointId}/history", hisController.getHis)
	tokenAuth.HandleFunc("POST /recs/{pointId}/history", hisController.postHis)
	tokenAuth.HandleFunc("DELETE /recs/{pointId}/history", hisController.deleteHis)
	tokenAuth.HandleFunc("GET /recs/{pointId}/current", currentController.getCurrent)
	tokenAuth.HandleFunc("POST /recs/{pointId}/current", currentController.postCurrent)
	server.Handle("/his/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))
	server.Handle("/recs", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))
	server.Handle("/recs/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	server.HandleFunc("GET /api/auth/token", authController.getAuthToken)
	tokenAuth.HandleFunc("GET /api/his/{pointId}", hisController.getHis)       // Deprecated
	tokenAuth.HandleFunc("POST /api/his/{pointId}", hisController.postHis)     // Deprecated
	tokenAuth.HandleFunc("DELETE /api/his/{pointId}", hisController.deleteHis) // Deprecated
	tokenAuth.HandleFunc("GET /api/recs", recController.getRecs)
	tokenAuth.HandleFunc("POST /api/recs", recController.postRecs)
	tokenAuth.HandleFunc("GET /api/recs/tag/siteMeter", recController.getRecsByTag) // Deprecated. Use /recs?tag=""
	tokenAuth.HandleFunc("GET /api/recs/{id}", recController.getRec)
	tokenAuth.HandleFunc("PUT /api/recs/{id}", recController.putRec)
	tokenAuth.HandleFunc("DELETE /api/recs/{id}", recController.deleteRec)
	tokenAuth.HandleFunc("GET /api/recs/{pointId}/history", hisController.getHis)
	tokenAuth.HandleFunc("POST /api/recs/{pointId}/history", hisController.postHis)
	tokenAuth.HandleFunc("DELETE /api/recs/{pointId}/history", hisController.deleteHis)
	tokenAuth.HandleFunc("GET /api/recs/{pointId}/current", currentController.getCurrent)
	tokenAuth.HandleFunc("POST /api/recs/{pointId}/current", currentController.postCurrent)
	server.Handle("/api/his/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))
	server.Handle("/api/recs", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))
	server.Handle("/api/recs/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	// Catch all others with public files. Not found fallback is app index for browser router.
	server.Handle("/app/", fileServerWithFallback(http.Dir("./public"), "./public/app/index.html"))

	return server, nil
}
