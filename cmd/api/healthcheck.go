//Filename: cmd/api/healthcheck.go

package main

import (
	"net/http"
)

// create a healthcheck handler
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	//create a map to hold healthcheck data
	//we can make it an enevlope as its a map
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"enviornment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
