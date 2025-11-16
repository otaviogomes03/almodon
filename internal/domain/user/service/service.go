package userserve

import (
	"github.com/alan-b-lima/almodon/internal/auth"
	"github.com/alan-b-lima/almodon/internal/domain/session"
	"github.com/alan-b-lima/almodon/internal/domain/user"
	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/opt"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

type Service struct {
	users    user.Repository
	sessions session.Repository
}

func NewService(users user.Repository, sessions session.Repository) user.Service {
	return &Service{
		users:    users,
		sessions: sessions,
	}
}

func (s *Service) List(act auth.Actor, req user.ListRequest) (user.ListResponse, error) {
	res, err := user.List(s.users, req.Offset, req.Limit)
	if err != nil {
		return user.ListResponse{}, err
	}

	lres := user.ListResponse{
		Offset:       res.Offset,
		Length:       res.Length,
		Records:      make([]user.Response, res.Length),
		TotalRecords: res.TotalRecords,
	}
	for i := range res.Records {
		transformP(&lres.Records[i], &res.Records[i])
	}

	return lres, nil
}

func (s *Service) Get(act auth.Actor, req user.GetRequest) (user.Response, error) {
	res, err := user.Get(s.users, req.UUID)
	if err != nil {
		return user.Response{}, err
	}

	return transform(&res), nil
}

func (s *Service) GetBySIAPE(act auth.Actor, req user.GetBySIAPERequest) (user.Response, error) {
	res, err := user.GetBySIAPE(s.users, req.SIAPE)
	if err != nil {
		return user.Response{}, err
	}

	return transform(&res), err
}

func (s *Service) Create(act auth.Actor, req user.CreateRequest) (uuid.UUID, error) {
	role, ok := auth.FromString(req.Role)
	if !ok {
		return uuid.UUID{}, xerrors.ErrBadRole
	}

	rcs, err := user.Create(s.users, req.SIAPE, req.Name, req.Email, req.Password, role)
	if err != nil {
		return uuid.UUID{}, err
	}

	return rcs, nil
}

func (s *Service) Patch(act auth.Actor, req user.PatchRequest) error {
	var string opt.Opt[string]
	var role opt.Opt[auth.Role]

	return user.Patch(s.users, req.UUID, req.Name, req.Email, string, role)
}

func (s *Service) UpdatePassword(act auth.Actor, req user.UpdatePasswordRequest) error {
	var string opt.Opt[string]
	var role opt.Opt[auth.Role]

	return user.Patch(s.users, req.UUID, string, string, opt.Some(req.Password), role)
}

func (s *Service) UpdateRole(act auth.Actor, req user.UpdateRoleRequest) error {
	var string opt.Opt[string]

	return user.Patch(s.users, req.UUID, string, string, string, opt.Some(req.Role))
}

func (s *Service) Delete(act auth.Actor, req user.DeleteRequest) error {
	return user.Delete(s.users, req.UUID)
}

func (s *Service) Authenticate(req user.AuthRequest) (user.AuthResponse, error) {
	res, err := user.Authenticate(s.users, s.sessions, req.SIAPE, req.Password)
	if err != nil {
		return user.AuthResponse{}, err
	}

	return user.AuthResponse(res), nil
}

func (s *Service) Actor(req user.ActorRequest) (auth.Actor, error) {
	return user.Actor(s.users, s.sessions, req.Session)
}

func transform(e *user.Entity) user.Response {
	return user.Response{
		UUID:  e.UUID,
		SIAPE: e.SIAPE,
		Name:  e.Name,
		Email: e.Email,
		Role:  e.Role.String(),
	}
}

func transformP(r *user.Response, e *user.Entity) {
	r.UUID = e.UUID
	r.SIAPE = e.SIAPE
	r.Name = e.Name
	r.Email = e.Email
	r.Role = e.Role.String()
}
