package promotion

import (
	"time"

	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

const _MaxAge = 1 * 24 * time.Hour

func Get(repo Getter, uuid uuid.UUID) (Entity, error) {
	res, err := repo.Get(uuid)
	if err != nil {
		return Entity{}, err
	}

	if time.Now().After(res.Expires) {
		return Entity{}, xerrors.ErrPromotionNotFound
	}

	return res, err
}

// TODO: verify validity of _MaxAge and turn it to an internal error
func Create(repo Creater, user uuid.UUID) (uuid.UUID, error) {
	return CreateWithMaxAge(repo, user, _MaxAge)
}

func CreateWithMaxAge(repo Creater, user uuid.UUID, maxAge time.Duration) (uuid.UUID, error) {
	p, err := New(user, _MaxAge)
	if err != nil {
		return uuid.UUID{}, err
	}

	return p.UUID(), repo.Create(translate(&p))
}

// TODO: verify validity of _MaxAge and turn it to an internal error
func Update(repo Updater, uuid uuid.UUID) error {
	return UpdateWithMaxAge(repo, uuid, _MaxAge)
}

func UpdateWithMaxAge(repo Updater, uuid uuid.UUID, maxAge time.Duration) error {
	var p Promotion
	if err := p.SetMaxAge(maxAge); err != nil {
		return err
	}

	return repo.Update(uuid, p.Expires())
}

func Delete(repo Deleter, uuid uuid.UUID) error {
	return repo.Delete(uuid)
}

func translate(p *Promotion) Entity {
	return Entity{
		UUID:    p.UUID(),
		User:    p.User(),
		Expires: p.Expires(),
	}
}
