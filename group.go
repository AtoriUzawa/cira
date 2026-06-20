package cira

// RouterGroup provides scoped routing with an optional prefix and middleware chain.
type RouterGroup struct {
	prefix string

	router *Router

	middlewares []Middleware
}

func joinPrefix(
	parent string,
	child string,
) string {
	switch {
	case parent == "":
		return child

	case child == "":
		return parent

	default:
		return parent + "." + child
	}
}

// Group creates a new sub-group with the given prefix and inherits the parent's middlewares.
func (g *RouterGroup) Group(
	prefix string,
) *RouterGroup {
	return &RouterGroup{
		prefix: joinPrefix(
			g.prefix,
			prefix,
		),

		router: g.router,

		middlewares: append(
			[]Middleware(nil),
			g.middlewares...,
		),
	}
}

// Use adds middlewares to the group's middleware chain.
func (g *RouterGroup) Use(
	middlewares ...Middleware,
) {
	g.middlewares = append(
		g.middlewares,
		middlewares...,
	)
}

// On registers a handler for the given event, wrapping it with the group's middlewares.
func (g *RouterGroup) On(
	event string,
	handler HandlerFunc,
) {
	fullRoute := joinPrefix(
		g.prefix,
		event,
	)

	for i := len(g.middlewares) - 1; i >= 0; i-- {
		handler = g.middlewares[i](handler)
	}

	g.router.On(
		fullRoute,
		handler,
	)
}
