package stitcher

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/mailgun/groupcache/v2"
)

// Host represents a single VHOSTed site
type Host struct {
	Hostname string `hcl:"hostname"`

	Routes []Route `hcl:"route,block"`

	Cache    *groupcache.Group
	MaxCache int64 `hcl:"max_cache,optional"`

	Router *mux.Router
}

// Init handles host specific initialization
func (host *Host) Init() {

	maxCache := host.MaxCache
	if maxCache == 0 {
		maxCache = 1 << 20
	}

	host.Router = mux.NewRouter()

	// Use a previously created group ie if reloading 
	host.Cache = groupcache.GetGroup(host.Hostname)

	// or Create a new group cache with a max cache size of 3MB
	if host.Cache == nil {
		host.Cache = groupcache.NewGroup(host.Hostname, maxCache, groupcache.GetterFunc(

			func(ctx context.Context, id string, dest groupcache.Sink) error {
				v := ctx.Value(requestContextKey("request"))

				r, ok := v.(ContentContextValue)

				if ok {

					content, err := r.Content.Render(r.Site, r.ContextData)

					if err != nil {
						log.Printf("Error: %v - '%+v'\n", err, r.Content)
						return err
					}

					ttl, err := time.ParseDuration(r.Content.CacheTTL)
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
	}
}

// ContentHandler uses the Source to render content
func ContentHandler(site *Host, route Route) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		route.ContentHandler(site, w, r)
	}
}

func NewHostFromFile(file string) (*Host, error) {

	host, err := ReadHostHCL(file)

	if err != nil {
		log.Printf("Error reading config file '%s': %v", file, err)
		return  nil, err
	}

	host.Init()

	// TODO Allow more flexible route definition/handling (ie Method, Protocol, etc)
	for _, r := range host.Routes {
		if r.Source != nil {
			r.Init()
			host.Router.HandleFunc(r.Path, ContentHandler(host, r))
		} else if r.StaticPath != nil {
			host.Router.PathPrefix(r.Path).Handler(http.StripPrefix(r.Path, http.FileServer(http.Dir(r.StaticPath.Directory))))
		}
	}

	return host, nil
}
