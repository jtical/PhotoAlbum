//Filename: cmd/api/helpers.go

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"photoalbum.joelical.net/internal/validator"
)

// define a new type named envelope. empty interface means it can be any type
type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	//Use the "ParamsFromContext()" function to get the request from context as a slice
	params := httprouter.ParamsFromContext(r.Context())
	//Get the value of the "id" parameter
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

// we create our method write json to create responses
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	//convert our map into a JSON object
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	// add a newline to make the viewing on the terminal easier
	js = append(js, '\n')
	// add the headers
	for key, value := range headers {
		w.Header()[key] = value
	}
	//specify that we will be serving our responses using JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//write out the [] byte slice containing the json response body
	w.Write(js)
	return nil

}

// dst stores decoded struct
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	//set the size of request. use http.MaxBytesReader() to limit the size of the request body to 1mb
	maxBytes := 1_048_576
	//decode the request body into the target destinatiin
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshallTypeError *json.UnmarshalTypeError
		var invalidUnmarshallError *json.InvalidUnmarshalError

		//switch to check for the errors
		switch {
		//check for syntax errors
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON(at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		//Check for wrong types passed by the user
		case errors.As(err, &unmarshallTypeError):
			if unmarshallTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshallTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type(at charcater %d)", unmarshallTypeError.Offset)
		//empty body
		case errors.Is(err, io.EOF):
			return errors.New("body cannot be empty")
		//unmappable fields
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		// check for request being too large
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		//pass a non-nil pointer error
		case errors.As(err, &invalidUnmarshallError):
			panic(err)
		default:
			return err
		}
	}
	// call decode again
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// the readString() method returns a string value from the query parameters
// string or returns a default value if no matching key is found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	//Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// split comma separted values. readCSV() method splits a value into a slice bases on the comma separator
// if no matching key is found then the default value is returned
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	//get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	//split the string based on the "," delimeter
	return strings.Split(value, ",")
}

// the readInt() method converts a string value from the query string to an integer value.
// if the value cannot be converted to an integer then a validation error is added to the validation errors map
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	//get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	//perform the conversion to an integer
	intValue, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return intValue
}

// create a background method. accepts a function as its parameter
func (app *application) background(fn func()) {

	//Increment the WaitGroup counter
	app.wg.Add(1)

	//launch a goroutine which runs an anobymous function that sends the welcome function
	go func() {
		defer app.wg.Done()
		//recover from panics
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		//Execute fn
		fn()
	}()
}
