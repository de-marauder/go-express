package main

import (
	"fmt"
	"go-express/server"
	// "github.com/de-marauder/go-express/server/server"
	// "go-express/server/server"
)

const (
	host         = "localhost"
	port         = "7000"
)

func main() {
	s := server.NewHTTPServer()

	// register handlers with param tokens and HTTP methods
	s.Get("/", middleware, handleRootRoute)
	s.Get("/foo", middleware2, handleFooRoute)
	s.Get("/foo/:id/bar/:id2", handleOneFooBarByIdRoute)
	s.Get("/foo/:id/bar", handleOneFooBarRoute)
	s.Get("/foo/:id", handleOneFooRoute)
	s.Post("/foo", handlePostFooRoute)
	s.Put("/foo", handlePutFooRoute)

	s.Listen(host+":"+port, func () {
		// Add some string outputs or
		// run some jobs after server starts
		fmt.Println("Awaiting connections...")
	})
}

// Define handlers //

func middleware(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	type body struct {
		id string
	}
	b := body{
		id: "1",
	}
	req.Body = b
	next()
	return 1
}
func middleware2(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	type body struct {
		id string
	}
	b := body{
		id: "2",
	}
	req.Body = b
	next()
	return 1
}

func handleRootRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	res.StatusCode = 200
	// send response using res.Send()
	res.Send("You just hit the " + req.Method + "  / route")
	return 1
}

func handleOneFooBarByIdRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo/:id/bar/:id2 route")
	return 1
}
func handleOneFooBarRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo/:id/bar route")
	return 1
}
func handleOneFooRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo/:id route")
	return 1
}
func handleFooRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo route")
	return 1
}

// Handle JSON response using res.Json
func handlePostFooRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	res.StatusCode = 200
	type JsonMap map[string]string
	data := make(JsonMap)
	data["message"] = "You just hit the " + req.Method + " /foo route"
	data["status"] = "success"
	res.Json(data)
	return 1
}
func handlePutFooRoute(req *server.HTTPRequest, res *server.HTTPResponse, next server.NextFunction) interface{} {
	res.StatusCode = 200
	res.Send("You just hit the " + req.Method + "  /foo route")
	return 1
}
