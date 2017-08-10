package server

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/engine"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/server/visibility"
	"github.com/gorilla/handlers"
	"net/http"
	"os"
	"time"
)

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./webui/favicon.ico")
}

func endpointsHandler(w http.ResponseWriter, r *http.Request) {
	// Load the previous usage state
	userLoader := NewAptomiUserLoader()
	state := engine.LoadServiceUsageState(userLoader)
	endpoints := visibility.Endpoints(getLoggedInUserID(r), state)

	writeJSON(w, endpoints)
}

func detailViewHandler(w http.ResponseWriter, r *http.Request) {
	userLoader := NewAptomiUserLoader()
	state := engine.LoadServiceUsageState(userLoader)
	userID := getLoggedInUserID(r)
	view := visibility.NewDetails(userID, state)
	writeJSON(w, view)
}

func consumerViewHandler(w http.ResponseWriter, r *http.Request) {
	userLoader := NewAptomiUserLoader()
	state := engine.LoadServiceUsageState(userLoader)
	userID := r.URL.Query().Get("userId")
	dependencyID := r.URL.Query().Get("dependencyId")
	view := visibility.NewConsumerView(userID, dependencyID, state)
	writeJSON(w, view.GetData())
}

func serviceViewHandler(w http.ResponseWriter, r *http.Request) {
	userLoader := NewAptomiUserLoader()
	state := engine.LoadServiceUsageState(userLoader)
	serviceName := r.URL.Query().Get("serviceName")
	view := visibility.NewServiceView(serviceName, state)
	writeJSON(w, view.GetData())
}

func globalOpsViewHandler(w http.ResponseWriter, r *http.Request) {
	userLoader := NewAptomiUserLoader()
	state := engine.LoadServiceUsageState(userLoader)
	userID := r.URL.Query().Get("userId")
	dependencyID := r.URL.Query().Get("dependencyId")
	view := visibility.NewConsumerView(userID, dependencyID, state)
	writeJSON(w, view.GetData())
}

func objectViewHandler(w http.ResponseWriter, r *http.Request) {
	userLoader := NewAptomiUserLoader()
	state := engine.LoadServiceUsageState(userLoader)
	objectID := r.URL.Query().Get("id")
	view := visibility.NewObjectView(objectID, state)
	writeJSON(w, view.GetData())
}

func summaryViewHandler(w http.ResponseWriter, r *http.Request) {
	userLoader := NewAptomiUserLoader()
	state := engine.LoadServiceUsageState(userLoader)
	userID := getLoggedInUserID(r)
	view := visibility.NewSummaryView(userID, state)
	writeJSON(w, view.GetData())
}

func timelineViewHandler(w http.ResponseWriter, r *http.Request) {
	userLoader := NewAptomiUserLoader()
	states := engine.LoadServiceUsageStatesAll(userLoader)
	userID := getLoggedInUserID(r)
	view := visibility.NewTimelineView(userID, states)
	writeJSON(w, view.GetData())
}

// Serve starts http server on specified address that serves Aptomi API and WebUI
func Serve(host string, port int) {
	r := http.NewServeMux()

	r.HandleFunc("/favicon.ico", faviconHandler)

	// redirect from "/" to "/ui/"
	r.Handle("/", http.RedirectHandler("/ui/", http.StatusTemporaryRedirect))

	// serve all files from "webui" folder and require auth for everything except login.html
	r.Handle("/ui/", publicFilesHandler("/ui/", http.Dir("./webui/")))
	r.Handle("/run/", runFilesHandler("/run/", http.Dir(GetAptomiBaseDir())))

	// serve all API endpoints at /api/ path and require auth
	r.Handle("/api/endpoints", requireAuth(endpointsHandler))
	r.Handle("/api/details", requireAuth(detailViewHandler))
	r.Handle("/api/service-view", requireAuth(serviceViewHandler))
	r.Handle("/api/consumer-view", requireAuth(consumerViewHandler))
	r.Handle("/api/globalops-view", requireAuth(globalOpsViewHandler))
	r.Handle("/api/object-view", requireAuth(objectViewHandler))
	r.Handle("/api/summary-view", requireAuth(summaryViewHandler))
	r.Handle("/api/timeline-view", requireAuth(timelineViewHandler))

	// serve login/logout api without auth
	r.HandleFunc("/api/login", loginHandler)
	r.HandleFunc("/api/logout", logoutHandler)

	listenAddr := fmt.Sprintf("%s:%d", host, port)
	fmt.Println("Serving at", listenAddr)

	var h http.Handler = r

	h = handlers.CombinedLoggingHandler(os.Stdout, h)
	h = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(h)

	srv := &http.Server{
		Handler:      h,
		Addr:         listenAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	panic(srv.ListenAndServe())
}
