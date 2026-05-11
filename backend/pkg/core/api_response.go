package core

import "net/http"

type APIResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message,omitempty"`
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
