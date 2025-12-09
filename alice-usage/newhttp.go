package main

import (
	"fmt"
	"net/http"
)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Write([]byte("Hello from go server"))
	}
}
func ayushHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ayush":
		w.Write([]byte("Hello from ayush server"))
	}
}
func notmain() {
	fmt.Println("Listening at port 9000")
	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/ayush", ayushHandler)
	http.ListenAndServe(":9000", nil)
}
