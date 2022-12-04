//Filename: cmd/api/context.go

package main

import (
	"context"
	"net/http"

	"photoalbum.joelical.net/internal/data"
)

// Define a custom contextKey Type
type contextKey string

// make a user a key
const userContextKey = contextKey("user")

// create a Method to add user to the context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)

}

// retrieve the User struct
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return &user
}
