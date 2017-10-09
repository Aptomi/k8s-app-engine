package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store/bolt"
	"github.com/Aptomi/aptomi/pkg/slinga/server/api"
	"github.com/Aptomi/aptomi/pkg/slinga/server/store"
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

// Server is a HTTP server which serves API and UI
type Server struct {
	config           *viper.Viper
	backgroundErrors chan string
	catalog          *object.Catalog
	codec            codec.MarshallerUnmarshaller

	store      store.ServerStore
	httpServer *http.Server
}

// NewServer creates a new HTTP Server
func NewServer(config *viper.Viper) *Server {
	s := &Server{
		config:           config,
		backgroundErrors: make(chan string),
	}

	s.catalog = object.NewCatalog().Append(lang.Objects...).Append(store.Objects...)
	s.codec = yaml.NewCodec(s.catalog)

	return s
}

// Start makes HTTP server start serving content
func (s *Server) Start() {
	s.initStore()
	s.initHTTPServer()

	s.runInBackground("HTTP Server", true, func() {
		panic(s.httpServer.ListenAndServe())
	})

	s.runInBackground("Policy Enforcer", true, func() {
		NewEnforcer(s.store).Enforce()
	})

	s.wait()
}

func (s *Server) initStore() {
	//todo(slukjanov): init bolt store, take file path from config
	b := bolt.NewBoltStore(s.catalog, s.codec)
	//todo load from config
	err := b.Open("/tmp/aptomi.bolt")
	if err != nil {
		panic(fmt.Sprintf("Can't open object store: %s", err))
	}
	s.store = store.New(b)
}

func (s *Server) initHTTPServer() {
	host, port := "", 8080 // todo(slukjanov): load this properties from config
	listenAddr := fmt.Sprintf("%s:%d", host, port)

	router := httprouter.New()

	version.Serve(router)
	api.ServePolicy(router, s.store, s.codec)
	api.ServeAdminStore(router, s.store)
	webui.Serve(router)

	var handler http.Handler = router

	handler = handlers.CombinedLoggingHandler(os.Stdout, handler) // todo(slukjanov): make it at least somehow configurable - for example, select file to write to with rotation
	handler = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(handler)
	// todo(slukjanov): add configurable handlers.ProxyHeaders to f behind the nginx or any other proxy
	// todo(slukjanov): add compression handler and compress by default in client

	s.httpServer = &http.Server{
		Handler:      handler,
		Addr:         listenAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
}
