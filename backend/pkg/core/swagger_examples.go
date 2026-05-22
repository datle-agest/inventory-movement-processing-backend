package core

// 400
type ErrResponseBadRequest struct {
    Code    int    `json:"code" example:"400"`
    Message string `json:"message" example:"invalid request"`
}

// 404
type ErrResponseNotFound struct {
    Code    int    `json:"code" example:"404"`
    Message string `json:"message" example:"resource not found"`
}

type ErrResponseItemNotFound struct {
    Code    int    `json:"code" example:"404"`
    Message string `json:"message" example:"item not found"`
}

type ErrResponseMovementNotFound struct {
    Code    int    `json:"code" example:"404"`
    Message string `json:"message" example:"movement not found"`
}

// 409
type ErrResponseConflict struct {
    Code    int    `json:"code" example:"409"`
    Message string `json:"message" example:"duplicate transaction detected"`
}

type ErrResponseItemConflict struct {
    Code    int    `json:"code" example:"409"`
    Message string `json:"message" example:"duplicate SKU detected"`
}

// 500
type ErrResponseInternal struct {
    Code    int    `json:"code" example:"500"`
    Message string `json:"message" example:"internal server error"`
}