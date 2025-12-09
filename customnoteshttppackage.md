# Go net/http package

> Author: Ayush Poudel

The net/http package in go provides everything needed to build http servers and clients in go

### Key Types 

Some key types used in this API are:
1. http.Handler     Interface for anything that handles HTTP Requests
2. http.HandlerFunc Function Adaptor to become a handler
3. http.Request    Contains all info about incoming request
4. http.ResponseWriter Used to send repsonse back to client
5. http.Cookie Represents a cookie (name, value, expiriy, flags)
6. http.Client  Used to make HTTP requests
7. http.Server  Represents an HTTP server


### BackBone of HTTP package

```go
type Handler interface{
    ServeHTTP(ResponseWriter, *Request)
}
```

If you want anything to handle an HTTP request → it must implement ServeHTTP.
Example: 

```go
type MyHandler struct{}
func (h MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello")
}
```

Use it with:
```go
http.ListenAndServe(":8080", MyHandler{})
```
