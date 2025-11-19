package promotion

import (
	"time"

	"github.com/alan-b-lima/almodon/pkg/uuid"
)

type Repository interface {
	Lister
	Getter
	GetterByUser
	Creater
	Updater
	Deleter
}

type (
	Lister interface {
		List(offset, limit int) (Entities, error)
	}

	Getter interface {
		Get(uuid.UUID) (Entity, error)
	}

	GetterByUser interface {
		GetByUser(uuid.UUID) (Entity, error)
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
	Entities struct {
		Offset       int
		Length       int
		Records      []Entity
		TotalRecords int
	}

	Entity struct {
		UUID    uuid.UUID
		User    uuid.UUID
		Expires time.Time
	}
)
