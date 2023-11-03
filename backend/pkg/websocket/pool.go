package websocket

import (
	"fmt"

	"github.com/google/uuid"
)

var poolTable = make(map[string]*Pool)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

func NewPool(key string) (*Pool, bool) {
	for k, p := range poolTable {
		if k == key {
			return p, true
		}
	}
	pool := &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
	}
	// poolId := generateUniquePoolId()
	poolTable[key] = pool
	return pool, false

}

func generateUniquePoolId() string {
	uuidString := uuid.NewString()
	firstEightChar := uuidString[:8]
	for k := range poolTable {
		if k == firstEightChar {
			return generateUniquePoolId()
		}
	}
	return firstEightChar
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("size of connection pool: ", len(pool.Clients))
			for client := range pool.Clients {
				fmt.Println(client)
				client.Conn.WriteJSON(Message{Type: 1, Body: "New user joined...", Author: "System", Variant: "connection"})
			}
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("size of connection pool: ", len(pool.Clients))
			for client := range pool.Clients {
				fmt.Println(client)
				client.Conn.WriteJSON(Message{Type: 1, Body: "A user disconnected...", Author: "System", Variant: "connection"})
			}

		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients")
			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}
