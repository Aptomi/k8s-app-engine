package server

import (
	"fmt"
	"github.com/Frostman/aptomi/pkg/slinga"
	"net/http"
)

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./webui/favicon.ico")
}

func endpointsHandler(w http.ResponseWriter, r *http.Request) {
	// Load the previous usage state
	state := slinga.LoadServiceUsageState()

	endpoints := state.Endpoints()

	writeJSON(w, endpoints)
}

// Serve starts http server on specified address that serves Aptomi API and WebUI
func Serve(host string, port int) {
	http.HandleFunc("/favicon.ico", faviconHandler)

	// redirect from "/" to "/ui/"
	http.Handle("/", http.RedirectHandler("/ui/", http.StatusPermanentRedirect))

	// serve all files from "webui" folder and require auth for everything except login.html
	http.Handle("/ui/", staticFilesHandler("/ui/", http.Dir("./webui/")))

	// serve all API endpoints at /api/ path and require auth
	http.Handle("/api/endpoints", requireAuth(endpointsHandler))

	// serve login/logout api without auth
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", logoutHandler)

	listenAddr := fmt.Sprintf("%s:%d", host, port)
	fmt.Println("Serving at", listenAddr)
	// todo better handle error returned from ListenAndServe (path to Fatal??)
	panic(http.ListenAndServe(listenAddr, nil))
}
