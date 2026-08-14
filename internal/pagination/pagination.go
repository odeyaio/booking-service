package pagination

import (
	"errors"
	"math"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

var ErrInvalid = errors.New("invalid pagination")

type Params struct {
	Page     int
	PageSize int
}

func New(page, pageSize *int) Params {
	params := Params{
		Page:     DefaultPage,
		PageSize: DefaultPageSize,
	}
	if page != nil {
		params.Page = *page
	}
	if pageSize != nil {
		params.PageSize = *pageSize
	}
	return params
}

func (p Params) Validate() error {
	if p.Page < 1 || p.PageSize < 1 || p.PageSize > MaxPageSize {
		return ErrInvalid
	}

	if p.Page-1 > math.MaxInt/p.PageSize {
		return ErrInvalid
	}

	return nil
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

type Metadata struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

func NewMetadata(params Params, total int) Metadata {
	return Metadata{
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}
}
