package core

import "inventory-movement-processing/common"

type APIResponse struct {
	Code       string      `json:"code"                example:"SUCCESS"`
	Message    string      `json:"message,omitempty"    example:"success"`
	Result     interface{} `json:"result,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func Success(data interface{}) *APIResponse {
	return &APIResponse{
		Code:   common.CodeSuccess,
		Result: data,
	}
}

func SuccessWithMessage(
	data interface{},
	message string,
) *APIResponse {

	return &APIResponse{
		Code:    common.CodeSuccess,
		Message: message,
		Result:  data,
	}
}

func SuccessWithPaging(
	data interface{},
	paging *Pagination,
	message string,
) *APIResponse {

	return &APIResponse{
		Code:       common.CodeSuccess,
		Message:    message,
		Result:     data,
		Pagination: paging,
	}
}

func Fail(
	code string,
	message string,
) *APIResponse {

	return &APIResponse{
		Code:    code,
		Message: message,
	}
}
