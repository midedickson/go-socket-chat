package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/Double-DOS/go-socket-chat/pkg/constants"
	"github.com/Double-DOS/go-socket-chat/pkg/websocket"
)

func addDefaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	// Allow requests from any origin

	w.Header().Set("Access-Control-Allow-Origin", "*")

}
func Matcher() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			addDefaultHeaders(w)
			// clone the handler
			sh := h

			// get route function
			route := Route{
				Path:    Path(r.URL),
				Method:  r.Method,
				Handler: nil,
				Params:  map[string]string{},
			}
			// remove preceding and trailing slashes
			route.Path = regexp.MustCompile("^/+|/+$").ReplaceAllString(route.Path, "")
			if route.Path == "" {
				route.Path = "/"
			}
			log.Println("reaching path: " + route.Path + " for " + route.Method + " request")
			// if path found in top level of table
			if rtable[route.Path] != nil && rtable[route.Path][route.Method] != nil {
				route = *rtable[route.Path][route.Method]
			} else {
				// handle path with parameters
				paths := strings.Split(route.Path, "/")
			TLoop:
				for path, methods := range rtable {
					variables := strings.Split(path, "/")
					if len(paths) == len(variables) && methods[r.Method] != nil {
						match := true
					VLoop:
						for i, variable := range variables {
							if variable != paths[i] && (len(variable) == 0 || variable[0] != ':') {
								match = false
								break VLoop
							}
						}
						if match {
							params := Params(route.Path, path)
							route = *methods[route.Method]
							route.Params = params
							break TLoop
						}
					}
				}
			}

			if route.Handler == nil {
				w.WriteHeader(http.StatusNotFound)
				msg, _ := json.Marshal(websocket.ApiResponse{Success: false, Message: fmt.Sprintf("Path: /%s Not Found", route.Path), Data: nil})
				w.Write(msg)
				return
			}

			ctx := context.WithValue(r.Context(), constants.ChannelNameCtxKey{}, route.Params)
			r = r.WithContext(ctx)
			// call route specific middleware(s)
			for i := range route.stack {
				sh = route.stack[len(route.stack)-1-i](sh)
			}
			sh.ServeHTTP(w, r)

		})
	}
}
