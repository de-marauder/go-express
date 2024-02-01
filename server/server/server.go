package server

import (
	"fmt"
	"log"
	"net"
	"strings"
)

// max connection buffer size
const buf_size = 1024

// Request handlers return an empty interface and require a pointer to request struct and pointer to response struct
type HTTPRequestHandler func(req *HTTPRequest, res *HTTPResponse) interface{}

// A map of the route to handler and method
// Used for raute and method based lookups after registering handlers
type routeMap map[string]routeMapValue
type routeMapValue struct {
	method  string
	handler HTTPRequestHandler
}

// Define a structure for our server
type Server struct {
	protocol string
	addr     string
	listener net.Listener
	routeMap routeMap
}

// Create a constructor function for the Server
func NewHTTPServer() *Server {
	return &Server{
		protocol: "tcp",
		addr:     "",
		listener: nil,
		routeMap: make(routeMap),
	}
}

// An interface to implement methods for the server struct in an ExpresJs fashion
type ServerMethods interface {
	Listen(string, func())
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

func NewHTTPResponse() *HTTPResponse {
	r := &HTTPResponse{
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
	res.Body = body
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
func (s *Server) Get(route string, handler HTTPRequestHandler) {
	s.routeMap["GET-"+route] = routeMapValue{
		method:  "GET",
		handler: handler,
	}
}
func (s *Server) Post(route string, handler HTTPRequestHandler) {
	s.routeMap["POST-"+route] = routeMapValue{
		method:  "POST",
		handler: handler,
	}
}
func (s *Server) Put(route string, handler HTTPRequestHandler) {
	s.routeMap["PUT-"+route] = routeMapValue{
		method:  "PUT",
		handler: handler,
	}
}
func (s *Server) Patch(route string, handler HTTPRequestHandler) {
	s.routeMap["PATCH-"+route] = routeMapValue{
		method:  "PATCH",
		handler: handler,
	}
}
func (s *Server) Delete(route string, handler HTTPRequestHandler) {
	s.routeMap["DELETE"+route] = routeMapValue{
		method:  "DELETE",
		handler: handler,
	}
}

// Start a new tcp server that listens on the specified address
// Tell the listener to close just before the function scope is left
// Start an accept loop to receive requests
func (s *Server) start(cb func()) {
	defer s.listener.Close()
	fmt.Println("TCP Server listening on", s.addr, "...")
	go cb()
	s.acceptLoop(s.listener)
	fmt.Printf("Hey!\n")
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
	log.Println("Connection from ", conn.RemoteAddr())
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

// Log Error and exit server
func logError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// Parse and process request and provide response to client via connection
func (s *Server) handleHTTPRequest(conn net.Conn, message string) {
	// fmt.Println("--Request--")
	// fmt.Println(message)
	req := parseReqToStruct(message)
	res := NewHTTPResponse()
	res.Version = req.Version
	res.conn = conn

	rMap, ok := s.routeMap[req.Method+"-"+req.Route]
	setResHeaders(res)
	if !ok {
		// Extract params
		if !s.tryExtractParams(req, res) {	
			res.StatusCode = 404
			res.Send(fmt.Sprintln("Path ", req.Method, req.Route, "Not Found"))
		}
	} else {
		rMap.handler(req, res)
	}
}

// Handle param extraction when registered route elements are tokenized for dynamism
func (s *Server) tryExtractParams (req *HTTPRequest, res *HTTPResponse) bool {
		var params = make(Params)
		for k := range s.routeMap {
			r := strings.Split(k, "-")[1]

			match, ok := performRoutePatternMatch(req.Route, r, params)
			if !ok {
				continue
			} else {
				req.Params = params
				rMp := s.routeMap[req.Method+"-"+match]
				rMp.handler(req, res)
				return true
			}
		}
		return false
}

// Compares requested route with registered route to see if they match
func performRoutePatternMatch(reqR string, r string, p Params) (string, bool) {
	reqRSl := strings.Split(reqR, "/")
	rSl := strings.Split(r, "/")

	if len(reqRSl) != len(rSl) {
		return "", false
	}

	for i := range reqRSl {
		vR, vr := reqRSl[i], rSl[i]
		if isParamToken(vr) {
			p[strings.TrimLeft(vr, ":")] = vR
		} else if vR == vr {
			continue
		} else {
			return "", false
		}
	}
	return r, true
}

// Check if a route element is tokenized (i.e. prefixed with a ":")
func isParamToken(s string) bool {
	if s != "" {
		return strings.Split(s, "")[0] == ":"
	} else {
		return false
	}
}

// Set default response headers
func setResHeaders(res *HTTPResponse) {
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Server"] = "go-express"
	res.Headers = headers
}

// Convert HTTP request to a more usable sreuct form
// Struct type is HTTPRequest
func parseReqToStruct(message string) *HTTPRequest {
	msgSlice := strings.Split(message, "\r")

	var (
		scheme    string
		fullRoute string
		route     string
		query     Query
		method    string
		version   string
		headers   Headers
		body      interface{}
	)

	counter := 0
	headers = make(map[string]string)
	for lineNo, content := range msgSlice {
		// Read details from scheme (first line of HTTP request)
		if lineNo == 0 {
			scheme = content
			schemeSlice := strings.Fields(scheme)
			method = schemeSlice[0]
			fullRoute = schemeSlice[1]
			version = schemeSlice[2]
			counter++
			continue
		}
		// build headers
		contentSlice := strings.Split(content, ": ")

		// end loop after headers or if a header line cannot be split into a key value pair
		if content == "\r\n" || len(contentSlice) != 2 {
			break
		}
		headers[contentSlice[0]] = contentSlice[1]
		counter += 1
	}

	// Extract route and query
	fullRouteSlice := strings.Split(fullRoute, "?")
	if len(fullRouteSlice) == 2 {
		route = fullRouteSlice[0]
		if len(route) > 1 {
			route = strings.TrimRight(route, "/")
		}
		query = parseQueryToMap(fullRouteSlice[1])
	}

	// build body
	body = strings.Join(msgSlice[counter+1:], "\r\n")

	req := NewHTTPRequest()
	req.Version = version
	req.Route = route
	req.Query = query
	req.Method = method
	req.Headers = headers
	req.Body = body

	return req
}

// Convert struct type HTTPResponse to raw HTTP response string
func parseResStructToRaw(res *HTTPResponse) string {
	response := parseResponseStatusLine(res) + "\r\n" + parseHeadersToString(res.Headers) + "\r\n" + parseBody(res) + "\r\n"
	return response
}

// Convert map Headers to a string
func parseHeadersToString(headers Headers) string {
	var parsedHeaders string
	for key, val := range headers {
		parsedHeaders += key + ": " + val + "\r\n"
	}
	return parsedHeaders
}

// Builds the first line in the raw HTTP response (the status line)
func parseResponseStatusLine(res *HTTPResponse) string {
	resStatusLine := fmt.Sprint(res.Version) + " " + fmt.Sprint(res.StatusCode) + " " + HTTPStatusCodeMap[res.StatusCode]
	return resStatusLine
}

// convert response body interface to string
func parseBody(res *HTTPResponse) string {
	body := fmt.Sprint(res.Body)
	return body
}

// Stringify JSON
func parseJsonToString(json map[string]string) string {
	var j string = "\"{"
	fmt.Println(len(json))
	counter := 1
	for key, val := range json {
		if counter == len(json) {
			j += fmt.Sprintf("\\\"%v\\\":\\\"%v\\\"", key, val)
		} else {
			j += fmt.Sprintf("\\\"%v\\\":\\\"%v\\\",", key, val)
		}
		counter++
	}
	j += "}\""
	fmt.Println(j)
	return j
}

// Convert query string from the full route to a key value map
func parseQueryToMap(q string) Query {
	qSlice := strings.Split(q, "&")
	query := make(Query)
	for _, qu := range qSlice {
		quSlice := strings.Split(qu, "=")
		key := quSlice[0]
		val := ""
		if len(quSlice) == 2 {
			val = quSlice[1]
		}
		query[key] = val
	}
	return query
}
