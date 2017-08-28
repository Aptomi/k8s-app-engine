package webui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	pathlib "path"
	"strings"
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

func stringHasAnyPrefix(str string, prefix ...string) bool {
	for _, prefix := range prefix {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}

	return false
}

func publicFilesHandler(path string, root http.FileSystem) http.Handler {
	return http.StripPrefix(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "login.html" && !stringHasAnyPrefix(pathlib.Clean(r.URL.Path), "public/") {
			if isUnauthorized(r) {
				http.Redirect(w, r, "/ui/login.html", http.StatusTemporaryRedirect)
				return
			}
		}

		fileServer := http.FileServer(root)
		fileServer.ServeHTTP(w, r)
	}))
}

func runFilesHandler(path string, root http.FileSystem) http.Handler {
	return http.StripPrefix(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := pathlib.Clean(r.URL.Path)
		if isUnauthorized(r) || !stringHasAnyPrefix(p, "run-") || !strings.HasSuffix(p, ".png") {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
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

func writeJSON(w io.Writer, obj interface{}) {
	res, err := json.Marshal(obj)
	if err != nil {
		fmt.Fprintf(w, "Unable to serialize object: %v", err)
	} else {
		fmt.Fprint(w, string(res))
	}
}
