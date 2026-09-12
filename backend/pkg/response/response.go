package response

import (
	"encoding/json"
	"errors"
	"net/http"

	appErrors "omnishield/pkg/errors"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func OK(w http.ResponseWriter, payload interface{}) {
	JSON(w, http.StatusOK, payload)
}

func Created(w http.ResponseWriter, payload interface{}) {
	JSON(w, http.StatusCreated, payload)
}

func Accepted(w http.ResponseWriter, payload interface{}) {
	JSON(w, http.StatusAccepted, payload)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	requestID := r.Header.Get("X-Request-ID")
	if reqIDCtx, ok := r.Context().Value("request_id").(string); ok && reqIDCtx != "" {
		requestID = reqIDCtx
	}

	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		JSON(w, appErr.StatusCode, ErrorResponse{
			Error: ErrorDetail{
				Code:      appErr.Code,
				Message:   appErr.Message,
				RequestID: requestID,
			},
		})
		return
	}

	JSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Code:      appErrors.CodeInternalError,
			Message:   "Terjadi kesalahan pada sistem internal server.",
			RequestID: requestID,
		},
	})
}
