# Architecture

## Package Structure

```
github.com/AtoriUzawa/cira
├── package cira              # Public API
│   ├── engine.go             # Server entry point, HTTP-to-WS upgrade
│   ├── group.go              # RouterGroup: scoped routing + middleware
│   ├── router.go             # Router: event → handler dispatch
│   ├── handler.go            # HandlerFunc and Middleware type definitions
│   ├── context.go            # Per-message context: Push/Req/Resp/Call
│   ├── conn.go               # Connection handle: ID, Do, OnClose
│   ├── hub.go                # Connection registry + message dispatch
│   ├── client.go             # Internal client: transport binding + decode
│   └── message.go            # Wire protocol message + Type constants
│
├── cmd/main.go               # Runnable example server
│
└── internal/
    ├── transport/conn.go     # Low-level WebSocket read/write/ping/pong
    ├── protocol/
    │   ├── codec.go          # Codec interface + JSONCodec implementation
    │   └── frame.go          # Frame message structure
    ├── runtime/
    │   ├── pending.go        # Pending call registry + result delivery
    │   └── op.go             # Operation type enumeration
    └── idgen/
        └── generator.go      # IDGenerator interface + UUIDGenerator
```

## Component Diagram

```
┌────────────────────────────────────────────┐
│  Engine (public API)                       │
│  ┌──────────────┐  ┌────────────────────┐  │
│  │ RouterGroup  │  │  HandlerFunc       │  │
│  │  - Group()   │  │  Middleware        │  │
│  │  - Use()     │  │                    │  │
│  │  - On()      │  │                    │  │
│  └──────┬───────┘  └────────────────────┘  │
│         │                                  │
│  ┌──────▼───────┐                          │
│  │  Router      │                          │
│  │  - dispatch()│                          │
│  └──────┬───────┘                          │
│         │                                  │
│  ┌──────▼───────┐                          │
│  │  Hub         │                          │
│  │  - Register()│                          │
│  │  - Dispatch()│                          │
│  │  - clients   │                          │
│  └──────┬───────┘                          │
│         │                                  │
│  ┌──────▼────────┐                         │
│  │  Client       │                         │
│  │  - onMessage()│                         │
│  │  - Runtime    │                         │
│  │  - Codec      │                         │
│  └──────┬────────┘                         │
│         │                                  │
│  ┌──────▼─────────┐                        │
│  │  transport.Conn│                        │
│  │  - readPump()  │                        │
│  │  - writePump() │                        │
│  │  - ping/pong   │                        │
│  └─────────────── ┘                        │
└────────────────────────────────────────────┘
```

## Messaging Protocol

Cira uses a JSON-based protocol over WebSocket text frames.

### Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `push` | Server → Client | One-way event, no response expected |
| `request` | Bidirectional | Request that expects a response |
| `response` | Bidirectional | Response correlated to a request via `reply_to` |

### Message Structure

```json
{
    "id": "uuid",
    "route": "event.name",
    "type": "push|request|response",
    "reply_to": "request-id",
    "data": {}
}
```

### Call Flow

```
Client                          Server
  │                               │
  │──── request(id=1,route="x")──▶│  ctx.Call("x", req, &resp)
  │                               │  ┌─ Register pending[1]
  │                               │  ├─ Send request
  │◀─── response(id=2,reply=1)────│  ├─ Resolve(1, data)
  │                               │  └─ Return result
```

## Key Design Decisions

### Router Dot-Separated Scoping

Routes use `.` as a prefix separator, inherited from `RouterGroup`:

```go
api := server.Group("api")
api.Group("user").On("list", handler)
// Full route: "api.user.list"
```

### Synchronous Call

`Context.Call()` blocks the calling goroutine until a response arrives or the timeout fires. The `Runtime` pending registry maps message IDs to result channels, enabling the correlation.

### Connection Lifecycle

1. HTTP upgrade → `transport.Conn`
2. `Hub.Register()` creates `client` with unique ID
3. Read pump decodes messages → `onMessage()` → `Dispatcher.Dispatch()`
4. `Router.dispatch()` creates `Context` and calls handler
5. On disconnect, `OnClose` callbacks fire and hub removes the entry

### Customization Points

| Option | Interface | Default |
|--------|-----------|---------|
| Codec | `protocol.Codec` | `JSONCodec` |
| ID Generator | `idgen.IDGenerator` | `UUIDGenerator` |
| Upgrader | `websocket.Upgrader` | Allow all origins |

```go
server := cira.New(
    cira.WithCodec(&customCodec{}),
    cira.WithIDGenerator(&snowflakeGenerator{}),
)
```
