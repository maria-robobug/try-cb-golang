package main

import (
	"net/http"

	"github.com/couchbaselabs/try-cb-golang/server"
)

func main() {
	// Connect to Couchbase
	db, err := server.NewCBRepository()
	if err != nil {
		panic(err)
	}

	server := server.New(db)

	// Set up our routing
	http.Handle("/", server)

	// Listen on port 8080
	http.ListenAndServe(":8080", server)
}
