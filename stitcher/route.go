package stitcher

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/x-way/crawlerdetect"
	"golang.org/x/time/rate"
)

// Route returns content for a given path
type Route struct {
	Path string // Respond to requests at this path

	RespondWith string 

	// RespondWith "fragemented_page' renders this
	Page     *FragmentedPage
	
	// ResponeWith 'static_files' servers this local directory to server
	StaticPath string   
	
	// TODO 
	// RedirectTo string  // TODO interpolate
	// RedirectStatus // Eg 302 (Found) or 302 (Moved permanently)

	// ResponseString string // TODO Interpolate
	// ResponseCode   string

	// ProxyHost string
	// ProxyString??? string

	// Rate limiter
	MaxRate       float64
	AllowBurst    int
	BotMaxRate    float64
	BotAllowBurst int

	normalLimiter *rate.Limiter // Really only makes sense/applies to Route
	botLimiter    *rate.Limiter
}

// Init creates runtime objects for the end point
func (route *Route) Init(host *Host) {

	if route.MaxRate > 0 && route.AllowBurst > 0 {
		route.normalLimiter = rate.NewLimiter(rate.Limit(route.MaxRate),
			route.AllowBurst)

		log.Printf("Added rate limiter for '%s' Rate: %f Burst: %d\n",
			route.Path, rate.Limit(route.MaxRate), route.AllowBurst)
	}

	if route.BotMaxRate > 0 && route.BotAllowBurst > 0 {
		route.botLimiter = rate.NewLimiter(rate.Limit(route.BotMaxRate),
			route.BotAllowBurst)

		log.Printf("Added bot limiter for '%s' Rate: %f Burst: %d\n",
			route.Path, rate.Limit(route.BotMaxRate), route.BotAllowBurst)
	}

	// Add the handler for the route
	switch route.RespondWith {
	case "fragmented_page":
		host.Router.HandleFunc(route.Path, FragmentedPageHandler(host, *route))
	case "static_content":
		host.Router.PathPrefix(route.Path).Handler(http.StripPrefix(route.Path, http.FileServer(http.Dir(route.StaticPath))))
	case "redirect":
	case "proxy":
	}

}

// Throttling returns true if the current request is to be rate limited.
func (route *Route) Throttling(r *http.Request) bool {

	var limiter = route.normalLimiter

	if route.botLimiter != nil && crawlerdetect.IsCrawler(r.UserAgent()) {
		log.Println("Crawler detected")
		limiter = route.botLimiter
	}

	if limiter != nil {
		return !limiter.Allow()
	}

	return false
}

func (route *Route) nextRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// FragmentedPageHandler Renders the route.Page
func (route *Route) FragmentedPageHandler(site *Host, w http.ResponseWriter, r *http.Request) {
	var err error
	start := time.Now()

	var fetchContext map[string]interface{} = make(map[string]interface{})

	if route.Throttling(r) {
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return
	}

	// TODO Better request tracing... (context.Context too?)
	fetchContext["_requestId"] = route.nextRequestID()

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

	content := route.Page.Render(site, fetchContext)

	if err != nil {
		log.Printf("Error from endpoint '%s': %v", route.Path, err)
		fmt.Fprintln(w, "")
	} else {
		fmt.Fprintln(w, content)
	}

	elapsed := time.Since(start)
	log.Println(fetchContext["_requestId"], r.Method, r.URL.Path, r.Proto, elapsed)
}

// FragmentedPageHandler uses the Source to render content
func FragmentedPageHandler(site *Host, route Route) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		route.FragmentedPageHandler(site, w, r)
	}
}
