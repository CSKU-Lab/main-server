package cserrors

import (
	"net/http"
	"strconv"
)

type ErrorCode int

func (c *ErrorCode) String() string {
	return strconv.Itoa(int(*c))
}

const (
	BAD_REQUEST           ErrorCode = http.StatusBadRequest
	UNAUTHORIZED          ErrorCode = http.StatusUnauthorized
	ALREADY_EXISTS        ErrorCode = http.StatusConflict
	INTERNAL_SERVER_ERROR ErrorCode = http.StatusInternalServerError
)

type Error struct {
	Code    ErrorCode
	Message string
}

func New(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (c *Error) Error() string {
	return c.Code.String() + "=" + c.Message
}

type RedirectCode string

const (
	REDIRECT_UNAUTHORIZED         RedirectCode = "UNAUTHORIZED"
	REDIRECT_SOMETHING_WENT_WRONG RedirectCode = "SOMETHING_WENT_WRONG"
)

type RedirectError interface {
	Code() string
	Error() string
}

type redirectError struct {
	code RedirectCode
}

func NewRedirect(code RedirectCode) RedirectError {
	return &redirectError{code: code}
}

func (r *redirectError) Code() string {
	return string(r.code)
}

func (r *redirectError) Error() string {
	return r.Code()
}
