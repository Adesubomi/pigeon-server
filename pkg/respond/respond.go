package respond

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/adesubomi/pigeon-server/pkg/apperr"
)

type successResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type errorResponse struct {
	Message string            `json:"message"`
	Code    string            `json:"code"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, successResponse{Message: "Success", Data: data})
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, successResponse{Message: "Success", Data: data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, err error) {
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		appErr = apperr.Internal(err)
	}

	JSON(w, appErr.Status, errorResponse{
		Message: appErr.Message,
		Code:    appErr.Code,
		Errors:  appErr.Errors,
	})
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
