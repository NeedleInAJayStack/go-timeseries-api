package main

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ServerConfig struct {
	// Auth
	authenticator        authenticator
	jwtSecret            string
	tokenDurationSeconds int

	// Stores
	historyStore historyStore
	recStore     recStore
	currentStore currentStore
}

func NewServer(serverConfig ServerConfig) (http.Handler, error) {
	// Register paths
	server := http.NewServeMux()
	tokenAuth := http.NewServeMux()

	authController := authController{
		jwtSecret:            serverConfig.jwtSecret,
		tokenDurationSeconds: serverConfig.tokenDurationSeconds,
		authenticator:        serverConfig.authenticator,
	}
	hisController := hisController{store: serverConfig.historyStore}
	recController := recController{store: serverConfig.recStore}
	currentController := currentController{store: serverConfig.currentStore}

	// handleFunc is a replacement for mux.HandleFunc that adds route data to the metrics.
	handleFunc := func(mux *http.ServeMux, pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	handleFunc(server, "GET /api/auth/token", authController.getAuthToken)
	handleFunc(tokenAuth, "GET /api/recs", recController.getRecs)
	handleFunc(tokenAuth, "POST /api/recs", recController.postRecs)
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
