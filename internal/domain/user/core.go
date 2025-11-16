package user

import (
	"time"

	"github.com/alan-b-lima/almodon/internal/auth"
	sessionpkg "github.com/alan-b-lima/almodon/internal/domain/session"
	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/errors"
	"github.com/alan-b-lima/almodon/pkg/hash"
	"github.com/alan-b-lima/almodon/pkg/opt"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

func List(users Lister, offset, limit int) (Entities, error) {
	return users.List(offset, limit)
}

func Get(users Getter, uuid uuid.UUID) (Entity, error) {
	return users.Get(uuid)
}

func GetBySIAPE(users GetterBySIAPE, siape int) (Entity, error) {
	return users.GetBySIAPE(siape)
}

func Create(users Creater, siape int, name, email, password string, role auth.Role) (uuid.UUID, error) {
	u, err := New(siape, name, email, password, role)
	if err != nil {
		return uuid.UUID{}, err
	}

	return u.UUID(), users.Create(translate(&u))
}

func Patch(users Patcher, uuid uuid.UUID, name, email, password opt.Opt[string], role opt.Opt[auth.Role]) error {
	var pu PartialEntity

	err := errors.Join(
		some_then(&pu.Name, name, ProcessName),
		some_then(&pu.Email, email, ProcessEmail),
		some_then(&pu.Password, password, ProcessPassword),
		some_then(&pu.Role, role, ProcessRole),
	)
	if err != nil {
		return err
	}

	return users.Patch(uuid, pu)
}

func Delete(users Deleter, uuid uuid.UUID) error {
	return users.Delete(uuid)
}

func Authenticate(users GetterBySIAPE, sessions sessionpkg.Creater, siape int, password string) (AuthEntity, error) {
	res, err := users.GetBySIAPE(siape)
	if err != nil {
		return AuthEntity{}, err
	}

	if !hash.Compare(res.Password[:], []byte(password)) {
		return AuthEntity{}, xerrors.ErrIncorrectPassword
	}

	s, err := sessions.Create(res.UUID, 10*time.Minute)
	if err != nil {
		return AuthEntity{}, err
	}

	ares := AuthEntity{
		UUID:    s.UUID,
		User:    res.UUID,
		Expires: s.Expires,
	}
	return ares, nil
}

func Actor(users Getter, sessions sessionpkg.Getter, session uuid.UUID) (auth.Actor, error) {
	res, err := sessionpkg.Get(sessions, session)
	if err != nil {
		return auth.NewUnlogged(), xerrors.ErrUnauthenticatedUser.New(err)
	}

	ures, err := users.Get(res.User)
	if err != nil {
		return auth.NewUnlogged(), xerrors.ErrUnauthenticatedUser.New(err)
	}

	return auth.NewLogged(
		ures.UUID,
		ures.Role,
	), nil
}

func translate(u *User) Entity {
	return Entity{
		UUID:     u.UUID(),
		SIAPE:    u.SIAPE(),
		Name:     u.Name(),
		Email:    u.Email(),
		Password: u.Password(),
		Role:     u.Role(),
	}
}

func some_then[F, R any](dst *opt.Opt[R], src opt.Opt[F], fn func(F) (R, error)) error {
	val, ok := src.Unwrap()
	if !ok {
		return nil
	}

	res, err := fn(val)
	if err != nil {
		return err
	}

	*dst = opt.Some(res)
	return nil
}
