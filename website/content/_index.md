+++
title = "Stitcherd"
sort_by = "weight"
+++

<div id="homepage">

Static sites are great. They're fast to serve, there's less to to go wrong and are hostable 
pretty much anywhere and any how you would like to.  But sometimes,  just sometimes there is
some page or some part of every page that has dynamic or even personalized content on it. 
Stitcherd is for those times.

## Features

### Current

  * Multiple vhosts
  * CSS Selector page assembly
  * Static Content catch all (but... will allow proxy fallback soon)
  * Simple cache controls per endpoint/route (more types coming soon - ie etag, last modified etc.)
  * Go templates (with HTML and JSON Data retrieval) for endpoints

### Coming Soon

  * Bot detection (>800 known bots)
  * Both General and (Bot == true) rate limiting (Site wide and per route)
  * (Optional) Sessions 
  * Proxy for fallback (eg / to some CMS) and for routes (eg /blog/ proxied to Wordpress)
  * Static content routes
  * Site Authentication (OAUTH/SAML end point config, basic auth?  Builtin user/password (agencies?)). Authenticate routes?
  * HMAC auth support for backend ends
  * More control over endpoint request (ACTION/Verb, protocol, headers, cookies, form vars, etc )

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
./stitcherd --host demo/site.hcl
http://localhost:3000/
```

The home page at https://stitcherd.vhodges.dev/ will also serve as an example (soon).

* It pulls the index page content from the Github README.md (ie this file)
* It pulls a list of recent posts to the Announcment mailing list from sourcehut
* Source (and generated static site - via zola - are checked into source control under website)

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


</div>
