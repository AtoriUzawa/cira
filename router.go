package cira

import "log"

// Router maps event names to their handler functions.
type Router struct {
	handlers map[string]HandlerFunc
}

func newRouter() *Router {
	return &Router{
		handlers: make(map[string]HandlerFunc),
	}
}

// On registers a handler for the given event.
func (r *Router) On(event string, h HandlerFunc) {
	r.handlers[event] = h
}

func (r *Router) dispatch(ctx *Context) bool {
	h, ok := r.handlers[ctx.Message.Route]
	if !ok {
		log.Printf(
			"[ws][route][dispatch] route: %s",
			ctx.Message.Route,
		)
		return false
	}

	go h(ctx)

	return true
}
