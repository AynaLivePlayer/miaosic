package miaosic

import "errors"

var ErrNotImplemented = errors.New("miaosic: not implemented")

var (
	ErrorExternalApi          = errors.New("miaosic: external api error")
	ErrorNoSuchProvider       = errors.New("miaosic: no such provider")
	ErrorProviderNotLoginable = errors.New("miaosic: provider not loginable")
	ErrorDifferentProvider    = errors.New("miaosic: different provider")
	ErrorInvalidPageSetting   = errors.New("miaosic: invalid page setting")
	ErrorInvalidMediaMeta     = errors.New("miaosic: invalid media meta")
)
