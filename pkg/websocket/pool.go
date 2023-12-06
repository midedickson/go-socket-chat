package websocket

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var poolTable = make(map[string]*Pool)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
	Data       *PoolData
}

type PoolData struct {
	PoolId       string
	LastActivity time.Time
}

func GetPool(key string) *Pool {
	for k := range poolTable {
		if k == key {
			return poolTable[key]

		}
	}

	return nil

}

func NewPool() (string, *Pool) {
	pool := &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
	}
	poolId := generateUniquePoolId()
	poolTable[poolId] = pool
	// todo: handle storage to database with PoolData struct
	pool.Data = &PoolData{
		PoolId:       poolId,
		LastActivity: time.Now(),
	}
	return poolId, pool
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

func (pool *Pool) Start(poolId string) {
	log.Printf("Starting pool with poolId: %s", poolId)
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
			// todo; handle updating last activity time
			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
			pool.updateLastActivity()
		}
	}
}

func (pool *Pool) updateLastActivity() {
	pool.Data.LastActivity = time.Now()
}

func ServeWS(pool *Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Websocket endpoint reached")

	conn, err := Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}
	client := &Client{
		Conn: conn,
		Pool: pool,
	}
	pool.Register <- client
	client.Read()
}
