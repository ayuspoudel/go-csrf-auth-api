package main

import (
	"fmt"
	"net/http"

	"github.com/justinas/alice"
)

func main() {
	http.HandleFunc("/", helloHandler) // This handlerFunction must satisfy the HandleFunc means it must have w and r as parameters
	fmt.Println("Listening on port 9000")
	pipeline := alice.New(loggingMiddleware, loggingMiddleware2).ThenFunc(helloHandler)
	http.ListenAndServe(":9000", pipeline)
}

// If in alice.New(func1, fun2) next from func1, will forward request to func2
// This loggingMiddleware just logs request has been recieved
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc((func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Received request for %s\n", r.URL.Path)
		next.ServeHTTP(w, r)
	}))
}

func loggingMiddleware2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Logging Middleware 2: Received request for %s\n", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Write([]byte("Hello from Go Server built by Ayush"))
	case "/ayush":
		w.Write([]byte("Hello Ayush!"))
	}

}
