package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Double-DOS/go-socket-chat/pkg/controllers"
	"github.com/Double-DOS/go-socket-chat/pkg/router"
	"github.com/Double-DOS/go-socket-chat/pkg/server"

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

func main() {
	fmt.Println("Mide's Chat Project")
	loadEnv()
	randommer_api_key := os.Getenv("RANDOMMER_API_KEY")
	randommer.Init(randommer_api_key)
	server := server.NewServer()

	router.SetupRoutes("GET", "/ws/:channel", controllers.ServeWebsocketPool)
	router.SetupRoutes("GET", "/name", controllers.GetRandomAnonNames)
	server.ListenAndServe()

}
