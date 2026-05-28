package core

// 400
type ErrResponseBadRequest struct {
	Code    string `json:"code" example:"INVALID_INPUT"`
	Message string `json:"message" example:"invalid request"`
}

// 404
type ErrResponseNotFound struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"resource not found"`
}

type ErrResponseItemNotFound struct {
	Code    string `json:"code" example:"ITEM_NOT_FOUND"`
	Message string `json:"message" example:"item not found"`
}

type ErrResponseMovementNotFound struct {
	Code    string `json:"code" example:"MOVEMENT_NOT_FOUND"`
	Message string `json:"message" example:"movement not found"`
}

// 409
type ErrResponseConflict struct {
	Code    string `json:"code" example:"DUPLICATE_TRANSACTION"`
	Message string `json:"message" example:"duplicate transaction detected"`
}

type ErrResponseItemConflict struct {
	Code    string `json:"code" example:"DUPLICATE_SKU"`
	Message string `json:"message" example:"duplicate SKU detected"`
}

// 500
type ErrResponseInternal struct {
	Code    string `json:"code" example:"INTERNAL_ERROR"`
	Message string `json:"message" example:"internal server error"`
}
