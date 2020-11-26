package stitcher

import (
	"context"
	"log"
	"time"

	"github.com/mailgun/groupcache/v2"
)

// Host represents a single VHOSTed site
type Host struct {
	Hostname     string `hcl:"hostname"`
	DocumentRoot string `hcl:"documentroot,optional"`

	Routes []Route `hcl:"route,block"`

	Cache    *groupcache.Group
	MaxCache int64 `hcl:"max_cache,optional"`
}

// Init handles host specific initialization
func (host *Host) Init() {

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
}
