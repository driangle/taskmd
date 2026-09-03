package main

import "fmt"

// TODO: add rate limiting to API endpoints
func handleRequest() {
	fmt.Println("handling request")
}

// FIXME: connection leak on error path - db connections not closed
func queryDatabase() error {
	return nil
}

// TODO: implement graceful shutdown
func startServer() {
	fmt.Println("server started")
}
