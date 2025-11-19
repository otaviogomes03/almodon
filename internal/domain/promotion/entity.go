package promotion

import (
	"time"

	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/errors"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

const _MaxMaxAge = 3 * 24 * time.Hour

type Promotion struct {
	uuid    uuid.UUID
	user    uuid.UUID
	expires time.Time
}

func New(user uuid.UUID, maxAge time.Duration) (Promotion, error) {
	session := Promotion{}

	err := errors.Join(
		session.setUser(user),
		session.SetMaxAge(maxAge),
	)
	if err != nil {
		return Promotion{}, err
	}

	session.uuid = uuid.NewUUIDv7()
	return session, nil
}

func (s *Promotion) UUID() uuid.UUID    { return s.uuid }
func (s *Promotion) User() uuid.UUID    { return s.user }
func (s *Promotion) Expires() time.Time { return s.expires }

func (s *Promotion) setUser(uuid uuid.UUID) error {
	s.user = uuid
	return nil
}

func (s *Promotion) SetMaxAge(maxAge time.Duration) error {
	val, err := ProcessMaxAge(maxAge)
	if err != nil {
		return err
	}

	s.expires = val
	return nil
}

func ProcessMaxAge(maxAge time.Duration) (time.Time, error) {
	if maxAge > _MaxMaxAge {
		return time.Time{}, xerrors.ErrPromotionTooLong.New(_MaxMaxAge)
	}

	return time.Now().Add(maxAge), nil
}
