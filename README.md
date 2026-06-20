# Cira

A lightweight, event-driven WebSocket framework for Go.

## Features

- **Event-driven routing** — named events with dot-separated scoping and middleware
- **Four messaging patterns** — Push, Request, Response, and synchronous Call
- **Connection management** — lookup by ID, close callbacks, per-connection execution
- **Customizable** — pluggable codec (JSON default), ID generator, WebSocket upgrader
- **Lightweight** — only `gorilla/websocket` and `google/uuid` as dependencies

## Installation

```bash
go get github.com/AtoriUzawa/cira
```

## Quick Start

```go
package main

import "github.com/AtoriUzawa/cira"

func main() {
    server := cira.New()

    server.On("hello", func(ctx *cira.Context) {
        ctx.Resp(map[string]string{"msg": "world"})
    })

    panic(server.Run(":8080"))
}
```

## Push

Send a one-way event to the client. No response expected.

```go
ctx.Push("chat.message", map[string]string{"from": "alice", "text": "hi"})
```

## Request

Send a request-type event without waiting for a correlated response.

```go
ctx.Req("status.update", map[string]bool{"online": true})
```

## Response

Reply to the current incoming message using its correlation ID.

```go
server.On("ping", func(ctx *cira.Context) {
    ctx.Resp(map[string]string{"msg": "pong"})
})
```

## Call

Send a request and block until the response arrives or the timeout fires.

```go
server.On("lookup", func(ctx *cira.Context) {
    ctx.Timeout = 3 * time.Second
    var result map[string]any
    if err := ctx.Call("db.query", query, &result); err != nil {
        ctx.Resp(map[string]string{"error": err.Error()})
        return
    }
    ctx.Resp(result)
})
```

## Middleware

```go
func Logger(next cira.HandlerFunc) cira.HandlerFunc {
    return func(ctx *cira.Context) {
        log.Println("event:", ctx.Message.Route)
        next(ctx)
    }
}

server.Use(Logger)
```

Middleware is applied in reverse registration order (last = outermost). Scoped groups inherit parent middleware.

## Routing

Routes use dot-separated prefixes inherited from `RouterGroup`:

```go
api := server.Group("api")
api.On("user.list", handleUsers) // full route: "api.user.list"
```

## Examples

| Example | Description |
|---------|-------------|
| [`examples/hello`](examples/hello/main.go) | Minimal server with one handler |
| [`examples/push`](examples/push/main.go) | Server-to-client push and heartbeat |
| [`examples/request`](examples/request/main.go) | Fire-and-forget request messaging |
| [`examples/call`](examples/call/main.go) | Synchronous RPC-style Call with timeout |
| [`examples/middleware`](examples/middleware/main.go) | Logging and timing middleware |
| [`examples/connection`](examples/connection/main.go) | Connection lifecycle and lookup |
| [`examples/chatroom`](examples/chatroom/main.go) | Multi-client chatroom with scoped routing |

## Architecture

See [`architecture.md`](architecture.md) for the component diagram, messaging protocol, and design decisions.

## License

MIT
