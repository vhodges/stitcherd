package stitcher

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mailgun/groupcache/v2"
)

// RunServer serves up sites specified in hosts
func RunServer(listenAddress string, hostConfigFiles []string) {
	r := mux.NewRouter()
	var hosts []*Host

	cacheService := InitCache()
	defer cacheService.Shutdown(context.Background())

	for _, hostConfigFile := range hostConfigFiles {
		host, err := ReadHostHCL(hostConfigFile)

		if err != nil {
			log.Printf("Error reading config file '%s': %v", hostConfigFile, err)
			continue
		}

		maxCache := host.MaxCache
		if maxCache == 0 {
			maxCache = 1 << 20
		}

		// Create a new group cache with a max cache size of 3MB
		host.Cache = groupcache.NewGroup(host.Hostname, maxCache, groupcache.GetterFunc(

			func(ctx context.Context, id string, dest groupcache.Sink) error {
				v := ctx.Value(requestContextKey("request"))

				r, ok := v.(EndPointContextValue)

				if ok {

					content, err := r.EndPoint.Render(r.Site, r.ContextData)

					if err != nil {
						log.Printf("Error: %v - '%+v'\n", err, r.EndPoint)
						return err
					}

					ttl, err := time.ParseDuration(r.EndPoint.CacheTTL)
					if err != nil {
						log.Printf("Error parsing TTL duration: '%v' for key '%s' defaulting to a TTL of one minute\n", err, id)
						ttl = time.Minute * 1
					}

					if err := dest.SetString(content, time.Now().Add(ttl)); err != nil {
						log.Println("SetString", err)
						return err
					}
				}

				return nil
			},
		))

		hosts = append(hosts, host)
		s := r.Host(host.Hostname).Subrouter()

		// TODO Allow more flexible route definition/handling (ie Method, Protocol, etc)
		for _, endPoint := range host.Routes {
			// TODO Static hosting should just be a route/handler
			endPoint.Init()
			s.HandleFunc(endPoint.Route, EndPointHandler(host, endPoint))
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
		Handler: handlers.RecoveryHandler()(r), // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	WaitForSignal(srv)
}
