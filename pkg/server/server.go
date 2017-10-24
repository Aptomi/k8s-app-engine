package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/external/secrets"
	"github.com/Aptomi/aptomi/pkg/external/users"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/object/codec"
	"github.com/Aptomi/aptomi/pkg/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/object/store/bolt"
	"github.com/Aptomi/aptomi/pkg/server/store"
	"github.com/Aptomi/aptomi/pkg/webui"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"time"
)

// Server is Aptomi server. It serves API and UI calls, as well as does policy enforcement
type Server struct {
	cfg              *config.Server
	backgroundErrors chan string
	catalog          *object.Catalog
	codec            codec.MarshallerUnmarshaller

	externalData *external.Data
	store        store.ServerStore
	httpServer   *http.Server
}

// NewServer creates a new Aptomi Server
func NewServer(cfg *config.Server) *Server {
	s := &Server{
		cfg:              cfg,
		backgroundErrors: make(chan string),
	}

	s.catalog = object.NewCatalog().Append(lang.Objects...).Append(store.Objects...)
	s.codec = yaml.NewCodec(s.catalog)

	return s
}

// Start initializes Aptomi server, starts serving API/UI, and as well as runs the required background jobs for actual policy enforcement
func (s *Server) Start() {
	s.initStore()
	s.initExternalData()

	// Register UI and API handlers
	s.initHTTPServer()

	// Start HTTP server
	s.runInBackground("HTTP Server", true, func() {
		panic(s.httpServer.ListenAndServe())
	})

	// Start policy enforcement job
	if !s.cfg.Enforcer.Disabled {
		s.runInBackground("Policy Enforcer", true, func() {
			panic(s.enforceLoop())
		})
	}

	s.wait()
}

func (s *Server) initExternalData() {
	userLoaders := []users.UserLoader{}
	for _, ldap := range s.cfg.Users.LDAP {
		userLoaders = append(userLoaders, users.NewUserLoaderFromLDAP(ldap, s.cfg.DomainAdminOverrides))
	}
	for _, file := range s.cfg.Users.File {
		userLoaders = append(userLoaders, users.NewUserLoaderFromFile(file, s.cfg.DomainAdminOverrides))
	}
	s.externalData = external.NewData(
		users.NewUserLoaderMultipleSources(userLoaders),
		secrets.NewSecretLoaderFromDir(s.cfg.SecretsDir),
	)
}

func (s *Server) initStore() {
	b := bolt.NewBoltStore(s.catalog, s.codec)
	err := b.Open(s.cfg.DB.Connection)
	if err != nil {
		panic(fmt.Sprintf("Can't open object store: %s", err))
	}
	s.store = store.New(b)
}

func (s *Server) initHTTPServer() {
	router := httprouter.New()

	api.Serve(router, s.store, s.externalData)
	webui.Serve(router)

	var handler http.Handler = router

	// todo write to logrus
	handler = handlers.CombinedLoggingHandler(os.Stdout, handler) // todo(slukjanov): make it at least somehow configurable - for example, select file to write to with rotation
	handler = newPanicHandler(handler)
	// todo(slukjanov): add configurable handlers.ProxyHeaders to f behind the nginx or any other proxy
	// todo(slukjanov): add compression handler and compress by default in client

	s.httpServer = &http.Server{
		Handler:      handler,
		Addr:         s.cfg.API.ListenAddr(),
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
}
