# Cira

A lightweight, event-driven WebSocket framework for Go.

Cira provides an ergonomic abstraction over WebSocket connections with
event routing, request-response messaging, synchronous RPC-style calls,
and long-lived streams.

## Features

- **Event-driven routing** — named events with dot-separated scoping and middleware
- **Bidirectional communication** — supports both listening and outbound dialing
- **Five messaging patterns** — Push, Request, Response, Call, and Stream
- **Streaming transport** — long-lived message streams identified by user-defined IDs
- **Connection management** — connection lookup, close callbacks, and contextual execution
- **Customizable** — pluggable codec, ID generator, and WebSocket upgrader
- **Lightweight** — built on top of gorilla/websocket

## Installation

```bash
go get github.com/AtoriUzawa/cira
````

## Quick Start

### Server

```go
package main

import "github.com/AtoriUzawa/cira"

func main() {
	server := cira.New()

	server.On("hello", func(c *cira.Context) {
		c.Resp(map[string]string{
			"message": "world",
		})
	})

	panic(server.Run(":8080"))
}
```

### Client

```go
package main

import (
	"github.com/AtoriUzawa/cira"
)

func main() {
	client := cira.New()

	conn, err := client.Dial("ws://localhost:8080/ws")
	if err != nil {
		panic(err)
	}

	conn.Do(func(c *cira.Context) {
		_ = c.Push("hello", "world")
	})

	select {}
}
```

---

## Messaging Patterns

| Pattern  | Description                   | Response |
| -------- | ----------------------------- | -------- |
| Push     | One-way event delivery        | No       |
| Request  | Request event                 | Optional |
| Response | Reply to a request            | Yes      |
| Call     | Request and wait for response | Yes      |
| Stream   | Continuous message transport  | Multiple |

---

## Push

Send a one-way event.

```go
ctx.Push("chat.message", map[string]string{
	"user": "alice",
	"text": "hello",
})
```

---

## Request

Send a request event without waiting for a response.

```go
ctx.Req("status.update", map[string]bool{
	"online": true,
})
```

---

## Response

Reply to the current request.

```go
server.On("ping", func(c *cira.Context) {
	c.Resp("pong")
})
```

---

## Call

Send a request and wait synchronously for a response.

```go
var resp map[string]any

err := ctx.Call(
	"user.info",
	map[string]string{
		"id": "123",
	},
	&resp,
)

if err != nil {
	return
}
```

### Timeout

```go
ctx.Timeout = 5 * time.Second
```

Default timeout:

```go
30 * time.Second
```

---

## Stream

Streams provide a long-lived communication channel identified by a stream ID.

### Sender

```go
stream := ctx.OpenStream("upload.file")

defer ctx.CloseStream()

_ = stream.Send("chunk_1")
_ = stream.Send("chunk_2")
_ = stream.Send("chunk_3")
```

### Receiver

```go
stream := ctx.OpenStream("upload.file")

defer ctx.CloseStream()

for {
	var chunk string

	err := stream.Recv(&chunk)
	if err != nil {
		break
	}

	fmt.Println(chunk)
}
```

### Stream Timeout

```go
err := stream.RecvTimeout(&chunk)
```

Default timeout uses:

```go
ctx.Timeout
```

---

## Middleware

```go
func Logger(next cira.HandlerFunc) cira.HandlerFunc {
	return func(c *cira.Context) {
		log.Println(c.Message.Route)
		next(c)
	}
}

server.Use(Logger)
```

Middleware executes in reverse registration order.

---

## Routing

Routes support dot-separated grouping.

```go
api := server.Group("api")

api.On("user.list", handler)
```

Resulting route:

```text
api.user.list
```

---

## Connection Management

### Access Current Connection

```go
server.On("hello", func(c *cira.Context) {
	fmt.Println(c.Conn.ID())
})
```

### Lookup Connection

```go
conn, err := server.Conn(id)
if err != nil {
	return
}
```

### Close Connection

```go
conn.Close()
```

### On Close

```go
conn.OnClose(func() {
	log.Println("connection closed")
})
```

---

## Configuration

### Custom Codec

```go
engine := cira.New(
	cira.WithCodec(myCodec),
)
```

### Custom ID Generator

```go
engine := cira.New(
	cira.WithIDGenerator(myGenerator),
)
```

### Custom WebSocket Upgrader

```go
engine := cira.New(
	cira.WithUpgrader(myUpgrader),
)
```

---

## Examples

| Example             | Description                        |
| ------------------- | ---------------------------------- |
| examples/hello      | Minimal server                     |
| examples/push       | Push messaging                     |
| examples/request    | Request messaging                  |
| examples/call       | RPC-style call                     |
| examples/stream     | Stream communication               |
| examples/client     | Client dial and file upload stream |
| examples/middleware | Middleware usage                   |
| examples/connection | Connection lifecycle               |
| examples/chatroom   | Multi-client chatroom              |

---

## Architecture

```text
┌─────────┐
│ Engine  │
└────┬────┘
     │
┌────▼────┐
│  Peer   │
└────┬────┘
     │
┌────▼────┐
│ Context │
└────┬────┘
     │
 ┌───┴───┐
 │Message│
 │Stream │
 └───────┘
```

A connection is represented internally as a Peer.
Business logic interacts through Conn and Context abstractions.

---

## License

MIT
