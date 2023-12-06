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
	params := r.Context().Value(constants.ChannelNameCtxKey{}).(map[string]string)
	log.Printf("Fetching channel: %v", params["channel"])
	pool := websocket.GetPool(params["channel"])
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
	poolId, pool := websocket.NewPool()
	log.Printf("Creating new pool with id %v", poolId)
	if pool == nil {
		msg, err := json.Marshal(websocket.ApiResponse{Success: false, Message: "Problem with creating new pool!", Data: nil})
		if err == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(msg)
		} else {
			log.Printf("err: %s", err)
		}
	} else {
		go pool.Start(poolId)

		msg, err := json.Marshal(websocket.ApiResponse{Success: true, Message: "Pool created successfully!", Data: pool.Data})
		if err != nil {
			log.Printf("err: %s", err)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(msg)
		}

	}

}
