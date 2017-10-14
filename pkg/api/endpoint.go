package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	"github.com/Aptomi/aptomi/pkg/server/store"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func ServeEndpoints(router *httprouter.Router, store store.ServerStore) {
	router.GET("/api/v1/endpoints", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		endpoints := make(map[string]map[string]string)
		actualState, err := store.GetActualState()
		if err != nil {
			panic("Can't load actual state")
		}
		for _, instance := range actualState.ComponentInstanceMap {
			if len(instance.Endpoints) > 0 {
				endpoints[instance.GetName()] = instance.Endpoints
			}
		}

		data := yaml.SerializeObject(endpoints)

		// todo bad logging
		fmt.Println("Response: " + string(data))

		_, err = fmt.Fprint(w, string(data))
		if err != nil {
			panic(fmt.Sprintf("Error while writing response bytes: %s", err))
		}
	})
}
