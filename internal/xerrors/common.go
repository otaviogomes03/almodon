package xerrors

import "github.com/alan-b-lima/almodon/pkg/errors"

var ErrNameEmpty = errors.New(errors.InvalidInput, "name-empty", "name cannot be empty", nil)

var (
	ErrResourceNotFound = errors.Fmt(errors.NotFound, "resource-not-found", "resource %q not found")

	ErrBadUUID        = errors.New(errors.InvalidInput, "bad-uuid", "given UUID could not be parsed", nil)
	ErrBadRole        = errors.New(errors.InvalidInput, "bad-role", "given role could not be parsed", nil)
	ErrBadQueryParams = errors.Imp(errors.InvalidInput, "bad-query", "bad query parameters")

	ErrNoContentType              = errors.New(errors.PreconditionFailed, "no-content-type", "content type must be informed", nil)
	ErrUnsupportedContentTypeJson = errors.New(errors.PreconditionFailed, "unsupported-content-type", "content type must be application/json", nil)
	ErrJsonSyntax                 = errors.Fmt(errors.InvalidInput, "json-syntax-error", "JSON syntax error at %d")
	ErrJsonType                   = errors.Fmt(errors.InvalidInput, "json-type-error", "JSON type error at %d, expected %v but got %v")
	ErrNotAcceptableJson          = errors.New(errors.PreconditionFailed, "not-acceptable-type", "client does not accept application/json", nil)
)

var ErrTODO = errors.New(errors.Internal, "todo", "implement me", nil)
