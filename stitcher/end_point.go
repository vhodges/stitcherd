package stitcher

import (
	"context"
	"html/template"
	"io/ioutil"

	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/PuerkitoBio/goquery"
	"github.com/bradfitz/iter"
	"github.com/mailgun/groupcache/v2"
	"github.com/valyala/fasttemplate"
	"github.com/x-way/crawlerdetect"
	"golang.org/x/time/rate"
)

// TODO The endpoint code needs to be re-done... getting crufty.
type requestContextKey string

// EndPoint itself is a RenderableContent
type EndPoint struct {
	Route string `hcl:",label"` // Respond to requests at this path

	Source   string `hcl:"source,optional"` // URL to fetch the main source
	Selector string `hcl:"select,optional"` // CSS Selector to extract content from - optional

	Merges []Merge `hcl:"replace,block"` // May be empty

	CacheKey string `hcl:"cache,optional"`
	CacheTTL string `hcl:"ttl,optional"`

	Template string `hcl:"template,optional"` // Go template source -- URL
	JSON     string `hcl:"json,optional"`     // Used by templates to retrieve data -- URL

	MaxRate       float64 `hcl:"maxrate,optional"`
	AllowBurst    int     `hcl:"burst,optional"`
	BotMaxRate    float64 `hcl:"botmaxrate,optional"`
	BotAllowBurst int     `hcl:"botburst,optional"`

	parsedTemplate *template.Template // Cached/preparsed template... parsed on first use.
	normalLimiter  *rate.Limiter      // Really only makes sense/applies to Route
	botLimiter     *rate.Limiter

	//Options hcl.Body `hcl:",remain"`
	//FetchData map[string]string `hcl:"rules"`
}

// Init creates runtime objects for the end point
func (endpoint *EndPoint) Init() {

	if endpoint.MaxRate > 0 && endpoint.AllowBurst > 0 {
		endpoint.normalLimiter = rate.NewLimiter(rate.Limit(endpoint.MaxRate),
			endpoint.AllowBurst)

		log.Printf("Added rate limiter for '%s' Rate: %f Burst: %d\n",
			endpoint.Route, rate.Limit(endpoint.MaxRate), endpoint.AllowBurst)
	}

	if endpoint.BotMaxRate > 0 && endpoint.BotAllowBurst > 0 {
		endpoint.botLimiter = rate.NewLimiter(rate.Limit(endpoint.BotMaxRate),
			endpoint.BotAllowBurst)

		log.Printf("Added bot limiter for '%s' Rate: %f Burst: %d\n",
			endpoint.Route, rate.Limit(endpoint.BotMaxRate), endpoint.BotAllowBurst)
	}

	// TODO Set up parsed template here (if Template is present)
}

// Throttling returns true if the current request is to be rate limited.
func (endpoint *EndPoint) Throttling(r *http.Request) bool {

	var limiter = endpoint.normalLimiter

	if endpoint.botLimiter != nil && crawlerdetect.IsCrawler(r.UserAgent()) {
		log.Println("Crawler detected")
		limiter = endpoint.botLimiter
	}

	if limiter != nil {
		return !limiter.Allow()
	}

	return false
}

// Content returns the rendered content (or from endpoint if endpoint is configured for the end point)
func (endpoint *EndPoint) Content(site *Host, contextdata map[string]interface{}) (string, error) {

	if endpoint.Caching() {

		var content string

		var contextvalue = EndPointContextValue{Site: site, EndPoint: endpoint, ContextData: contextdata}
		ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), requestContextKey("request"), contextvalue),
			time.Millisecond*500)

		defer cancel()

		if err := site.Cache.Get(ctx, endpoint.InterpolatedCacheKey(contextdata), groupcache.StringSink(&content)); err != nil {
			log.Printf("Error getting from cache: %v\n", err)
			return "", err
		}

		return content, nil
	}

	return endpoint.Render(site, contextdata)
}

// Render loads content from SourceURI and merges an fragements
// into the resulting document and returns the string representation
func (endpoint *EndPoint) Render(site *Host, contextdata map[string]interface{}) (string, error) {

	var renderedContent string

	content, err2 := endpoint.Fetch(contextdata)

	if err2 != nil {
		log.Println(err2)
		return "", err2
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))

	if err != nil {
		log.Println(err)
		return "", err
	}

	// Might be empty but otherwise, get and merge the content for each one
	// TODO Add context handling and cancelation
	for _, merge := range endpoint.Merges {

		// Retrieve fragement.Source content
		fragment, _ := merge.Content.Content(site, contextdata)

		// Find the insertion point in sourceFile at merge.InsertAt
		insertSelection := doc.Find(merge.At) // Potentially costly, look into caching the source and insert selection points!?!

		// Insert the extracted content
		insertSelection.ReplaceWithHtml(fragment)
	}

	// By having the selector we can treat endpoints as a component
	if endpoint.Selector != "" {
		// Get content at Selector
		renderedContent, err = doc.Find(endpoint.Selector).Html()
	} else {
		renderedContent, err = doc.Html()
	}

	return renderedContent, err
}

// Fetch returns the content
func (endpoint *EndPoint) Fetch(contextdata map[string]interface{}) (string, error) {
	return endpoint.fetcher(contextdata).Fetch()
}

// Factory method to return a fetcher for the end point
func (endpoint *EndPoint) fetcher(contextdata map[string]interface{}) DocumentFetcher {
	var fetcher DocumentFetcher
	fetcher = &StringFetcher{Body: ""} // Default to empty string

	if endpoint.Source != "" {
		t := fasttemplate.New(endpoint.Source, "{{", "}}")
		s := t.ExecuteString(contextdata)

		u, err := url.Parse(s)
		if err != nil {
			// TODO Log the error
			return &StringFetcher{Body: ""}
		}

		switch u.Scheme {
		case "": // Pathname eg: "path/to/file" or "/abs/path/to/file"
			fetcher = &FileFetcher{Path: u.Path}
		case "string": // Inline string data  eg "string:This is my String"
			fetcher = &StringFetcher{Body: u.Opaque}
		default: // Any other supported uri/url (ie http/https)
			fetcher = &URIFetcher{URI: s}
		}
	}

	if endpoint.Template != "" {

		if true /*endpoint.parsedTemplate == nil*/ {
			templateBytes, _ := ioutil.ReadFile(endpoint.Template)
			templateContents := string(templateBytes)

			funcs := sprig.GenericFuncMap()
			funcs["N"] = iter.N
			funcs["unescape"] = unescape

			endpoint.parsedTemplate = template.Must(template.New(endpoint.Template).Funcs(template.FuncMap(funcs)).Parse(templateContents))
		}

		// interpolate the path for the any JSON source
		t := fasttemplate.New(endpoint.JSON, "{{", "}}")
		json := t.ExecuteString(contextdata)

		return &RenderedTemplateFetcher{Template: endpoint.parsedTemplate,
			DataURL:        json,
			SourceFetcher:  fetcher,
			RequestContext: contextdata,
		}
	}

	// Fallback to empty string
	return fetcher
}

func unescape(s string) template.HTML {
	return template.HTML(s)
}

// EndPointContextValue is passed via Context.WithValue() to the endpoint Getter Func
type EndPointContextValue struct {
	Site        *Host
	EndPoint    *EndPoint
	ContextData map[string]interface{}
}

// Caching returns true if we are to use the endpoint
func (endpoint *EndPoint) Caching() bool {
	return endpoint.CacheKey != ""
}

// InterpolatedCacheKey returns the interpolated endpoint key
func (endpoint *EndPoint) InterpolatedCacheKey(contextData map[string]interface{}) string {
	t := fasttemplate.New(endpoint.CacheKey, "{{", "}}")
	return t.ExecuteString(contextData)

}
