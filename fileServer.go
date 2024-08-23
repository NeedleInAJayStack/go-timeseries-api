package main

import "net/http"

// fileServerWithFallback returns a file server that falls back to a specific file when no matches are found.
// This is useful when a frontend app is using a browser-router.
func fileServerWithFallback(dir http.Dir, fallbackFile string) http.Handler {
	fileserver := http.FileServer(dir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := dir.Open(r.URL.Path)
		if err == nil {
			fileserver.ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, fallbackFile)
		}
	})
}
