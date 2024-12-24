package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// WriteBusinessError writes a standardized error response
func WriteBusinessError(w http.ResponseWriter, err error, mapErrorToAPIError func(err error) *APIError) {
	fmt.Printf("Error: %v\n", err)
	var apiError *APIError
	if !errors.As(err, &apiError) && mapErrorToAPIError != nil {
		apiError = mapErrorToAPIError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.Status)
	json.NewEncoder(w).Encode(apiError)
}

func WriteHttpError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}

// MatchError checks if err matches any of the provided errors
func MatchError(err error, errs ...error) bool {
	for _, e := range errs {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
