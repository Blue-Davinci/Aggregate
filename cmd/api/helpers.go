package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/jsonlog"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Define an envelope type.
type envelope map[string]any

// Perform a quick marshal instead of a marshalIndent for a tiny more speed
// as we will use this for converting data to json for our payment
// operations
func (app *application) covertToByteArray(data interface{}) ([]byte, error) {
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, err
	}
	return js, nil
}

func (app *application) returnEnvInfo() envelope {
	return envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}
}

// writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')
	// At this point, we know that we won't encounter any more errors before writing the
	// response, so it's safe to add any headers that we want to include.
	for key, value := range headers {
		w.Header()[key] = value
	}
	// Add the "Content-Type: application/json" header, then write the status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	// Initialize the json.Decoder, and call the DisallowUnknownFields() method on it
	// before decoding. This means that if the JSON from the client now includes any
	// field which cannot be mapped to the target destination, the decoder will return
	// an error instead of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	// Decode the request body to the destination.
	err := dec.Decode(dst)
	err = app.jsonReadAndHandleError(err)
	if err != nil {
		return err
	}
	// Call Decode() again, using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a single JSON value this will
	// return an io.EOF error. So if we get anything else, we know that there is
	// additional data in the request body and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (app *application) readJSONFromReader(reader io.Reader, dst any) error {
	// Use a limited reader to restrict the size of the body to 1MB.
	maxBytes := 1_048_576
	limitedReader := io.LimitReader(reader, int64(maxBytes))

	// Initialize the json.Decoder and disallow unknown fields.
	dec := json.NewDecoder(limitedReader)
	dec.DisallowUnknownFields()

	// Decode the JSON data to the destination.
	err := dec.Decode(dst)
	err = app.jsonReadAndHandleError(err)
	if err != nil {
		return err
	}
	// Call Decode() again, using a pointer to an empty anonymous struct as the destination.
	// If the request body only contained a single JSON value this will return an io.EOF error.
	// So if we get anything else, we know that there is additional data in the request body
	// and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) jsonReadAndHandleError(err error) error {
	if err != nil {
		// Vars to carry our errors
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		// Add a new maxBytesError variable.
		var maxBytesError *http.MaxBytesError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format "json: unknown
		// field "<name>"". We check for this, extract the field name from the error,
		// and interpolate it into our custom error message. Note that there's an open
		// issue at https://github.com/golang/go/issues/29035 regarding turning this
		// into a distinct error type in the future.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		// Use the errors.As() function to check whether the error has the type
		// *http.MaxBytesError. If it does, then it means the request body exceeded our
		// size limit of 1MB and we return a clear error message.
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	return nil
}

// The background() helper accepts an arbitrary function as a parameter.
func (app *application) background(fn func()) {
	app.wg.Add(1)
	// Launch a background goroutine.
	go func() {
		//defer our done()
		defer app.wg.Done()
		// Recover any panic.
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		// Execute the arbitrary function that we passed as the parameter.
		fn()
	}()
}

// The readString() helper returns a string value from the query string, or the provided
// default value if no matching key could be found.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Extract the value for a given key from the query string. If no key exists this
	// will return the empty string "".
	s := qs.Get(key)
	// If no key exists (or the value is empty) then return the default value.
	if s == "" {
		return defaultValue
	}
	// Otherwise return the string.
	return s
}

// The readInt() helper reads a string value from the query string and converts it to an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record an
// error message in the provided Validator instance.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	// Extract the value from the query string.
	s := qs.Get(key)
	// If no key exists (or the value is empty) then return the default value.
	if s == "" {
		return defaultValue
	}
	// Try to convert the value to an int. If this fails, add an error message to the
	// validator instance and return the default value.
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	// Otherwise, return the converted integer value.
	return i
}

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and an error.
func (app *application) readIDIntParam(r *http.Request, parameterName string) (int64, error) {
	// We can then use the ByName() method to get the value of the "id" parameter from
	// the slice. So we try to convert it to a base 10 integer (with a bit size of 64).
	// If the parameter couldn't be converted, or is less than 1, we know the ID is invalid
	params := chi.URLParam(r, parameterName)
	id, err := strconv.ParseInt(params, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid i-id parameter")
	}
	return id, nil
}

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return a nil UUID and an error.
func (app *application) readIDParam(r *http.Request, parameterName string) (uuid.UUID, error) {
	// We use chi's URLParam method to get our ID

	params := chi.URLParam(r, parameterName)
	feedID, isvalid := data.ValidateUUID(params)
	if !isvalid {
		return uuid.Nil, errors.New("invalid u-id parameter")
	}
	return feedID, nil
}

func (app *application) readIDStrParam(r *http.Request, parameterName string) (string, error) {
	params := chi.URLParam(r, parameterName)
	app.logger.PrintInfo(fmt.Sprintf("Raw Param: %s", params), nil)

	// URL-decode the parameter value
	decodedParam, err := url.QueryUnescape(params)
	if err != nil {
		app.logger.PrintInfo("Decoding failed", nil)
		return "", errors.New("invalid st-id-de parameter")
	}
	app.logger.PrintInfo(fmt.Sprintf("Decoded Param: %s", decodedParam), nil)

	if decodedParam == "" {
		return "", errors.New("invalid st-id parameter")
	}
	return decodedParam, nil
}

func (app *application) readIDFromQuery(r *http.Request, parameterName string) (uuid.UUID, error) {
	idParam := r.URL.Query().Get(parameterName)
	result, isValid := data.ValidateUUID(idParam)
	if !isValid {
		return uuid.Nil, errors.New("invalid id parameter")
	}
	return result, nil
}

func (app *application) returnEndDate(duration string, start_date time.Time) time.Time {
	switch {
	case duration == "day":
		return start_date.AddDate(0, 0, 1)
	case duration == "week":
		return start_date.AddDate(0, 0, 7)
	case duration == "month":
		return start_date.AddDate(0, 1, 0)
	case duration == "year":
		return start_date.AddDate(1, 0, 0)
	default:
		return start_date
	}
}

func (app *application) formatDate(date string) string {
	// Parse the date string to a time.Time value.
	d, err := time.Parse(time.RFC3339, date)
	if err != nil {
		app.logger.PrintError(err, nil)
		return ""
	}
	// Return the date formatted in the "YYYY-MM-DD HH:MM:SS" layout.
	return d.Format("2006-01-02 15:04:05")
}

// getEnvPath returns the path to the .env file based on the current working directory.
func getEnvPath(logger *jsonlog.Logger) string {
	dir, err := os.Getwd()
	if err != nil {
		logger.PrintFatal(err, nil)
		return ""
	}
	if strings.Contains(dir, "cmd/api") || strings.Contains(dir, "cmd") {
		return ".env"
	}
	return filepath.Join("cmd", "api", ".env")
}
