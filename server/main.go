package main

import (
	"fmt"
	"github.com/de-marauder/go-express/server/server"
)

const (
	host = "localhost"
	port = "7000"
)

func main() {
	s := server.NewHTTPServer()

	// create new routers
	r1 := server.NewRouter()
	r2 := server.NewRouter()

	// register handlers with param tokens and HTTP methods
	s.Use(mid1) // attaches a global middleware
	s.Get("/", middleware, handleRootRoute)

	// add router scoped handlers
	r1.Get("/", middleware2, handleFooRoute)
	r1.Get("/:id/bar/:id2", handleOneFooBarByIdRoute)

	r2.Use(mid2, mid3) // attaches middlewares scoped to router structure
	r2.Get("/:id/bar", handleOneFooBarRoute)
	s.Use(middleware2)
	r2.Get("/:id", handleOneFooRoute)
	r2.Post("/", handlePostFooRoute)
	r2.Put("/", handlePutFooRoute)

	// register routers
	s.UseRouter("/foo", r1)
	s.UseRouter("/bar", r2)

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
