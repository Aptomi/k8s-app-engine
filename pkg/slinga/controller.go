package slinga

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/version"
	"github.com/Aptomi/aptomi/pkg/slinga/webui"
	"github.com/gorilla/handlers"
	"net/http"
	"os"
	"time"
)

// Init http server with all handlers
// 1. version handler
// 2. api handler
// 3. event logs api (should it be separate?)
// 4. webui handler (serve static files)

// Start some go routines
// 1. users fetcher
// 2. revisions applier

// Some notes
// * in dev mode serve webui files from specified directory, otherwise serve from inside of binary

func Serve() {
	host, port := "", 8080 // todo(slukjanov): load this properties from config
	listenAddr := fmt.Sprintf("%s:%d", host, port)

	router := http.NewServeMux()

	version.Serve(router)
	webui.Serve(router)

	var handler http.Handler = router

	handler = handlers.CombinedLoggingHandler(os.Stdout, handler) // todo(slukjanov): make it at least somehow configurable - for example, select file to write to with rotation
	handler = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(handler)
	// todo(slukjanov): add configurable handlers.ProxyHeaders to run behind the nginx or any other proxy

	srv := &http.Server{
		Handler:      handler,
		Addr:         listenAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	panic(srv.ListenAndServe())
}
