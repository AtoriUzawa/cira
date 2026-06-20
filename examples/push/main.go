// Example: push
// Demonstrates server-to-client push messaging via Conn.Do.
package main

import (
	"log"
	"time"

	"github.com/AtoriUzawa/cira"
)

func main() {
	server := cira.New()

	server.OnConnect = func(c *cira.Conn) {
		log.Println("client connected:", c.ID())

		// Push a welcome message to the new connection.
		c.Do(func(ctx *cira.Context) {
			if err := ctx.Push("welcome", map[string]string{
				"id":      c.ID(),
				"message": "connected to server",
			}); err != nil {
				log.Println("push failed:", err)
			}
		})
	}

	// Broadcast a heartbeat to all known connections every 5 seconds.
	server.On("subscribe.heartbeat", func(ctx *cira.Context) {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := ctx.Push("heartbeat", map[string]int64{
				"ts": time.Now().Unix(),
			}); err != nil {
				return
			}
		}
	})

	panic(server.Run(":8081"))
}
