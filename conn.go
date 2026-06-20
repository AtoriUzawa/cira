package cira

// Conn represents a single WebSocket client connection.
type Conn struct {
	client *client
	hub    *hub
	id     string
}

func newConn(c *client, h *hub) *Conn {
	return &Conn{
		client: c,
		hub:    h,
		id:     c.ID(),
	}
}

// ID returns the unique identifier of this connection.
func (c *Conn) ID() string {
	return c.id
}

// Do executes the given function within a new Context bound to this connection.
func (c *Conn) Do(fn func(*Context)) {
	ctx := newContext(c.client, c.hub.idGenerator, c.hub.codec)
	ctx.Conn = c
	fn(ctx)
}

// Close terminates the connection.
func (c *Conn) Close() {
	c.client.Close()
}

// OnClose registers a callback to be invoked when the connection is closed.
func (c *Conn) OnClose(fn func()) {
	c.client.OnClose(fn)
}
