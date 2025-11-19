package session

import (
	"time"

	"github.com/alan-b-lima/almodon/pkg/uuid"
)

type Repository interface {
	Getter
	Creater
	Updater
	Deleter
}

type (
	Getter interface {
		Get(uuid.UUID) (Entity, error)
	}

	Creater interface {
		Create(Entity) error
	}

	Updater interface {
		Update(uuid.UUID, time.Time) error
	}

	Deleter interface {
		Delete(uuid.UUID) error
	}
)

type (
	Entity struct {
		UUID    uuid.UUID
		User    uuid.UUID
		Expires time.Time
	}
)
