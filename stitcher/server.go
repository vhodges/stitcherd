package stitcher

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// RouteHandler uses the Source to render content
func RouteHandler(site *Host, route Route) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		route.Handler(site, w, r)
	}
}

// RunServer serves up sites specified in hosts
func RunServer(listenAddress string, hostConfigFiles []string) {
	router := mux.NewRouter()
	var hosts []*Host

	cacheService := InitCache()
	defer cacheService.Shutdown(context.Background())

	for _, hostConfigFile := range hostConfigFiles {
		host, err := ReadHostHCL(hostConfigFile)

		if err != nil {
			log.Printf("Error reading config file '%s': %v", hostConfigFile, err)
			continue
		}

		host.Init()
		hosts = append(hosts, host)
		s := router.Host(host.Hostname).Subrouter()

		// TODO Allow more flexible route definition/handling (ie Method, Protocol, etc)
		for _, r := range host.Routes {
			// TODO Static hosting should just be a route/handler
			r.Init()
			s.HandleFunc(r.Path, RouteHandler(host, r))
		}

		// If a document root was supplied, set up a default route for static content
		// See Previous TODO
		if len(host.DocumentRoot) > 0 {
			s.PathPrefix("/").Handler(http.FileServer(http.Dir(host.DocumentRoot)))
		}
	}

	srv := &http.Server{
		Addr: listenAddress,

		// Good practice to set timeouts to avoid Slowloris attacks.  TODO Timeout config
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,

		//		Handler: handlers.CombinedLoggingHandler(os.Stderr, handlers.RecoveryHandler()(r)), // Pass our instance of gorilla/mux in.
		Handler: handlers.RecoveryHandler()(router), // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	WaitForSignal(srv)
}
