package common

import "net/http"

// ─── Application Error Code Constants ────────────────────────────────────────
// These codes are returned in the JSON response body for FE to handle specific UI logic.
// HTTP status codes remain in the response header for general error categorization.
const (
	// Success
	CodeSuccess = "SUCCESS"

	// 400 - Bad Request variants
	CodeInvalidInput        = "INVALID_INPUT"
	CodeInvalidDateFormat   = "INVALID_DATE_FORMAT"
	CodeInvalidPagination   = "INVALID_PAGINATION"
	CodeInvalidCSVFormat    = "INVALID_CSV_FORMAT"
	CodeInvalidCSVHeader    = "INVALID_CSV_HEADER"
	CodeFileRequired        = "FILE_REQUIRED"
	CodeFileEmpty           = "FILE_EMPTY"
	CodeFileMustBeCSV       = "FILE_MUST_BE_CSV"
	CodeInsufficientStock   = "INSUFFICIENT_STOCK"
	CodeInvalidMovementType = "INVALID_MOVEMENT_TYPE"

	// 401 - Unauthorized
	CodeUnauthorized = "UNAUTHORIZED"
	CodeInvalidToken = "INVALID_TOKEN"
	CodeMissingToken = "MISSING_TOKEN"

	// 403 - Forbidden
	CodeForbidden        = "FORBIDDEN"
	CodePermissionDenied = "PERMISSION_DENIED"

	// 404 - Not Found variants
	CodeNotFound         = "NOT_FOUND"
	CodeItemNotFound     = "ITEM_NOT_FOUND"
	CodeMovementNotFound = "MOVEMENT_NOT_FOUND"

	// 409 - Conflict variants
	CodeConflict             = "CONFLICT"
	CodeDuplicateSKU         = "DUPLICATE_SKU"
	CodeDuplicateTransaction = "DUPLICATE_TRANSACTION"

	// 500 - Internal Server Error
	CodeInternalError = "INTERNAL_ERROR"
)

// ─── AppError ────────────────────────────────────────────────────────────────

type AppError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) HttpStatusCode() int { return e.StatusCode }

func (e *AppError) GetCode() string { return e.Code }

// ─── Custom Error Constructors ───────────────────────────────────────────────
// Use these to create errors with specific error codes.

func NewCustomError(statusCode int, code, msg string) *AppError {
	return &AppError{StatusCode: statusCode, Code: code, Message: msg}
}

func NewBadRequestError(code, msg string) *AppError {
	return NewCustomError(http.StatusBadRequest, code, msg)
}
func NewNotFoundError(code, msg string) *AppError {
	return NewCustomError(http.StatusNotFound, code, msg)
}
func NewConflictError(code, msg string) *AppError {
	return NewCustomError(http.StatusConflict, code, msg)
}
func NewUnauthorizedError(code, msg string) *AppError {
	return NewCustomError(http.StatusUnauthorized, code, msg)
}
func NewForbiddenError(code, msg string) *AppError {
	return NewCustomError(http.StatusForbidden, code, msg)
}

// ─── Generic Error Constructors ──────────────────────────────────────────────
// Use these for general errors without specific codes.

func ErrBadRequest(msg string) *AppError   { return NewBadRequestError(CodeInvalidInput, msg) }
func ErrNotFound(msg string) *AppError     { return NewNotFoundError(CodeNotFound, msg) }
func ErrConflict(msg string) *AppError     { return NewConflictError(CodeConflict, msg) }
func ErrUnauthorized(msg string) *AppError { return NewUnauthorizedError(CodeUnauthorized, msg) }
func ErrForbidden(msg string) *AppError    { return NewForbiddenError(CodeForbidden, msg) }
func ErrInternal(msg string) *AppError {
	return NewCustomError(http.StatusInternalServerError, CodeInternalError, msg)
}
