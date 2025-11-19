package xerrors

import "github.com/alan-b-lima/almodon/pkg/errors"

var (
	ErrPromotionTooLong = errors.Fmt(errors.InvalidInput, "promotion-too-long", "promotion must not last longer than %v")

	ErrPromotionNotFound = errors.New(errors.NotFound, "promotion-not-found", "promotion not found", nil)
)
