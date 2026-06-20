// Example: connection
// Demonstrates connection lifecycle — OnConnect, OnClose, and connection
// lookup via Engine.Conn.
package main

import (
	"log"
	"sync"

	"github.com/AtoriUzawa/cira"
)

func main() {
	server := cira.New()

	var mu sync.Mutex
	active := make(map[string]*cira.Conn)

	server.OnConnect = func(c *cira.Conn) {
		mu.Lock()
		active[c.ID()] = c
		mu.Unlock()
		log.Printf("connected: %s (total=%d)", c.ID(), len(active))

		c.OnClose(func() {
			mu.Lock()
			delete(active, c.ID())
			mu.Unlock()
			log.Printf("disconnected: %s (total=%d)", c.ID(), len(active))
		})
	}

	// Look up a connection by ID and execute a function in its context.
	server.On("lookup", func(ctx *cira.Context) {
		conn, err := server.Conn("some-connection-id")
		if err != nil {
			ctx.Resp(map[string]string{"error": "not found"})
			return
		}

		conn.Do(func(c *cira.Context) {
			c.Push("alert", map[string]string{"msg": "you were looked up"})
		})

		ctx.Resp(map[string]string{"status": "ok"})
	})

	panic(server.Run(":8085"))
}
