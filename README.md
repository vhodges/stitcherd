# Introduction

Static sites are great. They're fast to serve, there's less to to go wrong and are hostable 
pretty much anywhere and any how you would like to.  But sometimes,  just sometimes there is
some page or some part of every page that has dynamic or even personalized content on it. 
Stitcherd is for those times.

It is a web server that reads some source content (typically a local static file, but could come from a remote source), fetches one or more pieces of remote “dynamic” content and injects them into the source document using a css selector to find the insertion point. These can be nested. The remote dynamic content can be remote html or output of a Go template that can also process remote html and/or remote json data.

The resulting content is then served to the client. You can have multiple routes with different sets of content to be replaced (and indeed source document). In other words it does server side includes, but with css/dom manipulation and so doesn’t require special directives in the source documents.

A couple of Use cases:

* Fast e-commerce site with static product pages, but dynamic pricing/availability/promotions. Plus server side carts
* Commenting system for static blogs. JS Free
* Micro services/frontends

Yes, it’s server that has to be hosted somewhere and you need to decide if that’s okay for your use case or not, but then so are most of the alternatives, except JS of course, but same-origin, et al leads it to be harder (imo) to use than this.

## Features

### Current

  * Multiple vhosts
  * CSS Selector page assembly
  * Static Content catch all (but... will allow proxy fallback soon)
  * Simple cache controls per endpoint/route (more types coming soon - ie etag, last modified etc.)
  * Go templates (with HTML and JSON Data retrieval) for endpoints
  * Bot detection (>800 known bots)
  * Both General and (Bot == true) rate limiting (per route)
  * Static content routes
  
### Coming Soon

  * (Optional) Sessions 
  * Site Authentication (OAUTH/SAML end point config, basic auth?  Builtin user/password (agencies?)). Authenticate routes?
  * HMAC auth support for backend ends
  * More control over endpoint request (ACTION/Verb, protocol, headers, cookies, form vars, etc )
  * Proxy for fallback (eg / to some CMS) and for routes (eg /blog/ proxied to Wordpress)


# Building

Requires Go > 1.12 

Clone the repo and then a simple 

``` 
go build 
```

Docker image at some point

# Examples

There is a 'demo' folder that serves as an example/testbed

```
./stitcherd --host demo/site.json
http://localhost:3000/
```

# Prior Art and Inspiration

* Edge side includes (ESI) using Varnish https://varnish-cache.org/
* shtml files and virtual paths in nginx
* Greasemonkey: https://addons.mozilla.org/en-CA/firefox/addon/greasemonkey/ (Client side)
* Mousehole: https://github.com/whymirror/mousehole
* Jigsaw: https://www.w3.org/Jigsaw/Overview.html (I vaguely recall it could do some of this kind of thing - I could be wrong though)
* Netlify: most recently has the same idea with their Edge (Handlers) https://www.netlify.com/products/edge/edge-handlers/
* Soupault: Is the main inspiration for how stitcherd works. https://soupault.neocities.org/ 

# License

Apache 2.0

