package common

import (
	"fmt"
)

//APIError api common
type APIError struct {
	Code    int
	Message string
	Data    interface{}
}
type APIErrorInfo interface {
	GetCode() string
	GetMessage() string
	GetData() interface{}
}

func newAPIError(code int, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

func NewAPIDataError(code int, message string, data interface{}) *APIError {
	return &APIError{Code: code, Message: message, Data: data}
}

func (r *APIError) Error() string {
	return fmt.Sprintf("%d", r.Code) + ":" + r.Message
}

func (r APIError) GetCode() int {
	return r.Code
}
func (r APIError) GetMessage() string {
	return r.Message
}
func (r APIError) GetData() interface{} {
	return r.Data
}

var (
	ErrUnknown                      = newAPIError(-1, "unknown error")
	Unauthorized                    = newAPIError(9999, "unauthorized")
	ErrSessionExpired               = newAPIError(10000, "session expired")
	ErrInvalidParameter             = newAPIError(10001, "invalid paramter")
)
