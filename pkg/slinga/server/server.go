package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/server/controller"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store/bolt"
	"github.com/Aptomi/aptomi/pkg/slinga/server/api"
	"github.com/Aptomi/aptomi/pkg/slinga/version"
	"github.com/Aptomi/aptomi/pkg/slinga/webui"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"time"
)

// Init http server with all handlers
// * version handler
// * api handler
// * event logs api (should it be separate?)
// * webui handler (serve static files)

// Start some go routines
// * users fetcher
// * revisions applier

// Some notes
// * in dev mode serve webui files from specified directory, otherwise serve from inside of binary

func Start(config *viper.Viper) {
	catalog := object.NewObjectCatalog(lang.ServiceObject, lang.ContextObject, lang.ClusterObject, lang.RuleObject, lang.DependencyObject)
	cod := yaml.NewCodec(catalog)

	db := initStore(config, catalog, cod)
	policyCtl := initPolicyController(config, db)

	srv := initHTTPServer(config, policyCtl, cod)

	panic(srv.ListenAndServe())
}

func initStore(config *viper.Viper, catalog *object.Catalog, cod codec.MarshalUnmarshaler) store.ObjectStore {
	//todo(slukjanov): init bolt store, take file path from config
	b := bolt.NewBoltStore(catalog, cod)
	//todo load from config
	err := b.Open("/tmp/aptomi.bolt")
	if err != nil {
		panic(fmt.Sprintf("Can't open object store: %s", err))
	}
	return b
}

func initPolicyController(config *viper.Viper, store store.ObjectStore) controller.PolicyController {
	return controller.NewPolicyController(store)
}

func initHTTPServer(config *viper.Viper, policyCtl controller.PolicyController, cod codec.MarshalUnmarshaler) *http.Server {
	host, port := "", 8080 // todo(slukjanov): load this properties from config
	listenAddr := fmt.Sprintf("%s:%d", host, port)

	router := httprouter.New()

	version.Serve(router)
	api.Serve(router, policyCtl, cod)
	webui.Serve(router)

	var handler http.Handler = router

	handler = handlers.CombinedLoggingHandler(os.Stdout, handler) // todo(slukjanov): make it at least somehow configurable - for example, select file to write to with rotation
	handler = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(handler)
	// todo(slukjanov): add configurable handlers.ProxyHeaders to run behind the nginx or any other proxy
	// todo(slukjanov): add compression handler and compress by default in client

	return &http.Server{
		Handler:      handler,
		Addr:         listenAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
}
