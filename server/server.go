package server

import (
	"fmt"
	"net/http"
)

func StartServer(hostName string, port string) error {
	host := hostName + ":" + port
	fmt.Printf("listening on: %s", host)

	handler := middleware.NewHandler()

	http.Handle("/", handler)
	return http.ListenAndServe(host, nil)

}
