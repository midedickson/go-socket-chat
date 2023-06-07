package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Double-DOS/go-socket-chat/pkg/websocket"
)

func serveWS(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Websocket endpoint reached")

	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}
	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
	}
	pool.Register <- client
	client.Read()
}

func addDefaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}

func setupRoutes() {
	pool := websocket.NewPool()
	go pool.Start()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(pool, w, r)
	})

	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			resp, err := http.Get("https://names.drycodes.com/1")
			res, _ := json.Marshal(resp.Body)
			addDefaultHeaders(w)

			if err == nil {
				msg, err := json.Marshal(websocket.ApiResponse{Success: true, Message: "Fetched Random Name Successfully", Data: res[0]})
				if err == nil {
					w.WriteHeader(http.StatusOK)
					w.Write(msg)
				} else {
					log.Printf("err: %s", err)
					msg, _ = json.Marshal(websocket.ApiResponse{Success: false, Message: "Fetching Random Name Failed", Data: nil})
					w.WriteHeader(http.StatusInternalServerError)

					w.Write(msg)
				}
			} else {
			}
		}
	})
}

func main() {
	fmt.Println("Mide's Chat Project")
	setupRoutes()
	http.ListenAndServe(":9000", nil)

}
