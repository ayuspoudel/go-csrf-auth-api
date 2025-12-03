package middleware

import (
	"net/http"

	// We are using alice package because it creates pipeline for request - response
	"github.com/justinas/alice"
)

/*
NewHandler function is a request pipeline.
For all incoming http requests to this server, NewHandler will handle it in the flow of:
> **Request** -> recoverHandler --> AuthHandler --> LogicHandler --> **Response**
 1. Recover Handler
    - If a bug causes panic inside the code, without recovery the server shuts down
    - With recovery one request fails but server continues running
    - All other requests will also fail, but server will not die
 2. Auth Handler
    - Allow only authenticated users to pass protected routes
    - /restricted must check JWT + CSRF
    - /logout must check if user is logged in
    - /deleteUser must check if its the same user
 3. LogicHandler
    - do the actual job
    - /login (GET) show login page
    - /login (POST) authenticate user and set cookies
    - /register (GET) show register page
    - /register (POST) create user
    - /logout clear cookies
    - /deleteUser delete user from db
*/
func NewHandler() http.Handler {
	return alice.New(recoverHandler, authHandler).ThenFunc(logicHandler)
}

func recoverHandler(next http.Handler) http.Handler {

}

func authHandler(next http.Handler) http.Handler {

}

func logicHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/restricted":
	case "/login":
		switch r.Method {
		case "POST":
		case "GET":
		default:
		}
	case "/register":
		switch r.Method {
		case "POST":
		case "GET":
		default:
		}
	case "/logout":
	case "/deleteUser":
	default:
	}
}

// We need this for logout so we can nullify token cookies on logout
func nullifyTokenCookies(w *http.ResponseWriter, r *http.Request) {

}

// On signup or login we need to set auth and refresh cookies
func setAuthAndRefreshCookies() {

}

// For each protected auth path, we need this to grab CSRF from http request and validate it
func grabCsrfFromRequest(r *http.Request) string {

}
