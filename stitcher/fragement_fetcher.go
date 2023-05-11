package stitcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"

//	"log"
	"net/http"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/PuerkitoBio/goquery"
	"github.com/bradfitz/iter"
	"github.com/valyala/fasttemplate"
)

type FragmentFetcher struct {
	Type string

	Source string                 // Filepath, URI, static/interpolated string, Cachekey name (For rendered)

	Template string               // Go template name/path
	IsJson bool                   // Parse fetched fragment as JSON, Only useful if Template is present.

	// Values suitable for URI sources
	URIVerb string                // GET POST PATCH DELETE(?) etc
	URIParams map[string]string   // Post and Get will be different
	Headers map[string]string     // Makes sense for remote fragments
}

func (fetcher *FragmentFetcher) Fetch(contextdata map[string]interface{}) (string, error) {

	t := fasttemplate.New(fetcher.Source, "{{", "}}")
	src := t.ExecuteString(contextdata)

	var fetched_fragment string = src // Default to String source
	var err error = nil
	
	switch fetcher.Type {
	case "uri":
		fetched_fragment, err = fetcher.FetchURI(src)
	case "file":
		fetched_fragment, err = fetcher.FetchFile(src)
	}

	if fetcher.Template != "" {
		templateBytes, _ := ioutil.ReadFile(fetcher.Template)
		templateContents := string(templateBytes)

		funcs := sprig.GenericFuncMap()
		funcs["N"] = iter.N
		funcs["unescape"] = unescape

		parsedTemplate := template.Must(template.New(fetcher.Template).Funcs(template.FuncMap(funcs)).Parse(templateContents))

		var buffer bytes.Buffer
		var data = make(map[string]interface{})

		if fetcher.IsJson {
			var jsonData interface{}
		
			json.Unmarshal([]byte(fetched_fragment), &jsonData)
		
			data["json"] = jsonData
		} else {
			var doc *goquery.Document
			doc, err = goquery.NewDocumentFromReader(strings.NewReader(fetched_fragment))
			data["document"] = doc  // Add the Dom tree to the data for the template
		}

		err = parsedTemplate.Execute(&buffer, data)
	
		if err != nil {
			return "", err
		}
	
		return buffer.String(), nil
	}

	return fetched_fragment, nil
}

func (fetcher *FragmentFetcher) FetchURI(src string) (string, error) {

	// TODO At somepoint we'll probably need finer grained control over the client/request
	// not to mention cookie/session handling
	res, err := http.Get(src)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Will need to handle/follow redirects - might be an option on the http client

	b, err2 := ioutil.ReadAll(res.Body)
	return string(b), err2
}

func (fetcher *FragmentFetcher) FetchFile(src string) (string, error) {
	b, err := ioutil.ReadFile(src)
	return string(b), err
}

func unescape(s string) template.HTML {
	return template.HTML(s)
}
