package stitcher

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// ContentHandler uses the Source to render content
func ContentHandler(site *Host, route Route) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		route.ContentHandler(site, w, r)
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
			if r.Source != nil {
				r.Init()
				s.HandleFunc(r.Path, ContentHandler(host, r))
			} else if r.StaticPath != nil {
				s.PathPrefix(r.Path).Handler(http.StripPrefix(r.Path, http.FileServer(http.Dir(r.StaticPath.Directory))))
			}
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
