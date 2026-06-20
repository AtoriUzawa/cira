// Example: call
// Demonstrates synchronous RPC-style Call — the server sends a request and
// blocks until the client responds or the timeout fires.
package main

import (
	"log"
	"time"

	"github.com/AtoriUzawa/cira"
)

func main() {
	server := cira.New()

	server.On("fetch", func(ctx *cira.Context) {
		// Set a per-call timeout.
		ctx.Timeout = 3 * time.Second

		var result map[string]any

		// Synchronous call — blocks until the client responds.
		if err := ctx.Call("db.query", map[string]string{
			"table": "users",
		}, &result); err != nil {
			ctx.Resp(map[string]string{"error": err.Error()})
			return
		}

		ctx.Resp(result)
	})

	log.Println("call server listening on :8083")
	panic(server.Run(":8083"))
}
