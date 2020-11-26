package stitcher

import (
	"context"
	"html/template"
	"io/ioutil"

	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/PuerkitoBio/goquery"
	"github.com/bradfitz/iter"
	"github.com/mailgun/groupcache/v2"
	"github.com/valyala/fasttemplate"
)

// TODO The endpoint code needs to be re-done... getting crufty.
type requestContextKey string

// Content is a piece of content representing a page or portion of a page
type Content struct {
	Source   string `hcl:"source,optional"` // URL to fetch the main source
	Selector string `hcl:"select,optional"` // CSS Selector to extract content from - optional

	Replacements []Replacement `hcl:"replacement,block"` // May be empty

	CacheKey string `hcl:"cache,optional"`
	CacheTTL string `hcl:"ttl,optional"`

	Template string `hcl:"template,optional"` // Go template source -- URL
	JSON     string `hcl:"json,optional"`     // Used by templates to retrieve data -- URL

	parsedTemplate *template.Template // Cached/preparsed template... parsed on first use.

	//Options hcl.Body `hcl:",remain"`
	//FetchData map[string]string `hcl:"rules"`
}

// Fetch returns the rendered content (or from endpoint if endpoint is configured for the end point)
func (c *Content) Fetch(site *Host, contextdata map[string]interface{}) (string, error) {

	//log.Printf("Content From: %+v\n", endpoint)

	if c.Caching() {

		var content string

		var contextvalue = ContentContextValue{Site: site, Content: c, ContextData: contextdata}
		ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), requestContextKey("request"), contextvalue),
			time.Millisecond*500)

		defer cancel()

		if err := site.Cache.Get(ctx, c.InterpolatedCacheKey(contextdata), groupcache.StringSink(&content)); err != nil {
			log.Printf("Error getting from cache: %v\n", err)
			return "", err
		}

		return content, nil
	}

	return c.Render(site, contextdata)
}

// Render loads content from SourceURI and merges an fragements
// into the resulting document and returns the string representation
func (c *Content) Render(site *Host, contextdata map[string]interface{}) (string, error) {

	var renderedContent string

	content, err2 := c.fetcher(contextdata).Fetch()

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
	for _, merge := range c.Replacements {

		// Retrieve fragement.Source content
		fragment, _ := merge.Content.Fetch(site, contextdata)

		// Find the insertion point in sourceFile at merge.InsertAt
		insertSelection := doc.Find(merge.At) // Potentially costly, look into caching the source and insert selection points!?!

		// Insert the extracted content
		insertSelection.ReplaceWithHtml(fragment)
	}

	// By having the selector we can treat endpoints as a component
	if c.Selector != "" {
		// Get content at Selector
		renderedContent, err = doc.Find(c.Selector).Html()
	} else {
		renderedContent, err = doc.Html()
	}

	return renderedContent, err
}

// Factory method to return a fetcher for the end point
func (c *Content) fetcher(contextdata map[string]interface{}) DocumentFetcher {
	var fetcher DocumentFetcher
	fetcher = &StringFetcher{Body: ""} // Default to empty string

	if c.Source != "" {
		t := fasttemplate.New(c.Source, "{{", "}}")
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

	if c.Template != "" {

		if true /*endpoint.parsedTemplate == nil*/ {
			templateBytes, _ := ioutil.ReadFile(c.Template)
			templateContents := string(templateBytes)

			funcs := sprig.GenericFuncMap()
			funcs["N"] = iter.N
			funcs["unescape"] = unescape

			c.parsedTemplate = template.Must(template.New(c.Template).Funcs(template.FuncMap(funcs)).Parse(templateContents))
		}

		// interpolate the path for the any JSON source
		t := fasttemplate.New(c.JSON, "{{", "}}")
		json := t.ExecuteString(contextdata)

		return &RenderedTemplateFetcher{Template: c.parsedTemplate,
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

// ContentContextValue is passed via Context.WithValue() to the endpoint Getter Func
type ContentContextValue struct {
	Site        *Host
	Content     *Content
	ContextData map[string]interface{}
}

// Caching returns true if we are to use the endpoint
func (c *Content) Caching() bool {
	return c.CacheKey != ""
}

// InterpolatedCacheKey returns the interpolated endpoint key
func (c *Content) InterpolatedCacheKey(contextData map[string]interface{}) string {
	t := fasttemplate.New(c.CacheKey, "{{", "}}")
	return t.ExecuteString(contextData)

}
