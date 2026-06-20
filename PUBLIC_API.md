# Cira Public API

## Core Types

### Engine
The top-level WebSocket server. Manages connections, routing, and HTTP integration.

```go
type Engine struct {
    *RouterGroup                // embedded: On, Group, Use
    OnConnect    func(*Conn)    // connection established callback
}
```

### Conn
Handle to a single WebSocket client connection.

```go
type Conn struct { /* unexported fields */ }
```

### Context
Per-message context. Provides messaging primitives and access to the current connection.

```go
type Context struct {
    ClientID string       // unique client identifier
    Message  *Message     // the incoming message being processed
    Timeout  time.Duration // timeout for Call operations (default: 10h)
    Conn     *Conn        // the connection this context belongs to
}
```

### RouterGroup
Scoped routing group with prefix and middleware chain.

```go
type RouterGroup struct { /* unexported fields */ }
```

### Router
Internal router. Exposed via `RouterGroup`.

```go
type Router struct { /* unexported fields */ }
```

### Message
Wire protocol message structure.

```go
type Message struct {
    ID      string          `json:"id"`
    Route   string          `json:"route"`
    Type    Type            `json:"type"`
    ReplyTo string          `json:"reply_to,omitempty"`
    Data    json.RawMessage `json:"data"`
}
```

### Type
Message type enumeration.

```go
type Type string
const (
    TypeReq  Type = "request"
    TypeResp Type = "response"
    TypePush Type = "push"
)
```

### HandlerFunc / Middleware
Function types for handlers and middleware.

```go
type HandlerFunc func(*Context)
type Middleware  func(HandlerFunc) HandlerFunc
```

### Option
Functional option for Engine construction.

```go
type Option func(*Engine)
```

---

## Core Methods

### Engine

| Method | Signature | Description |
|--------|-----------|-------------|
| `New` | `func New(opts ...Option) *Engine` | Create a new Engine |
| `Run` | `func (e *Engine) Run(addr string) error` | Start HTTP server, listen at `/ws` |
| `HandleWS` | `func (e *Engine) HandleWS(w http.ResponseWriter, r *http.Request)` | Upgrade HTTP to WebSocket (for custom mux) |
| `Conn` | `func (e *Engine) Conn(id string) (*Conn, error)` | Look up connection by ID |
| `WithIDGenerator` | `func WithIDGenerator(g idgen.IDGenerator) Option` | Custom ID generator option |
| `WithCodec` | `func WithCodec(c protocol.Codec) Option` | Custom codec option |
| `WithUpgrader` | `func WithUpgrader(u websocket.Upgrader) Option` | Custom upgrader option |

### RouterGroup

| Method | Signature | Description |
|--------|-----------|-------------|
| `Group` | `func (g *RouterGroup) Group(prefix string) *RouterGroup` | Create scoped sub-group |
| `Use` | `func (g *RouterGroup) Use(middlewares ...Middleware)` | Append middleware |
| `On` | `func (g *RouterGroup) On(event string, h HandlerFunc)` | Register event handler |

### Router

| Method | Signature | Description |
|--------|-----------|-------------|
| `On` | `func (r *Router) On(event string, h HandlerFunc)` | Register handler for event |

### Conn

| Method | Signature | Description |
|--------|-----------|-------------|
| `ID` | `func (c *Conn) ID() string` | Get connection ID |
| `Do` | `func (c *Conn) Do(fn func(*Context))` | Execute function in connection context |
| `Close` | `func (c *Conn) Close()` | Terminate connection |
| `OnClose` | `func (c *Conn) OnClose(fn func())` | Register close callback |

### Context

| Method | Signature | Description |
|--------|-----------|-------------|
| `Push` | `func (c *Context) Push(event string, data any) error` | Send one-way push event |
| `Req` | `func (c *Context) Req(event string, data any) error` | Send request event (no response wait) |
| `Resp` | `func (c *Context) Resp(data any)` | Send response to current message |
| `Call` | `func (c *Context) Call(route string, req any, resp any) error` | Send request and await response |

### Type

| Method | Signature | Description |
|--------|-----------|-------------|
| `String` | `func (t Type) String() string` | String representation |

---

## Sentinel Errors

| Error | Description |
|-------|-------------|
| `ErrNotFoundConn` | Connection not found by ID |
| `ErrCallTimeout` | Call exceeded Context.Timeout |
| `ErrConnNotFound` | Hub-level connection not found |

---

## Lifecycle

```
New Engine ──► Register handlers ──► Run(:addr)
                    │                      │
                    │   OnConnect callback │
                    │                      ▼
                    │              HandleWS (upgrade)
                    │                      │
                    │              ┌───────┴────────┐
                    │              │  transport.Conn│
                    │              │  readPump      │
                    │              │  writePump     │
                    │              └───────┬────────┘
                    │                      │ onMessage
                    │              ┌───────▼────────┐
                    │              │  client.decode │
                    │              │  → Dispatch    │
                    │              └───────┬────────┘
                    │                      │
                    ▼                      ▼
              Handler invoked      Response sent
              (via Router)         (via Context)
```

---

## Routing

Cira uses dot-separated route naming with prefix inheritance.

```
Engine (prefix: "")
├── Group("api")
│   ├── On("user.list", h1)     → route: "api.user.list"
│   └── Group("admin")
│       └── On("ban", h2)       → route: "api.admin.ban"
└── On("ping", h3)              → route: "ping"
```

---

## Middleware

Middleware wraps `HandlerFunc` and applies in reverse registration order (last added = outermost).

```go
g.Use(func(next cira.HandlerFunc) cira.HandlerFunc {
    return func(ctx *cira.Context) {
        // before
        next(ctx)
        // after
    }
})
```

Middleware is scoped: a `RouterGroup`'s middlewares only apply to handlers registered on that group.

---

## Push / Request / Response / Call Model

```
┌──────────────────────────────────────────────────┐
│  Push: Server ──push(event, data)──► Client      │
│         No response expected                     │
├──────────────────────────────────────────────────┤
│  Req:  Server ──request(event, data)──► Client   │
│         No response correlation                  │
├──────────────────────────────────────────────────┤
│  Resp: Client ──request──► Server                │
│         Server ──response(data)──► Client        │
│         Correlated via reply_to field            │
├──────────────────────────────────────────────────┤
│  Call: Server ──request(route, req)──► Client    │
│         Server waits for response                │
│         Server ◄──response(data)── Client        │
│         Blocking with timeout                    │
└──────────────────────────────────────────────────┘
```

### Message Flow

| Method | Wire Type | Wait for Response | Use Case |
|--------|-----------|-------------------|----------|
| `Push` | `push` | No | Broadcast, notification |
| `Req` | `request` | No | Fire-and-forget command |
| `Resp` | `response` | N/A | Reply to incoming request |
| `Call` | `request` | Yes (sync) | RPC-style query |
