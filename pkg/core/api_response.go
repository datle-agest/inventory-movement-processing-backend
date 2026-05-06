package core

type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type SuccessResponseWithPaging struct {
	Data       interface{} `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}

func NewSuccess(data interface{}) *SuccessResponse {
	return &SuccessResponse{Data: data}
}

func NewSuccessWithMessage(data interface{}, message string) *SuccessResponse {
	return &SuccessResponse{Data: data, Message: message}
}

func NewSuccessWithPaging(data interface{}, paging *Pagination) *SuccessResponseWithPaging {
	return &SuccessResponseWithPaging{Data: data, Pagination: paging}
}

func NewError(code int, err error) *ErrorResponse {
	return &ErrorResponse{Code: code, Message: err.Error()}
}

func NewErrorWithType(code int, errType string, err error) *ErrorResponse {
	return &ErrorResponse{Code: code, Type: errType, Message: err.Error()}
}

/*
HOW TO USE RESPONSE

Success: c.JSON(200, common.NewSuccess(data))
Success + message: c.JSON(200, common.NewSuccessWithMessage(data, "created"))
Success + paging: c.JSON(200, common.NewSuccessWithPaging(data, paging))


Error: c.JSON(400, common.NewError(400, err))
Error + type: c.JSON(404, common.NewErrorWithType(404, "NOT_FOUND", err))

NOTE:
- code phải giống HTTP status (400, 404, 500...)
- không trả lỗi nhạy cảm (DB, internal...)
- dùng type cho FE handle (VALIDATION_ERROR, NOT_FOUND,...)
*/
