package user

import (
	"time"

	"github.com/alan-b-lima/almodon/internal/auth"
	"github.com/alan-b-lima/almodon/pkg/opt"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

type Repository interface {
	Lister
	Getter
	GetterBySIAPE
	Creater
	Patcher
	Deleter
}

type (
	Lister interface {
		List(offset, limit int) (Entities, error)
	}

	Getter interface {
		Get(uuid.UUID) (Entity, error)
	}

	GetterBySIAPE interface {
		GetBySIAPE(int) (Entity, error)
	}

	Creater interface {
		Create(Entity) error
	}

	Patcher interface {
		Patch(uuid.UUID, PartialEntity) error
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
		UUID     uuid.UUID
		SIAPE    int
		Name     string
		Email    string
		Password [60]byte
		Role     auth.Role
	}

	PartialEntity struct {
		SIAPE    opt.Opt[int]
		Name     opt.Opt[string]
		Email    opt.Opt[string]
		Password opt.Opt[[60]byte]
		Role     opt.Opt[auth.Role]
	}

	AuthEntity struct {
		UUID    uuid.UUID
		User    uuid.UUID
		Expires time.Time
	}
)
