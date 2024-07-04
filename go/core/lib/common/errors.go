package common

import (
	"errors"
	"strings"
)

var (
	// ErrNoResults error
	ErrNoResults = errors.New("no results")

	// ErrIAPUnauthorized error
	ErrIAPUnauthorized = errors.New("IAP unauthorized")

	// ErrUnauthorized error
	ErrUnauthorized = errors.New("unauthorized")

	// ErrConsistency error
	ErrConsistency = errors.New("consistency")

	// ErrLimit error
	ErrLimit = errors.New("limit")
)

// ErrAlreadyExists error
type ErrAlreadyExists struct {
	Err error
	Msg string
	Src string
}

func (e ErrAlreadyExists) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrAlreadyExists) Is(err error) bool {
	as := ErrAlreadyExists{}
	return errors.As(err, &as)
}

func (e ErrAlreadyExists) Unwrap() error {
	return e.Err
}

// ErrBadGateway error
type ErrBadGateway struct {
	Err error
	Msg string
	Src string
}

func (e ErrBadGateway) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrBadGateway) Is(err error) bool {
	as := ErrBadGateway{}
	return errors.As(err, &as)
}

func (e ErrBadGateway) Unwrap() error {
	return e.Err
}

// ErrBadRequest error
type ErrBadRequest struct {
	Err error
	Msg string
	Src string
}

func (e ErrBadRequest) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrBadRequest) Is(err error) bool {
	as := ErrBadRequest{}
	return errors.As(err, &as)
}

func (e ErrBadRequest) Unwrap() error {
	return e.Err
}

// ErrConnection error
type ErrConnection struct {
	Err error
	Msg string
	Src string
}

func (e ErrConnection) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrConnection) Is(err error) bool {
	as := ErrConnection{}
	return errors.As(err, &as)
}

func (e ErrConnection) Unwrap() error {
	return e.Err
}

// ErrConstraintViolation error
type ErrConstraintViolation struct {
	Err error
	Msg string
	Src string
}

func (e ErrConstraintViolation) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrConstraintViolation) Is(err error) bool {
	as := ErrConstraintViolation{}
	return errors.As(err, &as)
}

func (e ErrConstraintViolation) Unwrap() error {
	return e.Err
}

// ErrContextValues error
type ErrContextValues struct {
	Err error
	Msg string
	Src string
}

func (e ErrContextValues) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrContextValues) Is(err error) bool {
	as := ErrContextValues{}
	return errors.As(err, &as)
}

func (e ErrContextValues) Unwrap() error {
	return e.Err
}

// ErrForbidden error
type ErrForbidden struct {
	Err error
	Msg string
	Src string
}

func (e ErrForbidden) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrForbidden) Is(err error) bool {
	as := ErrForbidden{}
	return errors.As(err, &as)
}

func (e ErrForbidden) Unwrap() error {
	return e.Err
}

// ErrGatewayTimeout error
type ErrGatewayTimeout struct {
	Err error
	Msg string
	Src string
}

func (e ErrGatewayTimeout) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrGatewayTimeout) Is(err error) bool {
	as := ErrGatewayTimeout{}
	return errors.As(err, &as)
}

func (e ErrGatewayTimeout) Unwrap() error {
	return e.Err
}

// ErrGone error
type ErrGone struct {
	Err error
	Msg string
	Src string
}

func (e ErrGone) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrGone) Is(err error) bool {
	as := ErrGone{}
	return errors.As(err, &as)
}

func (e ErrGone) Unwrap() error {
	return e.Err
}

// ErrInternalServer error
type ErrInternalServer struct {
	Err error
	Msg string
	Src string
}

func (e ErrInternalServer) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrInternalServer) Is(err error) bool {
	as := ErrInternalServer{}
	return errors.As(err, &as)
}

func (e ErrInternalServer) Unwrap() error {
	return e.Err
}

// ErrUnprocessable error
type ErrUnprocessable struct {
	Err error
	Msg string
	Src string
}

func (e ErrUnprocessable) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrUnprocessable) Is(err error) bool {
	as := ErrUnprocessable{}
	return errors.As(err, &as)
}

func (e ErrUnprocessable) Unwrap() error {
	return e.Err
}

// ErrPaymentRequired error
type ErrPaymentRequired struct {
	Err error
	Msg string
	Src string
}

func (e ErrPaymentRequired) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")

}

// Is interface method
func (e ErrPaymentRequired) Is(err error) bool {
	as := ErrPaymentRequired{}
	return errors.As(err, &as)
}

func (e ErrPaymentRequired) Unwrap() error {
	return e.Err
}

// ErrNotFound error
type ErrNotFound struct {
	Err error
	Msg string
	Src string
}

func (e ErrNotFound) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrNotFound) Is(err error) bool {
	as := ErrNotFound{}
	return errors.As(err, &as)
}

func (e ErrNotFound) Unwrap() error {
	return e.Err
}

// ErrPreconditionFailed error
type ErrPreconditionFailed struct {
	Err error
	Msg string
	Src string
}

func (e ErrPreconditionFailed) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrPreconditionFailed) Is(err error) bool {
	as := ErrPreconditionFailed{}
	return errors.As(err, &as)
}

func (e ErrPreconditionFailed) Unwrap() error {
	return e.Err
}

// ErrTooManyRequests error
type ErrTooManyRequests struct {
	Err error
	Msg string
	Src string
}

func (e ErrTooManyRequests) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrTooManyRequests) Is(err error) bool {
	as := ErrTooManyRequests{}
	return errors.As(err, &as)
}

func (e ErrTooManyRequests) Unwrap() error {
	return e.Err
}

// ErrModeration error
type ErrModeration struct {
	Err error
	Msg string
	Src string
}

func (e ErrModeration) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	if e.Src == "" {
		return e.Msg
	}

	return strings.Join([]string{e.Src, e.Msg}, ": ")
}

// Is interface method
func (e ErrModeration) Is(err error) bool {
	as := ErrModeration{}
	return errors.As(err, &as)
}

func (e ErrModeration) Unwrap() error {
	return e.Err
}
