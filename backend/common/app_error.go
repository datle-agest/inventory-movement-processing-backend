package common

import "net/http"

type AppError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

func ErrNotFound(msg string) *AppError {
	return &AppError{StatusCode: http.StatusNotFound, Code: "NOT_FOUND", Message: msg}
}

func ErrBadRequest(msg string) *AppError {
	return &AppError{StatusCode: http.StatusBadRequest, Code: "BAD_REQUEST", Message: msg}
}

func ErrConflict(msg string) *AppError {
	return &AppError{StatusCode: http.StatusConflict, Code: "CONFLICT", Message: msg}
}

func ErrInternal(msg string) *AppError {
	return &AppError{StatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: msg}
}
