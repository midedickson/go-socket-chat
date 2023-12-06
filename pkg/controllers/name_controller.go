package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Double-DOS/go-socket-chat/pkg/websocket"
	"github.com/Double-DOS/randommer-go"
)

func GetRandomAnonNames(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		names := randommer.GetRandomNames("fullname", 1)

		msg, err := json.Marshal(websocket.ApiResponse{Success: true, Message: "Fetched Random Name Successfully", Data: names[0]})
		if err == nil {
			w.WriteHeader(http.StatusOK)
			w.Write(msg)
		} else {
			log.Printf("err: %s", err)
			msg, _ = json.Marshal(websocket.ApiResponse{Success: false, Message: "Fetching Random Name Failed", Data: nil})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(msg)
		}
	}
}
