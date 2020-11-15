package stitcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TODO Split these up... getting crufty

// DocumentFetcher are something that retrieves content
type DocumentFetcher interface {
	Fetch() (string, error)
}

// FileFetcher is a type of Fetcher that loads content from a file.
type FileFetcher struct {
	Path string
}

// Fetch returns an io.ReadCloser for the file
func (fetcher *FileFetcher) Fetch() (string, error) {
	b, err := ioutil.ReadFile(fetcher.Path)
	return string(b), err
}

// StringFetcher is a type of Fetcher that loads content from a string
type StringFetcher struct {
	Body string
}

// Fetch returns the string
func (fetcher *StringFetcher) Fetch() (string, error) {
	return fetcher.Body, nil
}

// URIFetcher is a fetcher from the network (typically http(s))
type URIFetcher struct {
	URI string
}

// Fetch returns res.Body
func (fetcher *URIFetcher) Fetch() (string, error) {

	// TODO At somepoint we'll probably need finer grained control over the client/request
	// not to mention cookie/session handling
	res, err := http.Get(fetcher.URI)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	b, err2 := ioutil.ReadAll(res.Body)
	return string(b), err2
}

// RenderedTemplateFetcher Parses and executes a Go template pulling data from JSONURL (if not empty)
type RenderedTemplateFetcher struct {
	Template      *template.Template
	DataURL       string // URL for the JSON source
	SourceFetcher DocumentFetcher

	RequestContext map[string]interface{}
}

func (fetcher *RenderedTemplateFetcher) fetchJSON() (string, error) {
	// TODO At somepoint we'll probably need finer grained control over the client/request
	// not to mention cookie/session handling
	res, err := http.Get(fetcher.DataURL)

	if err != nil {
		log.Println(err)
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		e := fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
		return "", e
	}

	b, err2 := ioutil.ReadAll(res.Body)
	return string(b), err2
}

// Fetch returns the rendered template TODO Break this up
func (fetcher *RenderedTemplateFetcher) Fetch() (string, error) {
	var buffer bytes.Buffer
	var err error
	var jsonString string
	var source string

	var data = make(map[string]interface{})

	data["request"] = fetcher.RequestContext

	source, err = fetcher.SourceFetcher.Fetch()
	if err == nil {
		data["source"] = source // Raw source from Source
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(source))
		if err == nil {
			data["document"] = doc // Domified version of source
		} else {
			log.Printf("Template Fetcher goquery Error: %v\n", err)
		}
	} else {
		log.Printf("Template Fetcher Source Fetch Error: %v\n", err)
	}

	if fetcher.DataURL != "" {
		jsonString, err = fetcher.fetchJSON()
		if err == nil {
			var jsonData interface{}
			json.Unmarshal([]byte(jsonString), &jsonData)
			data["json"] = jsonData
		} else {
			log.Printf("JSON Fetch Error: %v\n", err)
		}
	}

	err = fetcher.Template.Execute(&buffer, data)

	if err != nil {
		log.Println("template.Execute", err)
		return "", err
	}

	return buffer.String(), nil
}
