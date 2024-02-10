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

	// append the quivalent options route to allow CORS preflight request from the browser
	optionR := &Route{
		Path:    r.Path,
		Handler: defaultCORSHandler,
		Method:  http.MethodOptions,
	}

	optionR.stack = append(optionR.stack, func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			optionR.Handler(w, req)
		})
	})
	addRouteToRtable(optionR)

}

func addRouteToRtable(r *Route) {
	if rtable[r.Path] == nil {
		rtable[r.Path] = make(map[string]*Route)
	}
	rtable[r.Path][r.Method] = r
	log.Printf("Adding Path /%s to routing table", r.Path)
}

func defaultCORSHandler(w http.ResponseWriter, r *http.Request) {
	// Allow requests from any origin

	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Allow specified HTTP methods

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	// Allow specified headers

	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

	w.WriteHeader(http.StatusAccepted)
}
