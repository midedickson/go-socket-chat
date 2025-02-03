package server

import (
	"fmt"
	"net/http"

	"github.com/Double-DOS/go-socket-chat/db"
	"github.com/Double-DOS/go-socket-chat/pkg/router"
)

func NewServer() *http.Server {
	// Use the global DB connection from the db package
	if db.DB == nil {
		fmt.Println("Database connection is not initialized")
		return nil
	}

	r := http.NewServeMux()
	return &http.Server{
		Addr:    "0.0.0.0:9000",
		Handler: router.Matcher()(r),
	}
}
