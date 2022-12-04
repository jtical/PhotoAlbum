//Filename: internal/data/photo.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"photoalbum.joelical.net/internal/validator"
)

// holds entries informatiom
// back tick character(struct tags) shows how the key should be formated
type Photo struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Title       string    `json:"title"`
	Photo       string    `json:"photo"`
	Description string    `json:"description"`
	Version     int32     `json:"version"`
}

func ValidatePhoto(v *validator.Validator, photo *Photo) {
	// use the check() method to execute our validation checks
	//check the map to determain if there were any validation errors
	v.Check(photo.Title != "", "name", "must be provided")
	v.Check(len(photo.Title) <= 200, "name", "must not be more than 200 bytes long")

	v.Check(photo.Photo != "", "photo", "must be provided")
	v.Check(len(photo.Photo) <= 200, "photo", "must not be more than 200 bytes long")

	v.Check(photo.Description != "", "description", "must be provided")
	v.Check(len(photo.Description) <= 800, "description", "must not be more than 800 bytes long")

}

// define a ListModel which wraps a sql.db connection pool
type PhotoModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new photo
func (m PhotoModel) Insert(photo *Photo) error {
	query := `
		INSERT INTO photos (title, photo, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version
	`
	// Create a context. time starts when context is created
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// cleanup to prevent memory leaks
	defer cancel()
	// collect the data fields into a slice
	args := []interface{}{
		photo.Title,
		photo.Photo,
		photo.Description,
	}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&photo.ID, &photo.CreatedAt, &photo.Version)
}

// Get() allows us to get a specific photo
func (m PhotoModel) Get(id int64) (*Photo, error) {
	//ensure that there is a valid id
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	//create the query
	query := `
		SELECT id, created_at, title, photo, description ,version
		FROM photos
		WHERE id = $1
	`
	//declare a list variable to hold the returned data
	var photo Photo
	//Create a context. time starts when context is created
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//cleanup to prevent memory leaks
	defer cancel()
	//execute the query using QueryRowcontext
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&photo.ID,
		&photo.CreatedAt,
		&photo.Title,
		&photo.Photo,
		&photo.Description,
		&photo.Version,
	)
	//handle any errors
	if err != nil {
		//check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	//success
	return &photo, nil
}

// Update() allows us to edit/alter a specific photo
// optimistic locking on version #
func (m PhotoModel) Update(photo *Photo) error {
	//create a query using the newly updated data
	query := `
		UPDATE photos
		SET title = $1,
			photo = $2,
			description = $3,
			version = version + 1
		WHERE id = $4
		AND version = $5 
		RETURNING version
	`
	//Create a context. time starts when context is created
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//cleanup to prevent memory leaks
	defer cancel()
	args := []interface{}{
		photo.Title,
		photo.Photo,
		photo.Description,
		photo.ID,
		photo.Version,
	}
	//check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&photo.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete() removes a specific photo
func (m PhotoModel) Delete(id int64) error {
	//check if the id exist
	if id < 1 {
		return ErrRecordNotFound
	}
	//create the delete query
	query := `
		DELETE FROM photos
		WHERE id = $1
	`

	//Create a context. time starts when context is created
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//cleanup to prevent memory leaks
	defer cancel()
	//execute the query
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	//check how many rows affected by the delte operation. we will use the RowsAffected() on the result variable
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	//check to see if zero rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// the GetAll() method returns a list of all the list sorted by id
func (m PhotoModel) GetAll(title string, photo string, description string, filters Filters) ([]*Photo, Metadata, error) {
	//construct the query to return all photos
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, created_at, title, photo, description, version
		FROM photos
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) or $1 = '')
		AND (to_tsvector('simple', photo) @@ plainto_tsquery('simple', $2) or $2 = '')
		AND (to_tsvector('simple', description) @@ plainto_tsquery('simple', $3) or $3 = '')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortOrder())

	//create a 3 second timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//execute the query
	args := []interface{}{title, photo, description, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	//defer the closing of the result set
	defer rows.Close()
	totalRecords := 0
	// initialize an empty slce to hold the photo data
	photos := []*Photo{}
	//iterate over the rows in the resultset
	for rows.Next() {
		var photo Photo
		//scan the values from the row into photo
		err := rows.Scan(
			&totalRecords,
			&photo.ID,
			&photo.CreatedAt,
			&photo.Title,
			&photo.Photo,
			&photo.Description,
			&photo.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		//add the Photo to our slice
		photos = append(photos, &photo)
	}
	//check for errors after looping through the resultset
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	//Returm the slice  of photos
	return photos, metadata, nil
}
