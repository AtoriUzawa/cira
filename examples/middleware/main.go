// Example: middleware
// Demonstrates logging and timing middleware chained around a handler.
package main

import (
	"log"
	"time"

	"github.com/AtoriUzawa/cira"
)

func main() {
	server := cira.New()

	// Attach middleware at the engine level.
	server.Use(LoggerMiddleware, TimerMiddleware)

	server.On("echo", func(ctx *cira.Context) {
		ctx.Resp(map[string]string{"echo": "ok"})
	})

	panic(server.Run(":8084"))
}

// LoggerMiddleware logs every incoming event with its route and client ID.
func LoggerMiddleware(next cira.HandlerFunc) cira.HandlerFunc {
	return func(ctx *cira.Context) {
		log.Printf("[req] client=%s route=%s", ctx.PeerID, ctx.Message.Route)
		next(ctx)
		log.Printf("[res] client=%s route=%s", ctx.PeerID, ctx.Message.Route)
	}
}

// TimerMiddleware measures and logs the duration of each handler.
func TimerMiddleware(next cira.HandlerFunc) cira.HandlerFunc {
	return func(ctx *cira.Context) {
		start := time.Now()
		next(ctx)
		log.Printf("[timer] route=%s duration=%v", ctx.Message.Route, time.Since(start))
	}
}
