package slinga

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func Endpoints(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Load the previous usage state
	state := LoadServiceUsageState()

	endpoints := state.Endpoints()

	res, _ := json.Marshal(endpoints)
	fmt.Fprint(w, string(res))
}

func Serve(host string, port int) {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)
	router.GET("/api/endpoints", Endpoints)
	router.ServeFiles("/static/*filepath", http.Dir("public/static"))

	fmt.Println("Serving")
	// todo handle error returned from ListenAndServe (path to Fatal??)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), router)
}
