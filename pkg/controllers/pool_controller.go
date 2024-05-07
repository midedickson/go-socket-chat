package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Double-DOS/go-socket-chat/pkg/constants"
	"github.com/Double-DOS/go-socket-chat/pkg/websocket"
)

func ServeWebsocketPool(w http.ResponseWriter, r *http.Request) {
	// Allow requests from any origin

	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Allow specified HTTP methods

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	// Allow specified headers

	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	params := r.Context().Value(constants.ChannelNameCtxKey{}).(map[string]string)
	poolChannel := params["channel"]
	log.Printf("Fetching channel: %v", poolChannel)
	if poolChannel == "room" {
		websocket.ServeRoomWS(w, r)
		return
	}
	pool := websocket.GetPool(poolChannel)
	if pool == nil {
		msg, err := json.Marshal(websocket.ApiResponse{Success: false, Message: fmt.Sprintf("Pool with key: %s does not exists", params["channel"]), Data: nil})
		if err == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write(msg)
		} else {
			log.Printf("err: %s", err)
		}
	} else {
		websocket.ServeWS(pool, w, r)
	}

}

func CreateNewPool(w http.ResponseWriter, r *http.Request) {
	_, pool := websocket.NewPool()
	// log.Printf("Creating new pool with id %v", poolId)
	if pool == nil {
		msg, err := json.Marshal(websocket.ApiResponse{Success: false, Message: "Problem with creating new pool!", Data: nil})
		if err == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(msg)
		} else {
			log.Printf("err: %s", err)
		}
	} else {
		go pool.Start()

		msg, err := json.Marshal(websocket.ApiResponse{Success: true, Message: "Pool created successfully!", Data: pool.Data})
		if err != nil {
			log.Printf("err: %s", err)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(msg)
		}

	}

}
