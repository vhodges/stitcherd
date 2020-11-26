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
	Path string `hcl:",label"` // Respond to requests at this path

	Source Content `hcl:"content,block"` // URL to fetch the main source

	MaxRate       float64 `hcl:"maxrate,optional"`
	AllowBurst    int     `hcl:"burst,optional"`
	BotMaxRate    float64 `hcl:"botmaxrate,optional"`
	BotAllowBurst int     `hcl:"botburst,optional"`

	normalLimiter *rate.Limiter // Really only makes sense/applies to Route
	botLimiter    *rate.Limiter

	//Options hcl.Body `hcl:",remain"`
	//FetchData map[string]string `hcl:"rules"`
}

// Init creates runtime objects for the end point
func (route *Route) Init() {

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

// Handler uses the Source to render content
func (route *Route) Handler(site *Host, w http.ResponseWriter, r *http.Request) {

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

	content, err := route.Source.Fetch(site, fetchContext)

	if err != nil {
		log.Printf("Error from endpoint '%s': %v", route.Path, err)
		fmt.Fprintln(w, "")
	} else {
		fmt.Fprintln(w, content)
	}

	elapsed := time.Since(start)
	log.Println(fetchContext["_requestId"], r.Method, r.URL.Path, r.Proto, elapsed)
}
