package main

import (
	"fmt"
	// "go-express/server"
	"github.com/de-marauder/go-express/server/server"
	// "go-express/server/server"
)

const (
	host = "localhost"
	port = "7000"
)

func main() {
	s := server.NewHTTPServer()

	// register handlers with param tokens and HTTP methods
	s.Use(mid1)
	s.Get("/", middleware, handleRootRoute)
	s.Get("/foo", middleware2, handleFooRoute)
	s.Get("/foo/:id/bar/:id2", handleOneFooBarByIdRoute)
	s.Use(mid2, mid3)
	s.Get("/foo/:id/bar", handleOneFooBarRoute)
	s.Get("/foo/:id", handleOneFooRoute)
	s.Post("/foo", handlePostFooRoute)
	s.Put("/foo", handlePutFooRoute)

	s.Listen(host+":"+port, func() {
		// Add some string outputs or
		// run some jobs after server starts
		fmt.Println("Awaiting connections...")
	})
}

// Define handlers //

func middleware(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	type body struct {
		id string
	}
	b := body{
		id: "1",
	}
	fmt.Println("Inside middleware 1")
	req.Body = b
	next()
}
func middleware2(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	type body struct {
		id string
	}
	b := body{
		id: "2",
	}
	fmt.Println("Inside middleware 2")
	req.Body = b
	next()
}

func handleRootRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	res.StatusCode = 200
	// send response using res.Send()
	res.Send("You just hit the " + req.Method + "  / route")
}

func handleOneFooBarByIdRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo/:id/bar/:id2 route")
}
func handleOneFooBarRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo/:id/bar route")
}
func handleOneFooRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo/:id route")
}
func handleFooRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo route")
}

// Handle JSON response using res.Json
func handlePostFooRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	res.StatusCode = 200
	type JsonMap map[string]string
	data := make(JsonMap)
	data["message"] = "You just hit the " + req.Method + " /foo route"
	data["status"] = "success"
	res.Json(data)
}
func handlePutFooRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo route")
}

func mid1(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	fmt.Println("Inside mid 1")
	next()
}
func mid2(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	fmt.Println("Inside mid 2")
	next()
}
func mid3(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) {
	fmt.Println("Inside mid 3")
	next()
}
