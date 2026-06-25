// Example: chatroom
// Demonstrates a simple multi-user chatroom — broadcast messages to all
// connected clients, scoped routing with RouterGroup, and OnConnect/OnClose.
package main

import (
	"log"
	"sync"

	"github.com/AtoriUzawa/cira"
)

func main() {
	server := cira.New()

	var mu sync.Mutex
	clients := make(map[string]*cira.Conn)

	// broadcast pushes an event to every connected client.
	broadcast := func(event string, data any) {
		mu.Lock()
		defer mu.Unlock()
		for _, conn := range clients {
			conn.Do(func(ctx *cira.Context) {
				ctx.Push(event, data)
			})
		}
	}

	server.OnConnect = func(c *cira.Conn) {
		mu.Lock()
		clients[c.ID()] = c
		mu.Unlock()

		c.OnClose(func() {
			mu.Lock()
			delete(clients, c.ID())
			mu.Unlock()
			broadcast("chat.leave", map[string]string{"id": c.ID()})
		})

		broadcast("chat.join", map[string]string{"id": c.ID()})
		log.Printf("connected: %s (total=%d)", c.ID(), len(clients))
	}

	// Scoped route group for chat events.
	chat := server.Group("chat")

	chat.On("message", func(ctx *cira.Context) {
		broadcast("chat.message", map[string]string{
			"from": ctx.PeerID,
			"text": "hello everyone",
		})
		ctx.Resp(map[string]string{"status": "sent"})
	})

	chat.On("who", func(ctx *cira.Context) {
		mu.Lock()
		defer mu.Unlock()
		ids := make([]string, 0, len(clients))
		for id := range clients {
			ids = append(ids, id)
		}
		ctx.Resp(map[string]any{"clients": ids})
	})

	log.Println("chatroom listening on :8086")
	panic(server.Run(":8086"))
}
