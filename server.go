package main

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

func NewServer(serverConfig ServerConfig) (http.Handler, error) {

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

	// handleFunc is a replacement for mux.HandleFunc that adds route data to the metrics.
	handleFunc := func(mux *http.ServeMux, pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// All are deprecated. Instead use /api/*
	handleFunc(server, "GET /auth/token", authController.getAuthToken)
	handleFunc(tokenAuth, "GET /his/{pointId}", hisController.getHis)       // Deprecated
	handleFunc(tokenAuth, "POST /his/{pointId}", hisController.postHis)     // Deprecated
	handleFunc(tokenAuth, "DELETE /his/{pointId}", hisController.deleteHis) // Deprecated
	handleFunc(tokenAuth, "GET /recs", recController.getRecs)
	handleFunc(tokenAuth, "POST /recs", recController.postRecs)
	handleFunc(tokenAuth, "GET /recs/tag/siteMeter", recController.getRecsByTag) // Deprecated. Use /recs?tag=""
	handleFunc(tokenAuth, "GET /recs/{id}", recController.getRec)
	handleFunc(tokenAuth, "PUT /recs/{id}", recController.putRec)
	handleFunc(tokenAuth, "DELETE /recs/{id}", recController.deleteRec)
	handleFunc(tokenAuth, "GET /recs/{pointId}/history", hisController.getHis)
	handleFunc(tokenAuth, "POST /recs/{pointId}/history", hisController.postHis)
	handleFunc(tokenAuth, "DELETE /recs/{pointId}/history", hisController.deleteHis)
	handleFunc(tokenAuth, "GET /recs/{pointId}/current", currentController.getCurrent)
	handleFunc(tokenAuth, "POST /recs/{pointId}/current", currentController.postCurrent)
	server.Handle("/his/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))
	server.Handle("/recs", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))
	server.Handle("/recs/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	handleFunc(server, "GET /api/auth/token", authController.getAuthToken)
	handleFunc(tokenAuth, "GET /api/his/{pointId}", hisController.getHis)       // Deprecated
	handleFunc(tokenAuth, "POST /api/his/{pointId}", hisController.postHis)     // Deprecated
	handleFunc(tokenAuth, "DELETE /api/his/{pointId}", hisController.deleteHis) // Deprecated
	handleFunc(tokenAuth, "GET /api/recs", recController.getRecs)
	handleFunc(tokenAuth, "POST /api/recs", recController.postRecs)
	handleFunc(tokenAuth, "GET /api/recs/tag/siteMeter", recController.getRecsByTag) // Deprecated. Use /recs?tag=""
	handleFunc(tokenAuth, "GET /api/recs/{id}", recController.getRec)
	handleFunc(tokenAuth, "PUT /api/recs/{id}", recController.putRec)
	handleFunc(tokenAuth, "DELETE /api/recs/{id}", recController.deleteRec)
	handleFunc(tokenAuth, "GET /api/recs/{pointId}/history", hisController.getHis)
	handleFunc(tokenAuth, "POST /api/recs/{pointId}/history", hisController.postHis)
	handleFunc(tokenAuth, "DELETE /api/recs/{pointId}/history", hisController.deleteHis)
	handleFunc(tokenAuth, "GET /api/recs/{pointId}/current", currentController.getCurrent)
	handleFunc(tokenAuth, "POST /api/recs/{pointId}/current", currentController.postCurrent)
	server.Handle("/api/his/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))
	server.Handle("/api/recs", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))
	server.Handle("/api/recs/", tokenAuthMiddleware(serverConfig.jwtSecret, tokenAuth))

	// Catch all others with public files. Not found fallback is app index for browser router.
	server.Handle("/app/", fileServerWithFallback(http.Dir("./public"), "./public/app/index.html"))

	// Observability
	observed := otelhttp.NewHandler(server, "/")

	return observed, nil
}
