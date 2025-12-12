package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	// We are using alice package because it creates pipeline for request - response
	"github.com/ayuspoudel/go-csrf-auth-api/db"
	"github.com/ayuspoudel/go-csrf-auth-api/server/myJwt"
	"github.com/ayuspoudel/go-csrf-auth-api/server/templates"
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
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(500), 500)
				log.Panic("Recovered! Panic:%+v", err)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func authHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/restricted", "/logout", "/deleteUser":
			// Here we need to check if user is authenticated
			// If authenticated, pass to next, else return 401 unauthorized

		default:
		}
	}
}

func logicHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/restricted":
		csrfSecert := grabCsrfFromRequest(r)
		templates.RenderTemplates(w, "restricted", &templates.RestrictedPage{csrfSecert, "Hello Ayush!"})

	case "/login":
		switch r.Method {
		case "POST":
		case "GET":
		default:
		}
	case "/register":
		switch r.Method {
		case "POST":
			r.ParseForm()
			log.Printf(r.Form)
			_, uuid, err := db.FetchUserByUserName(strings.Join(r.Form["username"], ""))
			if err == nil {
				w.WriteHeader(http.StatusUnauthorized)
				log.Println("uuid: " + uuid + "already exists")
			} else {
				role := "user"
				uuid, err := db.StoreUser(strings.Join(r.Form["username"], ""), strings.Join(r.Form["password"], ""), role)
				if err != nil {
					http.Error(w, http.StatusText(500), 500)
					log.Println("uuid: " + uuid)
				}
				authTokenString, refreshTokenString, csrfSecret, err := myJwt.CreateNewTokens(uuid, role)
				if err != nil {
					http.Error(w, http.StatusText(500), 500)
				}
				setAuthAndRefreshCookies(&w, authTokenString, refreshTokenString)
				w.Header().Set("X-CSRF-Token", csrfSecret)
				w.WriteHeader(http.StatusOK)

			}
		case "GET":
			templates.RenderTemplates(w, "register", &templates.RegisterPage{false, ""})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)

		}
	case "/logout":
		nullifyTokenCookies(&w, r)
		http.Redirect(w, r, "/login", 302)

	case "/deleteUser":
	default:
	}
}

// We need this for logout so we can nullify token cookies on logout
func nullifyTokenCookies(w *http.ResponseWriter, r *http.Request) {
	authCookie := http.Cookie{
		Name:     "AuthToken",
		Value:    "",
		Expires:  time.Now().Add(-1000 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(*w, &authCookie)

	refreshCookie := http.Cookie{
		Name:     "RefreshToken",
		Value:    "",
		Expires:  time.Now().Add(-1000 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(*w, &refreshCookie)

	RefreshCookie, RefreshError := r.Cookie("RefreshToken")
	if RefreshError == http.ErrNoCookie {
		log.Printf("Refresh Cookie not found")
	} else if RefreshError != nil {
		log.Printf("Error while retrieving Refresh Cookie: %v", RefreshError)
		http.Error(*w, http.StatusText(500), 500)
	}
	log.Printf("Refresh Cookie after nullify: %v", RefreshCookie)
	myJwt.RevokeRefreshToken(RefreshCookie.Value)
}

// On signup or login we need to set auth and refresh cookies
func setAuthAndRefreshCookies(w *http.ResponseWriter, authToken, refreshToken string) {
	authCookie := http.Cookie{
		Name:     "AuthToken",
		Value:    authToken,
		HttpOnly: true,
	}
	http.SetCookie(*w, &authCookie)

	refreshCookie := http.Cookie{
		Name:     "RefreshToken",
		Value:    refreshToken,
		HttpOnly: true,
	}
	http.SetCookie(*w, &refreshCookie)
}

// For each protected auth path, we need this to grab CSRF from http request and validate it
func grabCsrfFromRequest(r *http.Request) string {
	csrfFromForm := r.FormValue("X-CSRF-Token")

	if csrfFromForm != "" {
		return csrfFromForm
	} else {
		csrfFromHeader := r.Header.Get("X-CSRF-Token")
		return csrfFromHeader
	}
}
