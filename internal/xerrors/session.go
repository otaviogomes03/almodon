package xerrors

import "github.com/alan-b-lima/almodon/pkg/errors"

var (
	ErrSessionTooLong = errors.Fmt(errors.InvalidInput, "session-too-long", "session must not last longer than %v")

	ErrSessionNotFound = errors.New(errors.NotFound, "session-not-found", "session not found", nil)
)
