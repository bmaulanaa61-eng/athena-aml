package errors

import (
	"fmt"
	"net/http"
)

const (
	CodeValidationError          = "VALIDATION_ERROR"
	CodeInvalidStateTransition   = "INVALID_STATE_TRANSITION"
	CodeRuleConfigurationInvalid = "RULE_CONFIGURATION_INVALID"
	CodeAlertStateInvalid        = "ALERT_STATE_INVALID"
	CodeWebhookURLInvalid        = "WEBHOOK_URL_INVALID"

	CodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	CodeSessionInvalid       = "SESSION_INVALID"
	CodeSessionExpired       = "SESSION_EXPIRED"

	CodeAuthorizationDenied = "AUTHORIZATION_DENIED"

	CodeTenantNotFound      = "TENANT_NOT_FOUND"
	CodeResourceNotFound    = "RESOURCE_NOT_FOUND"
	CodeRuleNotFound        = "RULE_NOT_FOUND"
	CodeAlertNotFound       = "ALERT_NOT_FOUND"
	CodeTransactionNotFound = "TRANSACTION_NOT_FOUND"
	CodeWebhookNotFound     = "WEBHOOK_NOT_FOUND"
	CodeJobNotFound         = "JOB_NOT_FOUND"

	CodeResourceConflict     = "RESOURCE_CONFLICT"
	CodeIdempotencyConflict  = "IDEMPOTENCY_CONFLICT"
	CodeRuleConflict         = "RULE_CONFLICT"
	CodeTransactionDuplicate = "TRANSACTION_DUPLICATE"

	CodeRateLimited = "RATE_LIMITED"

	CodeInternalError         = "INTERNAL_ERROR"
	CodeWebhookDeliveryFailed = "WEBHOOK_DELIVERY_FAILED"
	CodeJobProcessingFailed   = "JOB_PROCESSING_FAILED"
	CodeAIProviderUnavailable = "AI_PROVIDER_UNAVAILABLE"
	CodeAIResponseInvalid     = "AI_RESPONSE_INVALID"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

func BadRequest(code, message string) *AppError {
	if code == "" {
		code = CodeValidationError
	}
	return New(code, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Autentikasi gagal atau API Key tidak valid."
	}
	return New(CodeAuthenticationFailed, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	if message == "" {
		message = "Akses ditolak untuk resource ini."
	}
	return New(CodeAuthorizationDenied, message, http.StatusForbidden)
}

func NotFound(code, message string) *AppError {
	if code == "" {
		code = CodeResourceNotFound
	}
	return New(code, message, http.StatusNotFound)
}

func Conflict(code, message string) *AppError {
	if code == "" {
		code = CodeResourceConflict
	}
	return New(code, message, http.StatusConflict)
}

func Internal(err error, message string) *AppError {
	if message == "" {
		message = "Terjadi kesalahan internal pada server."
	}
	return Wrap(err, CodeInternalError, message, http.StatusInternalServerError)
}
