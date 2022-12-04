// Filename: internal/validator/validator.go

package validator

import (
	"net/url"
	"regexp"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	PhoneRX = regexp.MustCompile(`^\+?\(?[0-9]{3}\)?\s?-\s?[0-9]{3}\s?-\s?[0-9]{4}$`)
)

// we create a type that will wrap our validation errors map
type Validator struct {
	Errors map[string]string
}

// New() creates a new validator instance
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// Create a method so can have access to validator
// valid() checks the errors map for entries
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// In() checks if an element can be found in a provided list of elements
// a utility function
func In(element string, list ...string) bool {
	for i := range list {
		if element == list[i] {
			return true
		}
	}
	return false
}

// checking if a regex string matches regex pattern using Matches() and then returns true if a string maches the pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// validwebsite() checks if the string value is a valid web url
func ValidWebsite(website string) bool {
	_, err := url.ParseRequestURI(website)
	return err == nil
}

// AddError() adds an error entry to the Errors map
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check() performs the validation checks and calls the AddError method in turn if an error entry need to be added
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Unique() it will check the entries in MODE to ensure no repeating valuse in the slice
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}
