package stitcher

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mailgun/groupcache/v2"
)

// EndPointHandler uses the EndPoint to render content
func EndPointHandler(site *Host, endpoint EndPoint) func(http.ResponseWriter, *http.Request) {

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var fetchContext map[string]interface{} = make(map[string]interface{})

		// TODO Better request tracing... (context.Context too?)
		fetchContext["_requestId"] = nextRequestID()

		for key, element := range mux.Vars(r) {
			fetchContext[key] = element
		}

		// TODO these probably need to be escaped
		fetchContext["requestPath"] = r.URL.Path
		fetchContext["queryString"] = r.URL.RawQuery

		for key, element := range r.URL.Query() {
			if len(element) == 0 {
				fetchContext[key] = ""
			} else {
				fetchContext[key] = element[0]
			}
		}

		// TODO Add Headers? Cookies? to fetchContext

		content, err := endpoint.Content(site, fetchContext)

		if err != nil {
			log.Printf("Error from endpoint '%s': %v", endpoint.Route, err)
			fmt.Fprintln(w, "")
		} else {
			fmt.Fprintln(w, content)
		}

		elapsed := time.Since(start)
		log.Println(fetchContext["_requestId"], r.Method, r.URL.Path, r.Proto, elapsed)
	}
}

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
			//      ideally allowing merges and content injection too
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

// InitCache sets up the cache service
func InitCache() *http.Server {

	// Keep track of peers in our cluster and add our instance to the pool `http://localhost:8080`
	// TODO pool/service config
	pool := groupcache.NewHTTPPoolOpts("http://localhost:8080", &groupcache.HTTPPoolOptions{})

	// TODO Config and Add more peers to the cluster
	//pool.Set("http://peer1:8080", "http://peer2:8080")
	// TODO Dynamim peer addition/removal

	server := http.Server{
		Addr:    "localhost:8080", // TODO Cache service address
		Handler: pool,
	}

	// Start a HTTP server to listen for peer requests from the groupcache
	go func() {
		log.Printf("Cache Server Running....\n")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	return &server
}

// WaitForSignal blocks until SIGINT arrives.
func WaitForSignal(srv *http.Server) {
	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	var wait time.Duration
	wait = time.Second * 15

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
