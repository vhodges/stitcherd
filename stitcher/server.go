package stitcher

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Stitcherd struct {

	ListenAddress    string
	WorkingDirectory string
	AdminHostName    string
  AdminEnabled     bool

	hosts            map[string]*Host
	adminRouter      *mux.Router
}

// Performs initialization
func (stitcherd *Stitcherd) Init() *Stitcherd {

	stitcherd.hosts = make(map[string]*Host)

	if stitcherd.AdminEnabled {
		stitcherd.adminRouter = mux.NewRouter().Host(stitcherd.AdminHostName).Subrouter()
		stitcherd.adminRouter.HandleFunc("/hosts/load/{filename:.*}", stitcherd.AdminHandler())
	}

	return stitcherd
}

func (stitcherd *Stitcherd) ServeHTTP(w http.ResponseWriter, request *http.Request) {

	log.Printf("Request for Host: '%s'", request.Host)

	host := stitcherd.hosts[request.Host]

	if host != nil {
		host.Router.ServeHTTP(w, request)
	} else if stitcherd.adminRouter != nil {
		stitcherd.adminRouter.ServeHTTP(w, request)
	} else {
		w.WriteHeader(http.StatusNotFound)
    w.Write([]byte("404 - Not found\n"))
	}
}

// RunStictcherd serves up sites specified in hosts
func (stitcherd *Stitcherd) Run(hostConfigFiles []string) {
	log.Printf("Start - admin hostname: '%s', working directory: '%s'\n", 
		stitcherd.AdminHostName, stitcherd.WorkingDirectory)

	for _, file := range hostConfigFiles {
		host, err := NewHostFromFile(file) 
		if err == nil {
			stitcherd.SetHost(host)
		}
	}

	cacheService := InitCache()
	defer cacheService.Shutdown(context.Background())

	srv := &http.Server{
		Addr: stitcherd.ListenAddress,

		// Good practice to set timeouts to avoid Slowloris attacks.  TODO Timeout config
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,

		//		Stictcherd: Stictcherds.CombinedLoggingStictcherd(os.Stderr, Stictcherds.RecoveryStictcherd()(r)), // Pass our instance of gorilla/mux in.
		Handler: handlers.RecoveryHandler()(stitcherd),
	}

	// Run our Stictcherd in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	WaitForSignal(srv)
}

// Add or replace a host for the Stictcherd
func (stitcherd *Stitcherd) SetHost(host *Host) {
	stitcherd.hosts[host.Hostname] = host
}

// Returns an AdminHandler func
func (stitcherd *Stitcherd) AdminHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		filename := vars["filename"]

		log.Printf("AdminHandler: New Host File?: '%s'\n", filename)

		host, err := NewHostFromFile(filename) 

		if err == nil {
			stitcherd.SetHost(host)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Filename: %v OK\n", vars["filename"])
			log.Printf("AdminHandler: (re)Loaded %s\n", filename)
  	} else {
			log.Printf("AdminHandler: Error Loading %s, %v\n", filename, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "filename: %v ERROR '%v'\n", vars["filename"], err)
		}
	}
}

