package stitcher

import (
	"regexp"
	
	"github.com/gorilla/mux"

	"github.com/mailgun/groupcache/v2"
)

// Host represents a single VHOSTed site
type Host struct {
	Hostname string 

	Routes []Route 

	Cache    *groupcache.Group
	MaxCache int64

	Router *mux.Router

	hostPattern *regexp.Regexp
}

// Init handles host specific initialization
func (host *Host) Init() {

	// Treat HostName as a regular expression 
	host.hostPattern = regexp.MustCompile(host.Hostname)

	maxCache := host.MaxCache
	if maxCache == 0 {
		maxCache = 1 << 24
	}

	host.Router = mux.NewRouter()

	// Use a previously created group ie if reloading 
	host.Cache = groupcache.GetGroup(host.Hostname)

	// or Create a new group cache with a max cache size of 3MB
	if host.Cache == nil {
		host.Cache = groupcache.NewGroup(host.Hostname, 
			maxCache, groupcache.GetterFunc(FillFragmentCache))
	}

	for _, r := range host.Routes {
		r.Init(host)
	}	
}

func (host *Host) Match(hostname string) bool {
	return host.hostPattern.MatchString(hostname)
}

