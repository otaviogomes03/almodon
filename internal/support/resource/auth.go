package resource

import (
	"net/http"

	"github.com/alan-b-lima/almodon/internal/auth"
	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/errors"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

const SessionCookie = "session"

func Session(rc gatekeeper, r *http.Request) (auth.Actor, error) {
	session, err := session(r)
	if err != nil {
		return auth.NewUnlogged(), nil
	}

	act, err := rc.Actor(session)
	if err, ok := errors.AsType[*errors.Error](err); ok && err.Kind.IsClient() {
		return auth.NewUnlogged(), nil
	}
	if err != nil {
		return auth.NewUnlogged(), err
	}

	return act, err
}

type gatekeeper interface {
	Actor(session uuid.UUID) (auth.Actor, error)
}

func session(r *http.Request) (uuid.UUID, error) {
	s, err := r.Cookie(SessionCookie)
	if err != nil {
		return uuid.UUID{}, xerrors.ErrUnauthenticatedUser.New(nil)
	}

	session, err := uuid.FromString(s.Value)
	if err != nil {
		return uuid.UUID{}, xerrors.ErrBadUUID
	}

	return session, nil
}
