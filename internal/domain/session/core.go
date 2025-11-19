package session

import (
	"time"

	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

const _MaxAge = 10 * time.Minute

func Get(repo Getter, uuid uuid.UUID) (Entity, error) {
	res, err := repo.Get(uuid)
	if err != nil {
		return Entity{}, err
	}

	if time.Now().After(res.Expires) {
		return Entity{}, xerrors.ErrSessionNotFound
	}

	return res, nil
}

// TODO: verify validity of [_MaxAge] and turn it to an internal error
func CreateAndGet(repo Creater, user uuid.UUID) (Entity, error) {
	return CreateAndGetWithMaxAge(repo, user, _MaxAge)
}

func CreateAndGetWithMaxAge(repo Creater, user uuid.UUID, maxAge time.Duration) (Entity, error) {
	s, err := New(user, maxAge)
	if err != nil {
		return Entity{}, err
	}

	session := Entity{
		s.UUID(),
		s.User(),
		s.Expires(),
	}

	return session, repo.Create(session)
}

// TODO: verify validity of _MaxAge and turn it to an internal error
func Update(repo Updater, uuid uuid.UUID) error {
	return UpdateWithMaxAge(repo, uuid, _MaxAge)
}

func UpdateWithMaxAge(repo Updater, uuid uuid.UUID, maxAge time.Duration) error {
	var s Session
	s.SetMaxAge(maxAge)

	return repo.Update(uuid, s.Expires())
}
