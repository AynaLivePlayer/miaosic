package miaosic

import "errors"

var ErrNotImplemented = errors.New("not implemented")

var (
	ErrorExternalApi    = errors.New("external api error")
	ErrorNoSuchProvider = errors.New("not such provider")
)
