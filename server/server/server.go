package server

import (
	"fmt"
	"log"
	"net"
	"strings"
)

// max connection buffer size
const buf_size = 1024

type NextFunction func()

// Request handlers return an empty interface and require a pointer to request struct and pointer to response struct
type HTTPRequestHandler func(req *HTTPRequest, res *HTTPResponse, next NextFunction)

// A map of the route to handler and method
// Used for raute and method based lookups after registering handlers
type routeMap map[string]routeMapValue
type routeMapValue struct {
	method   string
	midx     int
	handlers []HTTPRequestHandler
}

// Define a structure for our server
type Server struct {
	protocol    string
	addr        string
	listener    net.Listener
	routeMap    routeMap
	middlewares [][]HTTPRequestHandler
	router      *router
}

// Create a constructor function for the Server
func NewHTTPServer() *Server {
	router := NewRouter()

	return &Server{
		protocol:    "tcp",
		addr:        "",
		listener:    nil,
		routeMap:    router.routeMap,
		middlewares: router.middlewares,
		router:      router,
	}
}

type MiddlewareMethods interface {
	Use(...HTTPRequestHandler)
	UseRouter(string, router)
}

// An interface to implement methods for the server struct in an ExpresJs fashion
type ServerMethods interface {
	Listen(string, func())
	MiddlewareMethods
	HTTPMethods
}
type HTTPMethods interface {
	Get(string, HTTPRequestHandler)
	Post(string, HTTPRequestHandler)
	Put(string, HTTPRequestHandler)
	Patch(string, HTTPRequestHandler)
	Delete(string, HTTPRequestHandler)
}

type Headers map[string]string
type Params map[string]string
type Query map[string]string
type HTTPRequest struct {
	Version string
	Method  string
	Route   string
	Params  Params
	Query   Query
	Headers Headers
	Body    interface{}
}

// Constructor for HTTPRequest
func NewHTTPRequest() *HTTPRequest {
	return &HTTPRequest{
		Headers: nil,
		Body:    nil,
	}
}

type statusMessage map[int]string

var HTTPStatusCodeMap = statusMessage{
	200: "OK",
	201: "CREATED",
	400: "BAD REQUEST",
	401: "UNAUTHORIZED",
	403: "FORBIDDEN",
	404: "NOT FOUND",
	500: "INTERNAL SERVER ERROR",
	502: "SERVICE UNAVAILABLE",
}

type HTTPResponse struct {
	Version    string
	conn       net.Conn
	StatusCode int
	Headers    Headers
	Body       interface{}
}
type HTTPResponseInterface interface {
	Json(map[string]string)
	Send(map[string]string)
}

func NewHTTPResponse(conn net.Conn) *HTTPResponse {
	r := &HTTPResponse{
		Version: "HTTP/1.1",
		conn: conn,
		Headers: nil,
		Body:    nil,
	}
	return r
}

func (res *HTTPResponse) Json(json map[string]string) {
	res.Body = parseJsonToString(json)
	response := parseResStructToRaw(res)
	_, err := res.conn.Write([]byte(response))
	logError(err)
}

func (res *HTTPResponse) Send(body string) {
	if body != "" {
		res.Body = body
	}
	response := parseResStructToRaw(res)
	_, err := res.conn.Write([]byte(response))
	logError(err)
}

func (s *Server) Listen(addr string, cb func()) {
	ln, err := net.Listen(s.protocol, addr)
	logError(err)
	s.addr = addr
	s.listener = ln
	s.start(cb)
}

// HTTP Method handlers for registering routes and their corresponding handlers
func (s *Server) Get(route string, handlers ...HTTPRequestHandler) {
	s.router.Get(route, handlers...)
}
func (s *Server) Post(route string, handlers ...HTTPRequestHandler) {
	s.router.Post(route, handlers...)
}
func (s *Server) Put(route string, handlers ...HTTPRequestHandler) {
	s.router.Put(route, handlers...)
}
func (s *Server) Patch(route string, handlers ...HTTPRequestHandler) {
	s.router.Patch(route, handlers...)
}
func (s *Server) Delete(route string, handlers ...HTTPRequestHandler) {
	s.router.Delete(route, handlers...)
}

// register middleware that runs before calls it precedes
func (s *Server) Use(middlewareHandlers ...HTTPRequestHandler) {
	s.router.middlewares = append(s.router.middlewares, middlewareHandlers)
	s.middlewares = s.router.middlewares
}

// Attaches routes defined by a router to the server
func (s *Server) UseRouter(prefix string, router *router) {
	// unpack router to obtain middlewares and route maps
	// loop through route map and prefix route keys with `${prefix}`
	// update middleware indexes on route map values by incremeting by `len(s.middlewares)`
	// update server routeMap with all router route maps
	// concatenate router middlewares to server middlewares

	rmid := router.middlewares
	rmap := router.routeMap

	for key, value := range rmap {
		keySl := strings.Split(key, "-")
		method := keySl[0]
		route := keySl[1]
		// rewrite root routes to match only prefix
		if route == "/" {
			route = ""
		}
		newKey := method + "-" + prefix + route
		s.routeMap[newKey] = routeMapValue{
			midx:     value.midx + len(s.middlewares),
			method:   value.method,
			handlers: value.handlers,
		}
	}

	s.middlewares = append(s.middlewares, rmid...)

}

// Start a new tcp server that listens on the specified address
// Tell the listener to close just before the function scope is left
// Start an accept loop to receive requests
func (s *Server) start(cb func()) {
	defer s.listener.Close()
	fmt.Println("TCP Server listening on", s.addr, "...")
	go cb()
	s.acceptLoop(s.listener)
}

// Create a while loop to accept all incoming connections on a listener
// Handle all accepted connections as goroutines
func (s *Server) acceptLoop(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		logError(err)
		go s.handleConnection(conn)
	}
}

// Close the connection before leaving function scope
// Make a buffer to accept incoming message. The buffer is currently 1024 but it can be a large as required
// Display message in the console
// Relay response to HTTP response if client is an HTTP client
// Else send regular string
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Accept request
	buf := make([]byte, buf_size)
	n, err := conn.Read(buf)
	logError(err)
	message := string(buf[:n])
	// fmt.Printf("%s\n", message)

	// Respond
	if isHTTPRequest(message) {
		s.handleHTTPRequest(conn, message)
	} else {
		response := "Thanks for connecting, but I only understand HTTP"
		_, err = conn.Write([]byte(response))
		logError(err)
	}
}

// Check if message is an HTTP request
func isHTTPRequest(message string) bool {
	scheme := strings.Split(message, "\r")[0]
	return strings.Contains(scheme, "HTTP")
}

// Parse and process request and provide response to client via connection
func (s *Server) handleHTTPRequest(conn net.Conn, message string) error {
	defer func () {
		if r := recover(); r != nil {
			// println("defer - recover => ", r)
			InternalServerError(conn)
		}
	}()
	// fmt.Println("--Request--")
	// fmt.Println(message)
	req := parseReqToStruct(message)
	res := NewHTTPResponse(conn)
	

	log.Println("Connection from ", conn.RemoteAddr(), "-", req.Method, "-", req.Route)

	rMap, ok := s.routeMap[req.Method+"-"+req.Route]
	setResHeaders(res)
	if !ok {
		// Try to extract params or fail
		if !s.tryExtractParams(req, res) {
			res.StatusCode = 404
			res.Send(fmt.Sprintln("Path ", req.Method, req.Route, "Not Found"))
		}
	} else {
		allHandlers := concatenateAllHandlers(s, rMap)
		e := runHandlers(req, res, allHandlers)
		logError(e)
	}
	return nil
}

// Handle param extraction when registered route elements are tokenized for dynamism
func (s *Server) tryExtractParams(req *HTTPRequest, res *HTTPResponse) bool {
	var params = make(Params)
	for k := range s.routeMap {

		match, ok := performRoutePatternMatch(req, k, params)

		if !ok {
			continue
		} else {
			req.Params = params
			rMp := s.routeMap[req.Method+"-"+match]
			allHandlers := concatenateAllHandlers(s, rMp)
			e := runHandlers(req, res, allHandlers)
			logError(e)
			return true
		}
	}

	return false
}

// Handle control switching using a call to next when dealing with middlewares (multiple request handlers)
func runHandlers(req *HTTPRequest, res *HTTPResponse, handlers []HTTPRequestHandler) error {
	n := 0
	var next NextFunction = func() {
		n++
	}
	for i, h := range handlers {
		if n != i {
			return &Error{"Error in middleware chain. `next` not called in a middleware. Unable to continue"}
		}
		h(req, res, next)
	}
	return nil
}
