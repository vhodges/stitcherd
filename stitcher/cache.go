package stitcher

import (
	"log"
	"net/http"

	"github.com/mailgun/groupcache/v2"
)

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
