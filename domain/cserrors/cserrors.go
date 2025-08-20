package cserrors

import (
	"net/http"
)

type Error struct {
	HttpStatus int
	Code       error
	Message    string
}

type Option struct {
	HttpStatus int
	Code       error
	Message    string
}

func New(opt *Option) *Error {
	if opt.HttpStatus == 0 {
		opt.HttpStatus = http.StatusOK
	}

	return &Error{
		HttpStatus: opt.HttpStatus,
		Code:       opt.Code,
		Message:    opt.Message,
	}
}

func (c *Error) Error() string {
	return c.Message
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
