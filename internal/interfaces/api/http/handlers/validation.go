package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	e "github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

type Validator interface {
	Validate() map[string]string // returns field:error map
}

func validateRequestBody[T Validator](w http.ResponseWriter, r *http.Request) (*T, bool) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e.WriteHttpError(w, e.NewValidationError("Invalid request payload", nil))
		return nil, false
	}

	if errors := req.Validate(); len(errors) > 0 {
		e.WriteHttpError(w, e.NewValidationError("Validation failed", errors))
		return nil, false
	}

	return &req, true
}

func requirePathParams(w http.ResponseWriter, params map[string]string) bool {
	errors := make(map[string]string)
	for key, value := range params {
		if value == "" {
			errors[key] = "required"
		}
	}
	if len(errors) > 0 {
		e.WriteHttpError(w, e.NewValidationError("Missing required path parameters", errors))
		return false
	}
	return true
}

func validateDate(date time.Time, field string) map[string]string {
	errors := make(map[string]string)

	if date.IsZero() {
		errors[field] = "required"
		return errors
	}

	if date.Before(time.Now()) {
		errors[field] = "must be in the future"
		return errors
	}
	return nil
}
