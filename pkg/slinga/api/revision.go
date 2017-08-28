package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/registry"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)

type RevisionHandler struct {
	registry *registry.Registry
}

func (h *RevisionHandler) handleGetPolicy(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rev, key, ns := p.ByName("rev"), p.ByName("key"), p.ByName("ns")

	fmt.Printf("[handleGetPolicy] rev: %s, key: %s, ns: %s\n", rev, key, ns)

	if len(rev) == 0 {
		// todo(slukjanov): fail better
		panic("Revision should be specified (0 for current)")
	}

	if (len(key) == 0) && (len(ns) == 0) {
		// get full policy from specific revision
	}

	if (len(key) > 0) && (len(ns) > 0) {
		// todo(slukjanov): unreachable, better failing
		panic("Only one of key or namespace could specified")
	}

	if len(key) != 0 {
		// get by key from specific revision
	}

	if len(ns) != 0 {
		// get all by ns from specific revision
	}
}

func (h *RevisionHandler) handleNewRevision(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(fmt.Sprintf("Error while reading bytes from request Body: %s", err))
	}

	fmt.Println(string(body))

	// objects, err := h.registry.Codec.UnmarshalOneOrMany(body)
	// initialize and resolve new revision here from current policy + objects
}

func Serve(router *httprouter.Router, reg *registry.Registry) {
	h := RevisionHandler{reg}

	router.GET("/api/v1/revision/:rev/policy", h.handleGetPolicy)               // get full policy from specific revision
	router.GET("/api/v1/revision/:rev/policy/key/:key", h.handleGetPolicy)      // get by key from specific revision
	router.GET("/api/v1/revision/:rev/policy/namespace/:ns", h.handleGetPolicy) // get policy for namespace from specific revision

	router.POST("/api/v1/revision", h.handleNewRevision)
}
