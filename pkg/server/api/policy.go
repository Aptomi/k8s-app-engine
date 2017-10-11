package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/object/codec"
	"github.com/Aptomi/aptomi/pkg/server/store"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)

// PolicyAPI is a an object which allows to retrieve and update policy via API
type PolicyAPI struct {
	store store.ServerStore
	codec codec.MarshallerUnmarshaller
}

func (a *PolicyAPI) handleGetPolicy(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rev, ns := p.ByName("rev"), p.ByName("ns")

	if len(rev) == 0 {
		rev = "0" // latest revision
	}

	fmt.Printf("[handleGetPolicy] rev: %s, ns: %s\n", rev, ns)

	if len(ns) != 0 {
		// todo get all by ns from specific revision
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
		// todo it should be badrequest
		panic(fmt.Sprintf("Error unmarshaling policy update request: %s", err))
	}

	_, policyData, err := a.store.UpdatePolicy(objects)
	if err != nil {
		panic(fmt.Sprintf("Error while updating policy: %s", err))
	}

	//if updated {
	data, err := a.codec.MarshalOne(policyData)
	if err != nil {
		panic(fmt.Sprintf("Error marshaling updated policy: %s", err))
	}

	// todo bad logging
	fmt.Println("Response: " + string(data))

	_, err = fmt.Fprint(w, string(data))
	if err != nil {
		panic(fmt.Sprintf("Error while writing response bytes: %s", err))
	}
	//} else { // nothing changed
	//	w.WriteHeader(http.StatusBadRequest)
	//	 todo write some error back to client
	//}
}

// ServePolicy registers policy processing handlers in API
func ServePolicy(router *httprouter.Router, store store.ServerStore, cod codec.MarshallerUnmarshaller) {
	h := PolicyAPI{store, cod}

	router.GET("/api/v1/policy", h.handleGetPolicy)
	router.GET("/api/v1/policy/:rev", h.handleGetPolicy)
	router.GET("/api/v1/policy/:rev/namespace/:ns", h.handleGetPolicy)

	router.POST("/api/v1/policy", h.handlePolicyUpdate)
}
