package stitcher

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// EndPointHandler uses the EndPoint to render content
func EndPointHandler(site *Host, endpoint EndPoint) func(http.ResponseWriter, *http.Request) {

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var fetchContext map[string]interface{} = make(map[string]interface{})

		if endpoint.Throttling(r) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		// TODO Better request tracing... (context.Context too?)
		fetchContext["_requestId"] = nextRequestID()

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

		content, err := endpoint.Content(site, fetchContext)

		if err != nil {
			log.Printf("Error from endpoint '%s': %v", endpoint.Route, err)
			fmt.Fprintln(w, "")
		} else {
			fmt.Fprintln(w, content)
		}

		elapsed := time.Since(start)
		log.Println(fetchContext["_requestId"], r.Method, r.URL.Path, r.Proto, elapsed)
	}
}
