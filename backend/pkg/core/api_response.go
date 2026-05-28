package core

import "inventory-movement-processing/common"

type APIResponse struct {
	Code       string      `json:"code"                example:"SUCCESS"`
	Message    string      `json:"message,omitempty"    example:"success"`
	Result     interface{} `json:"result,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Details    interface{} `json:"details,omitempty"`
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
	details ...interface{},
) *APIResponse {

	resp := &APIResponse{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 {
		resp.Details = details[0]
	}

	return resp
}
