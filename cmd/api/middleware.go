//Filename: cmd/api/middleware.go

package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"photoalbum.joelical.net/internal/data"
	"photoalbum.joelical.net/internal/validator"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}

		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	//create a client type
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	//launch a background Goroutine that removes old entries
	//from the clients map once every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			//lock the map before starting to clean
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {

			//get the ip address of the request.
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
			//Lock()
			mu.Lock()
			//check if ip address is in the map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(
					rate.Limit(app.config.limiter.rps),
					app.config.limiter.burst,
				)}
			}
			//update the last seen time of the client
			clients[ip].lastSeen = time.Now()
			//check if request allowed
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		} //end of enabled conditional
		next.ServeHTTP(w, r)
	})

}

// Authentication
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Add a "Vary: Authorization" header to the response
		//A note to caches that the response may vary
		w.Header().Add("Vary", "Authorization")

		//Retrieve the value of the Authorization header from the request
		authorizationHeader := r.Header.Get("Authorization")

		//If no authorization found, then we will create an anonymous user
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		//check if the provided Authorization header is in the right format
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		//Extract the token
		token := headerParts[1]
		//validate the token
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		//retrieve detail about the user
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		//Add the user information to the request context
		r = app.contextSetUser(r, user)

		//call the next handler in the chain
		next.ServeHTTP(w, r)

	})
}

// check for activated user
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//get the user
		user := app.contextGetUser(r)
		//check for anonymous user
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// check for activated user
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//get the user
		user := app.contextGetUser(r)
		//check if user is activated
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.requireAuthenticatedUser(fn)
}

// check for activated user
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//get the user
		user := app.contextGetUser(r)
		//get the permission slice for the user
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		//check for the permission
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}
		//OK
		next.ServeHTTP(w, r)
	})
	return app.requireAuthenticatedUser(fn)

}

// Enable CORS
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Add the "Vary:Origin" header
		w.Header().Add("Vary", "Origin")
		//Get the value of the request's origin's headers
		origin := r.Header.Get("Origin")
		//check to see if origin header present
		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					//set the Access-Control-ALlow-Origin header
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
