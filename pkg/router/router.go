package router

import (
	"log"
	"net/http"
	"regexp"
)

func SetupRoutes(method, path string, handler func(http.ResponseWriter, *http.Request)) {
	r := &Route{
		Path:    path,
		Handler: handler,
		Method:  method,
	}
	// remove preceding and trailing slashes
	r.Path = regexp.MustCompile("^/+|/+$").ReplaceAllString(r.Path, "")
	if r.Path == "" {
		r.Path = "/"
	}
	// the handler is added as just another middleware
	r.stack = append(r.stack, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			r.Handler(w, req)
		})
	})
	addRouteToRtable(r)
}

func addRouteToRtable(r *Route) {
	if rtable[r.Path] == nil {
		rtable[r.Path] = make(map[string]*Route)
	}
	rtable[r.Path][r.Method] = r
	log.Printf("Adding Path /%s to routing table", r.Path)
}
