package userserve

import (
	"github.com/alan-b-lima/almodon/internal/auth"
	"github.com/alan-b-lima/almodon/internal/domain/user"
	"github.com/alan-b-lima/almodon/internal/support/service"
)

type AuthService struct {
	service   user.Service
	hierarchy auth.Hierarchy
}

func New(service user.Service) user.Service {
	return &AuthService{
		service:   service,
		hierarchy: auth.DefaultHierarchy,
	}
}

var permChief = auth.Permit(auth.Chief)

func (s *AuthService) List(act auth.Actor, req user.ListRequest) (user.ListResponse, error) {
	if err := service.Authorize(permChief, act); err != nil {
		return user.ListResponse{}, err
	}

	return s.service.List(act, req)
}

func (s *AuthService) Get(act auth.Actor, req user.GetRequest) (user.Response, error) {
	if act.User() == req.UUID {
		goto Do
	}

	if err := service.Authorize(permChief, act); err != nil {
		return user.Response{}, err
	}

Do:
	return s.service.Get(act, req)
}

func (s *AuthService) GetBySIAPE(act auth.Actor, req user.GetBySIAPERequest) (user.Response, error) {
	res, err := s.service.GetBySIAPE(act, req)
	if err != nil {
		return user.Response{}, err
	}

	if act.User() == res.UUID {
		goto Do
	}

	if err := service.Authorize(permChief, act); err != nil {
		return user.Response{}, err
	}

Do:
	return res, nil
}

func (s *AuthService) Create(act auth.Actor, req user.CreateRequest) (user.Response, error) {
	if err := service.Authorize(permChief, act); err != nil {
		return user.Response{}, err
	}

	return s.service.Create(act, req)
}

func (s *AuthService) Patch(act auth.Actor, req user.PatchRequest) (user.Response, error) {
	if act.User() == req.UUID {
		goto Do
	}

	if err := service.Authorize(permChief, act); err != nil {
		return user.Response{}, err
	}

Do:
	return s.service.Patch(act, req)
}

func (s *AuthService) UpdatePassword(act auth.Actor, req user.UpdatePasswordRequest) error {
	if act.User() == req.UUID {
		goto Do
	}

	if err := service.Authorize(permChief, act); err != nil {
		return err
	}

Do:
	return s.service.UpdatePassword(act, req)
}

func (s *AuthService) UpdateRole(act auth.Actor, req user.UpdateRoleRequest) (user.Response, error) {
	if err := service.Authorize(permChief, act); err != nil {
		return user.Response{}, err
	}

	return s.service.UpdateRole(act, req)
}

func (s *AuthService) Delete(act auth.Actor, req user.DeleteRequest) error {
	if act.User() == req.UUID {
		goto Do
	}

	if err := service.Authorize(permChief, act); err != nil {
		return err
	}

Do:
	return s.service.Delete(act, req)
}

func (s *AuthService) Authenticate(req user.AuthRequest) (user.AuthResponse, error) {
	return s.service.Authenticate(req)
}

func (s *AuthService) Actor(req user.ActorRequest) (auth.Actor, error) {
	return s.service.Actor(req)
}
