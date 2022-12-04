//Filename: cmd/api/errors.go

package main

import (
	"fmt"
	"net/http"
)

// create method to log errors
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

// we want to display json formated error message
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	//created json resposne
	env := envelope{"error": message}
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// server error response
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	//log the error
	app.logError(r, err)
	//prepare the response with error
	message := "the server encounter a problem and count not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// not found response
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	//create error message
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// a method not allowed response
func (app *application) methodNotAllowedesponse(w http.ResponseWriter, r *http.Request) {
	//create error message
	message := fmt.Sprintf("the %s method is not allowed for this resource", r.Method) // use formated string to allow message to be dynamic
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// a method for bad request. user provided a bad request
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	//create our message
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// User provided invalid values -- validation errors
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// edit conflict error
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// Exceeded number of request. rate limit error
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)

}

// Invalid credentials
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// Invalid token
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// Unauthorized access
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// Users who have not activated their accounts
func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)

}
