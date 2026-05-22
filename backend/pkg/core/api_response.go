package core

import "net/http"

type APIResponse struct {
	Code       int         `json:"code" example:"200"`
	Message    string      `json:"message,omitempty" example:"success"`
	Result     interface{} `json:"result,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func Success(data interface{}) *APIResponse {
	return &APIResponse{
		Code:   http.StatusOK,
		Result: data,
	}
}

func SuccessWithMessage(
	data interface{},
	message string,
) *APIResponse {

	return &APIResponse{
		Code:    http.StatusOK,
		Message: message,
		Result:  data,
	}
}

func SuccessWithPaging(
	data interface{},
	paging *Pagination,
) *APIResponse {

	return &APIResponse{
		Code:       http.StatusOK,
		Result:     data,
		Pagination: paging,
	}
}

func Fail(
	code int,
	message string,
) *APIResponse {

	return &APIResponse{
		Code:    code,
		Message: message,
	}
}
