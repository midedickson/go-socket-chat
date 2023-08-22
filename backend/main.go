package main

import (
	"encoding/json"
	"fmt"
	"io"
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

	http.HandleFunc("/name", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			resp, err := http.Get("https://names.drycodes.com/1")

			addDefaultHeaders(w)

			if err != nil {
				msg, _ := json.Marshal(websocket.ApiResponse{Success: false, Message: "Fetching Random Name Failed", Data: nil})
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(msg)
			} else {
				defer resp.Body.Close()
				read_body, err := io.ReadAll(resp.Body)
				if err != nil {
					msg, _ := json.Marshal(websocket.ApiResponse{Success: false, Message: "Fetching Random Name Failed", Data: nil})
					w.WriteHeader(http.StatusInternalServerError)
					w.Write(msg)
				}
				var arrayResponse []string
				if err := json.Unmarshal(read_body, &arrayResponse); err != nil {
					log.Printf("err: %s", err)
					msg, _ := json.Marshal(websocket.ApiResponse{Success: false, Message: "Fetching Random Name Failed", Data: nil})
					w.WriteHeader(http.StatusInternalServerError)

					w.Write(msg)
				}
				msg, err := json.Marshal(websocket.ApiResponse{Success: true, Message: "Fetched Random Name Successfully", Data: arrayResponse[0]})
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
	})
}

func main() {
	fmt.Println("Mide's Chat Project")
	setupRoutes()
	http.ListenAndServe(":9000", nil)

}
