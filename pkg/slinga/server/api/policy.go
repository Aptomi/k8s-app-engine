package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/server/controller"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)

type PolicyAPI struct {
	ctl   controller.PolicyController
	codec codec.MarshalUnmarshaler
}

func (a *PolicyAPI) handleGetPolicy(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rev, ns := p.ByName("rev"), p.ByName("ns")

	if len(rev) == 0 {
		rev = "0" // latest revision
	}

	fmt.Printf("[handleGetPolicy] rev: %s, ns: %s\n", rev, ns)

	if len(ns) != 0 {
		// get all by ns from specific revision
	}
}

func (a *PolicyAPI) handlePolicyUpdate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(fmt.Sprintf("Error while reading bytes from request Body: %s", err))
	}

	// todo remove bad logging
	fmt.Println(string(body))

	objects, err := a.codec.UnmarshalOneOrMany(body)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshaling policy update request: %s", err))
	}
	policy, err := a.ctl.UpdatePolicy(objects)
	if err != nil {
		panic(fmt.Sprintf("Error while updating policy: %s", err))
	}

	// todo remove bad logging
	fmt.Println(policy)

	// temp send back received data (to impl some table output on client side)
	// todo send full updated policy
	_, err = fmt.Fprint(w, string(body))
	if err != nil {
		panic(fmt.Sprintf("Error while writing response bytes: %s", err))
	}
}

func Serve(router *httprouter.Router, ctl controller.PolicyController, cod codec.MarshalUnmarshaler) {
	h := PolicyAPI{ctl, cod}

	router.GET("/api/v1/policy", h.handleGetPolicy)
	router.GET("/api/v1/policy/:rev", h.handleGetPolicy)
	router.GET("/api/v1/policy/:rev/namespace/:ns", h.handleGetPolicy)

	router.POST("/api/v1/policy", h.handlePolicyUpdate)
}
