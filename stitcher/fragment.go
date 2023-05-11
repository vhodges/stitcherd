package stitcher

import (
	"context"

	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mailgun/groupcache/v2"
	"github.com/valyala/fasttemplate"
)

// Fragments are renderable pieces of markup with one at the top level
// representing the page being sent to the browser.
type Fragment struct {

	Fetcher FragmentFetcher

	Fragments []Fragment // Can be nested 0 or more
	DocumentTransforms []DocumentTransform
	TransformSelfTransforms []DocumentTransform

	CacheKey string `json:"cache,optional"`
	CacheTTL string `json:"ttl,optional"`
}

// Caching returns true if we are to use the endpoint
func (fragment *Fragment) Cachable() bool {
	return fragment.CacheKey != ""
}

// InterpolatedCacheKey returns the interpolated endpoint key
func (fragment *Fragment) InterpolatedCacheKey(contextData map[string]interface{}) string {
	t := fasttemplate.New(fragment.CacheKey, "{{", "}}")
	return t.ExecuteString(contextData)

}

func (fragment *Fragment) Render(site *Host, contextdata map[string]interface{}) string {

	var this_content string
	var this_doc *goquery.Document
	var err error 

	this_content, err= fragment.Fetcher.Fetch(contextdata)

	if err != nil {
		return "" // TODO Handle error better.
	}

	this_doc, err = goquery.NewDocumentFromReader(strings.NewReader(this_content))
	if err != nil {
		return "" // TODO Handle error better.
	}


	for _, frag := range fragment.Fragments {
		var child_content string
		var child_doc *goquery.Document

		if frag.Cachable() {
			child_content, err = frag.FromCache(site, contextdata)
			if err != nil {
				continue; // Skip on error... TODO Log the error
			}
		} else {
			child_content = frag.Render(site, contextdata)		
		}

		child_doc, err = goquery.NewDocumentFromReader(strings.NewReader(child_content))
		if err != nil {
			continue; // Skip on error, TODO log the error
		}

		for _, transformation := range frag.DocumentTransforms {
			transformation.Transform(this_doc, child_doc)
		}
	}

	for _, transformation := range fragment.TransformSelfTransforms {
		transformation.Transform(this_doc, nil) // Only makes sense for add_class
	}


	html, err2 := this_doc.Html()

	if err2 != nil {
		return "" // TODO Handle error better
	}

	return html
}


func (fragment *Fragment) FromCache(site *Host, contextdata map[string]interface{}) (string, error) {

	var content string

	var contextvalue = FragmentRenderContext{Site: site, Fragment: fragment, ContextData: contextdata}
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), requestContextKey("request"), contextvalue),
		time.Millisecond*2000) // TODO Make this configurable

	defer cancel()

	if err := site.Cache.Get(ctx, fragment.InterpolatedCacheKey(contextdata), groupcache.StringSink(&content)); err != nil {
		log.Printf("Error getting from cache: %v\n", err)
		return "", err
	}

	return content, nil
}

func FillFragmentCache(ctx context.Context, id string, dest groupcache.Sink) error {
	var err error

	v := ctx.Value(requestContextKey("request"))

	r, ok := v.(FragmentRenderContext)

	if ok {

		content := r.Fragment.Render(r.Site, r.ContextData)

		if err != nil {
			log.Printf("Error: %v - '%+v'\n", err, r.Fragment)
			return err
		}

		ttl, err := time.ParseDuration(r.Fragment.CacheTTL)
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
}
