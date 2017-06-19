package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// todo enforce login/logout to work only through POST

//func requireMethod(method string, handler http.HandlerFunc) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		if r.Method != method {
//			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
//			return
//		}
//
//		handler(w, r)
//	}
//}

//func requirePost(handler http.HandlerFunc) http.HandlerFunc {
//	return requireMethod(http.MethodPost, handler)
//}

func staticFilesHandler(path string, root http.FileSystem) http.Handler {
	return http.StripPrefix(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "login.html" {
			if isUnauthorized(r) {
				http.Redirect(w, r, "/ui/login.html", http.StatusTemporaryRedirect)
				return
			}
		}

		fileServer := http.FileServer(root)
		fileServer.ServeHTTP(w, r)
	}))
}

func handleAutoRedirect(w http.ResponseWriter, r *http.Request) {
	if redirect := r.URL.Query().Get("redirect"); redirect != "" {
		http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
	}
}

func writeJSON(w http.ResponseWriter, obj interface{}) {
	// todo handle errors
	res, _ := json.Marshal(obj)
	fmt.Fprint(w, string(res))
}
