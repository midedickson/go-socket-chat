package router

import "net/http"

var rtable = map[string]map[string]*Route{}

type Route struct {
	Path    string
	Method  string
	Handler func(http.ResponseWriter, *http.Request)
	Query   map[string]string
	Params  map[string]string
	stack   []func(http.Handler) http.Handler
}
