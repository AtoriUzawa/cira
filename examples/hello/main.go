// Example: hello
// Demonstrates the simplest cira server — register a handler and respond.
package main

import "github.com/AtoriUzawa/cira"

func main() {
	server := cira.New()

	server.On("hello", func(ctx *cira.Context) {
		ctx.Resp(map[string]string{"msg": "world"})
	})

	panic(server.Run(":8080"))
}
