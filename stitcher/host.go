package stitcher

import "github.com/mailgun/groupcache/v2"

// Host represents a single VHOSTed site
type Host struct {
	Hostname     string `hcl:"hostname"`
	DocumentRoot string `hcl:"documentroot,optional"`

	Routes []Route `hcl:"route,block"`

	Cache    *groupcache.Group
	MaxCache int64 `hcl:"max_cache,optional"`
}
