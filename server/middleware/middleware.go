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

/*
In Alice's chain, authHandler is the middleware function which authorizes if user can go to
logic handler function or not with credentials provided.
The first credential is always the Auth Token, which comes from Auth Cookie.

STEP 1 - Auth Cookie Check
We will first look if the request contains Auth Cookie. This is the user's first proof of
being logged in. Browser will automatically send this cookie with every request, *only after
a successful login*. If this is missing it means
	- user never logged in ||
	- user logged out already ||
	- browser cleared cookies ||
	- tokens expired and browser removed it ||
	- attacker is trying to directly hit protected page
Then, we return a 401 Unauthorized Error. Because "no user session exists at all"
	- a missing auth token means authentication failure
	- we should not redirect to login page, because this will tell the attacker that this
	  endpoint exists and bring valid credentials to access it. This is called Endpoint Enumeration
We also nullify any stale cookies as a step, because it will remove all expired cookies in
the browser if they exist.
Then we return from the function authHandler, because if not the logic will move to
logic handler in alice's pipeline

STEP 2 - Refresh cookie Check
If Auth Token is present we move to this step. It checks if refresh token exists. Refresh
token is the long lived credential that allows broswer to "renew" the user's sessions
without having them login again and again.
If this is missing it means
	- user once had a session, but it can no longer be refreshed
	- the session has expired already
Then, we redirect to login in this case
	- because, if refresh token has expired, but auth token is there, this means user has to
		login to get new cookies. There is not auto-renew


STEP 3 - CSRF Token Extraction
Now that both AuthToken and RefreshToken are present, we must verify that the request
originated from a real user, not from an attacker.
For that we extract the CSRF token from cookie by extracting the X-CSRF-Token using
grabCsrfFromRequest() function. The CSRF token can
come from either:
    - A hidden form field:  r.FormValue("X-CSRF-Token")
    - A custom request header: r.Header.Get("X-CSRF-Token")
If the CSRF token is missing or incorrect, it indicates:
    - A Cross-Site Request Forgery attack attempt
    - A request triggered from a different website or malicious script
    - A forged POST/DELETE triggered without user's intention
    - A session inconsistency where browser did not attach correct CSRF token
*/

func authHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/restricted", "/logout", "/deleteUser":
			// Here we need to check if user is authenticated
			// If authenticated, pass to next, else return 401 unauthorized
			log.Println("In auth restricted section")
			AuthCookie, authErr := r.Cookie("AuthToken")
			if authErr == http.ErrNoCookie {
				log.Println("Unauthorized Attempt! no auth cookie")
				nullifyTokenCookies(&w, r)
				http.Error(w, http.StatusText(401), 401)
				return
			} else if authErr != nil {
				log.Panic("panic: %+v", authErr)
				nullifyTokenCookies(&w, r)
				http.Error(w, http.StatusText(500), 500)
				return
			}

			RefreshCookie, refreshErr := r.Cookie("RefreshToken")
			if refreshErr == http.ErrNoCookie {
				log.Println("Unauthorized Attempt! no auth cookie")
				nullifyTokenCookies(&w, r)
				http.Redirect(w, r, "/login", 302)
				return
			} else if refreshErr != nil {
				log.Panic("panic: %+v", refreshErr)
				nullifyTokenCookies(&w, r)
				http.Error(w, http.StatusText(500), 500)
				return
			}
			requestCsrfToken := grabCsrfFromRequest(r)
			log.Println(requestCsrfToken)

			authTokenString, refreshTokenString, csrfSecret, err := myJwt.checkAndRefreshToken(AuthCookie.Value, RefreshCookie.Value, requestCsrfToken)
			if err != nil {
				if err.Error() == "Unauthorized" {
					log.Println("Unauthorized Attempt! JWT's not valid")
					http.Error(w, http.StatusText(401), 401)
					return
				} else {
					log.Panic("panic: %+v", err)
					http.Error(w, http.StatusText(500), 500)
					return
				}
			}
			log.Println("Successfully recreated jwts")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			setAuthAndRefreshCookies(&w, authTokenString, refreshTokenString)
			w.Header().Set("X-CSRF-Token", csrfSecret)

		default:
		}
		next.ServeHttp(w, r)
	}
	return http.HandlerFunc(fn)
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
