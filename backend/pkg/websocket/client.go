package websocket

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}
type ReceivedMessage struct {
	Body    string `json:"body"`
	Author  string `json:"author"`
	Variant string `json:"variant"`
}

type Message struct {
	Type    int    `json:"type"`
	Body    string `json:"body"`
	Author  string `json:"author"`
	Variant string `json:"variant"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {

		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		var receivedMessage *ReceivedMessage

		if err = json.Unmarshal(p, &receivedMessage); err != nil {
			log.Fatal(err)
			fmt.Println("Message received was unprocessable")
		} else {
			message := Message{Type: messageType, Body: receivedMessage.Body, Author: receivedMessage.Author, Variant: receivedMessage.Variant}
			c.Pool.Broadcast <- message
			fmt.Println("Message received:%+V\n", message)
		}

	}
}
