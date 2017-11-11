package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/api/middleware"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/external/secrets"
	"github.com/Aptomi/aptomi/pkg/external/users"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"github.com/Aptomi/aptomi/pkg/runtime/store/core"
	"github.com/Aptomi/aptomi/pkg/runtime/store/generic/bolt"
	"github.com/Aptomi/aptomi/pkg/webui"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"net/http"
	"os"
	"time"
)

// Server is Aptomi server. It serves API and UI calls, as well as does policy enforcement
type Server struct {
	cfg              *config.Server
	backgroundErrors chan string

	externalData *external.Data
	store        store.Core
	httpServer   *http.Server
}

// NewServer creates a new Aptomi Server
func NewServer(cfg *config.Server) *Server {
	s := &Server{
		cfg:              cfg,
		backgroundErrors: make(chan string),
	}

	return s
}

// Start initializes Aptomi server, starts serving API/UI, and as well as runs the required background jobs for actual policy enforcement
func (server *Server) Start() {
	server.initStore()
	server.initExternalData()

	// Register UI and API handlers
	server.initHTTPServer()

	// Start HTTP server
	server.runInBackground("HTTP Server", true, func() {
		panic(server.httpServer.ListenAndServe())
	})

	// Start policy enforcement job
	if !server.cfg.Enforcer.Disabled {
		server.runInBackground("Policy Enforcer", true, func() {
			panic(server.enforceLoop())
		})
	}

	server.wait()
}

func (server *Server) initExternalData() {
	userLoaders := []users.UserLoader{}
	for _, ldap := range server.cfg.Users.LDAP {
		userLoaders = append(userLoaders, users.NewUserLoaderFromLDAP(ldap, server.cfg.DomainAdminOverrides))
	}
	for _, file := range server.cfg.Users.File {
		userLoaders = append(userLoaders, users.NewUserLoaderFromFile(file, server.cfg.DomainAdminOverrides))
	}
	server.externalData = external.NewData(
		users.NewUserLoaderMultipleSources(userLoaders),
		secrets.NewSecretLoaderFromDir(server.cfg.SecretsDir),
	)
}

func (server *Server) initStore() {
	registry := runtime.NewRegistry().Append(store.Objects...)
	b := bolt.NewGenericStore(registry)
	err := b.Open(server.cfg.DB)
	if err != nil {
		panic(fmt.Sprintf("Can't open object store: %s", err))
	}
	server.store = core.NewStore(b)
}

func (server *Server) initHTTPServer() {
	router := httprouter.New()

	api.Serve(router, server.store, server.externalData)
	webui.Serve(router)

	var handler http.Handler = router

	// todo write to logrus
	handler = handlers.CombinedLoggingHandler(os.Stdout, handler) // todo(slukjanov): make it at least somehow configurable - for example, select file to write to with rotation
	handler = cors.Default().Handler(handler)
	handler = middleware.NewPanicHandler(handler)
	// todo(slukjanov): add configurable handlers.ProxyHeaders to f behind the nginx or any other proxy
	// todo(slukjanov): add compression handler and compress by default in client

	server.httpServer = &http.Server{
		Handler:      handler,
		Addr:         server.cfg.API.ListenAddr(),
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
}
