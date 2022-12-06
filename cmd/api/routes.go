//Filename: cmd/api/routes.go

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// create a method that returns a http router
func (app *application) routes() http.Handler {
	//Create a new httrouter router instance
	router := httprouter.New()
	//implement error handling in router
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedesponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/photo", app.requirePermission("photo:read", app.listPhotoHandler))
	router.HandlerFunc(http.MethodPost, "/v1/photo", app.requirePermission("photo:write", app.createPhotoHandler))

	router.HandlerFunc(http.MethodGet, "/v1/photo/:id", app.requirePermission("photo:read", app.showPhotoHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/photo/:id", app.requirePermission("photo:write", app.updatePhotoHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/photo/:id", app.requirePermission("photo:write", app.deletePhotoHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
