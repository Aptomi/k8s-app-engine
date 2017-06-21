package server

import (
	"fmt"
	"github.com/Frostman/aptomi/pkg/slinga"
	"github.com/Frostman/aptomi/pkg/slinga/visibility"
	"github.com/gorilla/handlers"
	"net/http"
	"os"
)

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./webui/favicon.ico")
}

func endpointsHandler(w http.ResponseWriter, r *http.Request) {
	// prefer explicitly passed username through query
	username := r.URL.Query().Get("username")
	if username == "" {
		username = getUsername(r)
	}

	// Load the previous usage state
	state := slinga.LoadServiceUsageState()

	filterUserID := ""
	for userID, user := range slinga.LoadUsers().Users {
		if user.Name == username {
			filterUserID = userID
		}
	}

	endpoints := state.Endpoints(filterUserID)

	writeJSON(w, endpoints)
}

func detailViewHandler(w http.ResponseWriter, r *http.Request) {
	// prefer explicitly passed username through query
	username := r.URL.Query().Get("username")
	if username == "" {
		username = getUsername(r)
	}

	state := slinga.LoadServiceUsageState()

	filterUserID := ""
	for userID, user := range slinga.LoadUsers().Users {
		if user.Name == username {
			filterUserID = userID
		}
	}

	view := visibility.NewDetails(filterUserID, slinga.LoadUsers(), state)
	writeJSON(w, view)
}

func consumerViewHandler(w http.ResponseWriter, r *http.Request) {
	state := slinga.LoadServiceUsageState()
	userID := r.URL.Query().Get("userId")
	dependencyID := r.URL.Query().Get("dependencyId")
	view := visibility.NewConsumerView(userID, dependencyID, state)
	writeJSON(w, view.GetData())
}

func serviceViewHandler(w http.ResponseWriter, r *http.Request) {
	state := slinga.LoadServiceUsageState()
	serviceName := r.URL.Query().Get("serviceName")
	view := visibility.NewServiceView(serviceName, state)
	writeJSON(w, view.GetData())
}

func globalOpsViewHandler(w http.ResponseWriter, r *http.Request) {
	state := slinga.LoadServiceUsageState()

	userID := r.URL.Query().Get("userId")
	view := visibility.NewGlobalConsumerView(userID, slinga.LoadUsers().Users, state)
	writeJSON(w, view.GetData())
}

func objectViewHandler(w http.ResponseWriter, r *http.Request) {
	state := slinga.LoadServiceUsageState()
	objectID := r.URL.Query().Get("id")
	ov := visibility.NewObjectView(objectID, state)
	writeJSON(w, ov.GetData())
}

// Serve starts http server on specified address that serves Aptomi API and WebUI
func Serve(host string, port int) {
	r := http.NewServeMux()

	r.HandleFunc("/favicon.ico", faviconHandler)

	// redirect from "/" to "/ui/"
	r.Handle("/", http.RedirectHandler("/ui/", http.StatusTemporaryRedirect))

	// serve all files from "webui" folder and require auth for everything except login.html
	r.Handle("/ui/", staticFilesHandler("/ui/", http.Dir("./webui/")))

	// serve all API endpoints at /api/ path and require auth
	r.Handle("/api/endpoints", requireAuth(endpointsHandler))
	r.Handle("/api/details", requireAuth(detailViewHandler))
	r.Handle("/api/service-view", requireAuth(serviceViewHandler))
	r.Handle("/api/consumer-view", requireAuth(consumerViewHandler))
	r.Handle("/api/globalops-view", requireAuth(globalOpsViewHandler))
	r.Handle("/api/object-view", requireAuth(objectViewHandler))

	// serve login/logout api without auth
	r.HandleFunc("/api/login", loginHandler)
	r.HandleFunc("/api/logout", logoutHandler)

	listenAddr := fmt.Sprintf("%s:%d", host, port)
	fmt.Println("Serving at", listenAddr)

	var server http.Handler = r

	server = handlers.CombinedLoggingHandler(os.Stdout, server)
	server = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(server)

	// todo better handle error returned from ListenAndServe (path to Fatal??)
	panic(http.ListenAndServe(listenAddr, server))
}
