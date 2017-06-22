package server

import (
	"net/http"
	"time"
)

func getLoggedInUserId(r *http.Request) string {
	userID := ""
	if cookie, err := r.Cookie("logUserID"); err == nil {
		userID = cookie.Value
	}
	return userID
}

func isUnauthorized(r *http.Request) bool {
	return len(getLoggedInUserId(r)) <= 0
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
	userID := r.URL.Query().Get("logUserID")
	http.SetCookie(w, &http.Cookie{Name: "logUserID", Value: userID, Path: "/"})
	handleAutoRedirect(w, r)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "logUserID", Value: "", Path: "/", Expires: time.Now().AddDate(-1, 0, 0)})
	handleAutoRedirect(w, r)
}
