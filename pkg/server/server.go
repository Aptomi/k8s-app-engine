package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/api/middleware"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/external/secrets"
	"github.com/Aptomi/aptomi/pkg/external/users"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/plugin/fake"
	"github.com/Aptomi/aptomi/pkg/plugin/helm"
	"github.com/Aptomi/aptomi/pkg/plugin/k8s"
	"github.com/Aptomi/aptomi/pkg/plugin/k8sraw"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"github.com/Aptomi/aptomi/pkg/runtime/store/core"
	"github.com/Aptomi/aptomi/pkg/runtime/store/generic/bolt"
	"github.com/Aptomi/aptomi/pkg/server/ui"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"runtime/trace"
	"syscall"
	"time"
)

// Server is Aptomi server. It serves UI front-end, API calls, as well as does policy resolution & continuous state enforcement
type Server struct {
	cfg              *config.Server
	backgroundErrors chan string

	externalData *external.Data
	store        store.Core

	httpServer *http.Server

	runDesiredStateEnforcement    chan bool
	desiredStateEnforcementIdx    uint
	enforcerPluginRegistryFactory plugin.RegistryFactory

	runActualStateUpdate         chan bool
	actualStateUpdateIdx         uint
	updaterPluginRegistryFactory plugin.RegistryFactory
}

// NewServer creates a new Aptomi Server
func NewServer(cfg *config.Server) *Server {
	s := &Server{
		cfg:                        cfg,
		backgroundErrors:           make(chan string),
		runDesiredStateEnforcement: make(chan bool, 2048),
		runActualStateUpdate:       make(chan bool, 2048),
	}

	return s
}

// Start initializes Aptomi server, starts API & UI processing, and as well as runs the required background jobs for
// continuous policy resolution and state enforcement
func (server *Server) Start() {
	// Init server
	server.initProfiling()
	server.initStore()
	server.initExternalData()
	server.initPluginRegistryFactory()
	server.initPolicyOnFirstRun()

	// Start API, UI, Enforcer and ActualStateUpdater
	server.startHTTPServer()
	server.startDesiredStateEnforcer()
	server.startActualStateUpdater()

	// Wait for jobs to complete (it essentially hangs forever)
	server.wait()
}

func (server *Server) initPolicyOnFirstRun() {
	policy, _, err := server.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while getting latest policy: %s", err))
	}

	// if policy does not exist, let's create the first version (it should be created here, before we start the server)
	if policy == nil {
		log.Infof("Policy not found in the store (likely, it's a first run of Aptomi server). Creating empty policy")
		initErr := server.store.InitPolicy()
		if initErr != nil {
			panic(fmt.Sprintf("error while creating empty policy: %s", initErr))
		}
	}
}

func (server *Server) initExternalData() {
	var userLoaders []users.UserLoader
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

func (server *Server) initProfiling() {
	if len(server.cfg.Profile.CPU) > 0 {
		// initiate CPU profiler
		f, err := os.Create(server.cfg.Profile.CPU)
		if err != nil {
			panic(fmt.Sprintf("can't create file to write CPU profiling information: %s", server.cfg.Profile.CPU))
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic(fmt.Sprintf("can't start CPU profiling: %s", err))
		}

		// CPU profiler needs to be stopped when server exits
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go func() {
			for sig := range c {
				log.Printf("captured %v, stopping CPU profiler", sig)
				pprof.StopCPUProfile()
			}
		}()
	}

	if len(server.cfg.Profile.Trace) > 0 {
		// start tracing
		f, err := os.Create(server.cfg.Profile.Trace)
		if err != nil {
			panic(fmt.Sprintf("can't create file to write tracing information: %s", server.cfg.Profile.Trace))
		}
		err = trace.Start(f)
		if err != nil {
			panic(fmt.Sprintf("can't start tracing: %s", err))
		}

		// Tracing needs to be stopped when server exits
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		go func() {
			for sig := range c {
				log.Printf("captured %v, stopping tracing", sig)
				trace.Stop()
			}
		}()
	}
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

func (server *Server) initPluginRegistryFactory() {
	fn := func(noop bool, noopSleep time.Duration) func() plugin.Registry {
		return func() plugin.Registry {
			clusterTypes := make(map[string]plugin.ClusterPluginConstructor)
			codeTypes := make(map[string]map[string]plugin.CodePluginConstructor)

			if !noop {
				clusterTypes["kubernetes"] = func(cluster *lang.Cluster, cfg config.Plugins) (plugin.ClusterPlugin, error) {
					return k8s.New(cluster, cfg)
				}

				codeTypes["kubernetes"] = make(map[string]plugin.CodePluginConstructor)
				codeTypes["kubernetes"]["helm"] = func(cluster plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
					return helm.New(cluster, cfg)
				}
				codeTypes["kubernetes"]["raw"] = func(cluster plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
					return k8sraw.New(cluster, cfg)
				}
			} else {
				clusterTypes["kubernetes"] = func(cluster *lang.Cluster, cfg config.Plugins) (plugin.ClusterPlugin, error) {
					return fake.NewNoOpClusterPlugin(noopSleep), nil
				}

				codeTypes["kubernetes"] = make(map[string]plugin.CodePluginConstructor)
				codeTypes["kubernetes"]["helm"] = func(cluster plugin.ClusterPlugin, cfg config.Plugins) (plugin.CodePlugin, error) {
					return fake.NewNoOpCodePlugin(noopSleep), nil
				}
			}

			return plugin.NewRegistry(server.cfg.Plugins, clusterTypes, codeTypes)
		}
	}

	server.enforcerPluginRegistryFactory = fn(server.cfg.Enforcer.Noop, server.cfg.Enforcer.NoopSleep)
	server.updaterPluginRegistryFactory = fn(server.cfg.Updater.Noop, server.cfg.Updater.NoopSleep)
}

func (server *Server) startHTTPServer() {
	router := httprouter.New()

	if len(server.cfg.Auth.Secret) == 0 {
		// todo better handle it
		// set some default insecure secret
		server.cfg.Auth.Secret = "Shahsh4e cohp8aeT Ifaic3ah ohs4eiSh vee7Qua7 eiCh2iLo eiroh3Ie oeg2ruPu"
		log.Warnf("The auth.secret not specified in config, using insecure default one")
	}

	api.Serve(router, server.store, server.externalData, server.enforcerPluginRegistryFactory, server.cfg.Auth.Secret, server.cfg.GetLogLevel(), server.runDesiredStateEnforcement)
	server.serveUI(router)

	var handler http.Handler = router

	// todo write to logrus
	handler = handlers.CombinedLoggingHandler(os.Stdout, handler) // todo(slukjanov): make it at least somehow configurable - for example, select file to write to with rotation
	handler = middleware.NewPanicHandler(handler)
	// todo(slukjanov): add configurable handlers.ProxyHeaders to f behind the nginx or any other proxy
	// todo(slukjanov): add compression handler and compress by default in client

	server.httpServer = &http.Server{
		Handler:      handler,
		Addr:         server.cfg.API.ListenAddr(),
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	// Start HTTP server
	server.runInBackground("HTTP Server / API", true, func() {
		panic(server.httpServer.ListenAndServe())
	})
}

func (server *Server) serveUI(router *httprouter.Router) {
	if !server.cfg.UI.Enable {
		log.Infof("UI isn't enabled. UI will not be served")
		return
	}

	fileServer := http.FileServer(ui.HTTP)

	router.GET("/", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		req.URL.Path = "/"
		fileServer.ServeHTTP(w, req)
	})

	router.GET("/index.html", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		req.URL.Path = "/"
		fileServer.ServeHTTP(w, req)
	})

	router.GET("/static/*filepath", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		req.URL.Path = "/static/" + ps.ByName("filepath")
		fileServer.ServeHTTP(w, req)
	})
}

func (server *Server) startDesiredStateEnforcer() {
	// Start policy enforcement job
	if !server.cfg.Enforcer.Disabled {
		server.runInBackground("Desired State Enforcer", true, func() {
			panic(server.desiredStateEnforceLoop())
		})
	}
}

func (server *Server) startActualStateUpdater() {
	// Start policy enforcement job
	if !server.cfg.Updater.Disabled {
		server.runInBackground("Actual State Updater", true, func() {
			panic(server.actualStateUpdateLoop())
		})
	}
}
