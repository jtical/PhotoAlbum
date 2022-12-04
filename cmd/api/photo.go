//Filename: cmd/api/photo.go

package main

import (
	"errors"
	"fmt"
	"net/http"

	"photoalbum.joelical.net/internal/data"
	"photoalbum.joelical.net/internal/validator"
)

// createPhotoHandler for the POST /v1/photo endpoint
func (app *application) createPhotoHandler(w http.ResponseWriter, r *http.Request) {
	//Our target decode destination
	var input struct {
		Title       string `json:"title"`
		Photo       string `json:"photo"`
		Description string `json:"description"`
	}
	//Initalize a new json.decoder instance
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//copy the values from the input struct to a new photo struct
	photo := &data.Photo{
		Title:       input.Title,
		Photo:       input.Photo,
		Description: input.Description,
	}
	//Initialize a new validator instance
	v := validator.New()

	//check the map to determine if there were any validator errors
	if data.ValidatePhoto(v, photo); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// create a photo record
	err = app.models.Photo.Insert(photo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// create a location header for the newly created resource
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/photo/%d", photo.ID))
	// write the response with 201 -created status code with the body being the photo data and the header being the headers map
	err = app.writeJSON(w, http.StatusCreated, envelope{"photo": photo}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// showPhotoHandler for the GET /v1/photo/:id endpoint
func (app *application) showPhotoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	//Fetch the specific list
	photo, err := app.models.Photo.Get(id)
	//handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//write the data returned by get
	err = app.writeJSON(w, http.StatusOK, envelope{"photo": photo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// updateListHandler for the "PUT /v1/list/:id" endpoint
func (app *application) updatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	//this method does a partial replacement
	//get the id for the list that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	//fetch the orginal record from the database
	photo, err := app.models.Photo.Get(id)
	//handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//create an input struct to hold data read in from the user
	// our target decode destination
	//update input struct to use pointers because pointers have a default value of nil
	//if the filed remains nil, then we know user did not update it
	var input struct {
		Title       *string `json:"title"`
		Photo       *string `json:"photo"`
		Description *string `json:"description"`
	}
	//initialize a new json.decode instance
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//check input struct for those updates
	//check input struct for those updates
	if input.Title != nil {
		photo.Title = *input.Title
	}
	if input.Photo != nil {
		photo.Photo = *input.Photo
	}
	if input.Description != nil {
		photo.Description = *input.Description
	}
	//perform validation on the updated photo record. if validation fails, then we send a 422 - unprocessable entity response to the user
	//Initialize a new validator instance
	v := validator.New()

	//check the map to determain if there were any validation errors
	if data.ValidatePhoto(v, photo); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	//pass the updated list record to the update() method
	err = app.models.Photo.Update(photo)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//write the data returned by get()
	err = app.writeJSON(w, http.StatusOK, envelope{"photo": photo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deletePhotoHandler for the "DELETE /v1/list/:id" endpoint
func (app *application) deletePhotoHandler(w http.ResponseWriter, r *http.Request) {
	//gets the id for the list that will be deleted
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	//delete the list from the database. sends a 404 not found status code to the user if there is no matching record.
	err = app.models.Photo.Delete(id)
	//handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//return a 200 status ok to the user with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "photo record successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// the listPhotoHandler() allows the user to see a lists of photo based on a set of criteria
func (app *application) listPhotoHandler(w http.ResponseWriter, r *http.Request) {
	//create an input struct to hold our query parameters
	var input struct {
		Title       string
		Photo       string
		Description string
		data.Filters
	}
	//Initialize a validator
	v := validator.New()
	//Get the URL values map
	qs := r.URL.Query()
	// use the helper methods to extract the values
	input.Title = app.readString(qs, "title", "")
	input.Photo = app.readString(qs, "photo", "")
	input.Description = app.readString(qs, "description", "")
	//get the page information
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	//get the sort information
	input.Filters.Sort = app.readString(qs, "sort", "id")
	//specify the allowed sort values
	input.Filters.SortList = []string{"id", "title", "description", "-id", "-title", "-description"}
	//chek for validation error
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	//get a listing of all photos
	photos, metadata, err := app.models.Photo.GetAll(input.Title, input.Photo, input.Description, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	//Send a JSON response containing all the schools
	err = app.writeJSON(w, http.StatusOK, envelope{"photos": photos, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
