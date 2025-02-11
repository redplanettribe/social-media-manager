package handlers

import (
	"encoding/json"
	"net/http"

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
