package user

import (
	"github.com/alan-b-lima/almodon/internal/auth"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

type Service interface {
	List(act auth.Actor, req ListRequest) (ListResponse, error)

	Get(act auth.Actor, req GetRequest) (Response, error)
	GetBySIAPE(act auth.Actor, req GetBySIAPERequest) (Response, error)

	Create(act auth.Actor, req CreateRequest) (uuid.UUID, error)

	Patch(act auth.Actor, req PatchRequest) error
	UpdatePassword(act auth.Actor, req UpdatePasswordRequest) error
	UpdateRole(act auth.Actor, req UpdateRoleRequest) error

	Delete(act auth.Actor, req DeleteRequest) error

	Gatekeeper
}

type Gatekeeper interface {
	Authenticate(req AuthRequest) (AuthResponse, error)
	Actor(req ActorRequest) (auth.Actor, error)
}
