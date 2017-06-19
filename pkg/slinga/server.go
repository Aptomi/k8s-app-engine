package slinga

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func serveEndpoints(w http.ResponseWriter, r *http.Request) {
	// Load the previous usage state
	state := LoadServiceUsageState()

	endpoints := state.Endpoints()

	// todo handle errors
	res, _ := json.Marshal(endpoints)
	fmt.Fprint(w, string(res))
}

func Serve(host string, port int) {
	// redirect from "/" to "/ui/"
	http.Handle("/", http.RedirectHandler("/ui/", http.StatusPermanentRedirect))

	// serve all files from "webui" folder as is at /ui/ path
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("./webui"))))

	// serve all API endpoints at /api/ path
	http.HandleFunc("/api/endpoints", serveEndpoints)

	//http.HandleFunc()

	fmt.Println("Serving")
	// todo handle error returned from ListenAndServe (path to Fatal??)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}
