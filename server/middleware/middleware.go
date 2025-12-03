package middleware

import (
	"net/http"

	"github.com/justinas/alice"
)

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

func nullifyTokenCookies(w *http.ResponseWriter, r *http.Request) {

}

func setAuthAndRefreshCookies() {

}

func grabCsrfFromRequest(r *http.Request) string {

}
