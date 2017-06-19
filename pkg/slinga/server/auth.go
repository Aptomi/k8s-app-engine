package server

import (
	"net/http"
	"time"
)

func getUsername(r *http.Request) string {
	username := r.Header.Get("username")
	if username == "" {
		if cookie, err := r.Cookie("username"); err == nil {
			username = cookie.Value
		}
	}
	return username
}

func isUnauthorized(r *http.Request) bool {
	return getUsername(r) == ""
}

func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isUnauthorized(r) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	http.SetCookie(w, &http.Cookie{Name: "username", Value: username, Path: "/"})

	handleAutoRedirect(w, r)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// == delete cookie
	http.SetCookie(w, &http.Cookie{Name: "username", Value: "", Path: "/", Expires: time.Now().AddDate(-1, 0, 0)})

	handleAutoRedirect(w, r)
}
