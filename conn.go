package cira

// Conn represents a single WebSocket peer connection.
type Conn struct {
	peer *peer
	hub  *hub
	id   string
}

func newConn(p *peer, h *hub) *Conn {
	return &Conn{
		peer: p,
		hub:  h,
		id:   p.ID(),
	}
}

// ID returns the unique identifier of this connection.
func (c *Conn) ID() string {
	return c.id
}

// Do executes the given function within a new Context bound to this connection.
func (c *Conn) Do(fn func(*Context)) {
	ctx := newContext(c.peer, c.hub.idGenerator, c.hub.codec)
	ctx.Conn = c
	fn(ctx)
}

// Close terminates the connection.
func (c *Conn) Close() {
	c.peer.Close()
}

// OnClose registers a callback to be invoked when the connection is closed.
func (c *Conn) OnClose(fn func()) {
	c.peer.OnClose(fn)
}
