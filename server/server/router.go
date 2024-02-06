package server

type Router interface {
	HTTPMethods
	MiddlewareMethods
}

type router struct {
	middlewares [][]HTTPRequestHandler
	routeMap routeMap
	// params string
	// stack []string
}

func NewRouter () *router {
	return &router{
		middlewares: [][]HTTPRequestHandler{},
		routeMap: make(map[string]routeMapValue),
	}
}

// HTTP Method handlers for registering routes and their corresponding handlers
func (r *router) Get(route string, handlers ...HTTPRequestHandler) {
	midx := len(r.middlewares)

	r.routeMap["GET-"+route] = routeMapValue{
		method:   "GET",
		midx:     midx,
		handlers: handlers,
	}
}
func (r *router) Post(route string, handlers ...HTTPRequestHandler) {
	midx := len(r.middlewares)

	r.routeMap["POST-"+route] = routeMapValue{
		method:   "POST",
		midx:     midx,
		handlers: handlers,
	}
}
func (r *router) Put(route string, handlers ...HTTPRequestHandler) {
	midx := len(r.middlewares)

	r.routeMap["PUT-"+route] = routeMapValue{
		method:   "PUT",
		midx:     midx,
		handlers: handlers,
	}
}
func (r *router) Patch(route string, handlers ...HTTPRequestHandler) {
	midx := len(r.middlewares)

	r.routeMap["PATCH-"+route] = routeMapValue{
		method:   "PATCH",
		midx:     midx,
		handlers: handlers,
	}
}
func (r *router) Delete(route string, handlers ...HTTPRequestHandler) {
	midx := len(r.middlewares)

	r.routeMap["DELETE"+route] = routeMapValue{
		method:   "DELETE",
		midx:     midx,
		handlers: handlers,
	}
}

// register middleware that runs before calls it precedes
func (r *router) Use(middlewareHandlers ...HTTPRequestHandler) {
	r.middlewares = append(r.middlewares, middlewareHandlers)
}
