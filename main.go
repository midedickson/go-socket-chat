package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Double-DOS/go-socket-chat/db"
	"github.com/Double-DOS/go-socket-chat/pkg/controllers"
	"github.com/Double-DOS/go-socket-chat/pkg/match"
	"github.com/Double-DOS/go-socket-chat/pkg/router"
	"github.com/Double-DOS/go-socket-chat/pkg/server"
	"github.com/Double-DOS/go-socket-chat/pkg/websocket"

	"github.com/Double-DOS/randommer-go"
)

func loadEnv() {
	readFile, err := os.Open(".env")
	if err != nil {
		fmt.Printf("Error opening .env file: %v\n", err)
		return
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		line := fileScanner.Text()

		// Skip empty lines or lines without '='
		if strings.TrimSpace(line) == "" || !strings.Contains(line, "=") {
			continue
		}

		lineKeyValuePair := strings.SplitN(line, "=", 2) // Split into at most 2 parts
		key := strings.TrimSpace(lineKeyValuePair[0])
		value := strings.TrimSpace(lineKeyValuePair[1])

		if key != "" {
			os.Setenv(key, value)
		}
	}
}

func main() {
	defer db.Close()
	db.Connect()
	db.Setup()
	match.CurrMaxGroup = 0
	websocket.RoomPool = websocket.NewRoomPool()
	go websocket.RoomPool.Start()
	fmt.Println("Mide's Chat Project")
	loadEnv()
	randommer_api_key := os.Getenv("RANDOMMER_API_KEY")
	randommer.Init(randommer_api_key)
	server := server.NewServer()

	router.SetupRoutes("POST", "/name", controllers.GetRandomAnonNames)
	router.SetupRoutes("GET", "/ws/:channel", controllers.ServeWebsocketPool)
	router.SetupRoutes("GET", "/ws/new", controllers.CreateNewPool)
	router.SetupRoutes("GET", "/stats", controllers.GetUserStats)
	server.ListenAndServe()

}
