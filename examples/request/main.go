// Example: request
// Demonstrates fire-and-forget request messaging — the server sends a
// request-type event without waiting for a correlated response.
package main

import (
	"log"

	"github.com/AtoriUzawa/cira"
)

func main() {
	server := cira.New()

	server.On("log.info", func(ctx *cira.Context) {
		// Unmarshal the incoming data.
		var msg map[string]any
		// In practice use c.codec or json.Unmarshal(ctx.Message.Data, &msg)
		_ = msg

		log.Printf("[INFO] client=%s route=%s", ctx.PeerID, ctx.Message.Route)

		// Send a fire-and-forget request (no response correlation).
		if err := ctx.Req("ack", map[string]string{
			"status": "received",
		}); err != nil {
			log.Println("req failed:", err)
		}
	})

	panic(server.Run(":8082"))
}
