package server

import (
	"net/http"

	"github.com/Double-DOS/go-socket-chat/pkg/router"
)

func NewServer() *http.Server {
	r := http.NewServeMux()
	return &http.Server{
		Addr:    "localhost:9000",
		Handler: router.Matcher()(r),
	}
}
