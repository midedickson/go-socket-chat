package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Double-DOS/go-socket-chat/pkg/websocket"

	"github.com/Double-DOS/randommer-go"
)

func loadEnv() {
	readFile, err := os.Open(".env")

	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}

	readFile.Close()

	for _, line := range fileLines {
		line_key_value_pair := strings.Split(line, "=")
		os.Setenv(line_key_value_pair[0], line_key_value_pair[1])
	}
}
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

	http.HandleFunc("/ws/mide", func(w http.ResponseWriter, r *http.Request) {
		pool, exists := websocket.NewPool("mide")
		if !exists {
			go pool.Start()

		}
		serveWS(pool, w, r)
	})
	http.HandleFunc("/ws/dickson", func(w http.ResponseWriter, r *http.Request) {
		pool, exists := websocket.NewPool("dickson")
		if !exists {
			go pool.Start()

		}
		serveWS(pool, w, r)
	})

	http.HandleFunc("/name", func(w http.ResponseWriter, r *http.Request) {
		addDefaultHeaders(w)
		if r.Method == "GET" {
			resp, err := http.Get("https://names.drycodes.com/1")

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
	loadEnv()
	randommer_api_key := os.Getenv("RANDOMMER_API_KEY")
	randommer.Init(randommer_api_key)

	setupRoutes()
	http.ListenAndServe(":9000", nil)

}
