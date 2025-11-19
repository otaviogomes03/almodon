package api

import (
	"errors"
	"net/http"

	promotionrepo "github.com/alan-b-lima/almodon/internal/domain/promotion/repository"
	sessionrepo "github.com/alan-b-lima/almodon/internal/domain/session/repository"
	userrepo "github.com/alan-b-lima/almodon/internal/domain/user/repository"
	users "github.com/alan-b-lima/almodon/internal/domain/user/resource"
	userserve "github.com/alan-b-lima/almodon/internal/domain/user/service"
)

type Handler struct {
	http.ServeMux
	cleanup []closer
}

func New() (*Handler, error) {
	var r Handler

	var (
		repoPromotions = promotionrepo.NewMap()
		repoSessions   = sessionrepo.NewMap()
		repoUsers      = userrepo.NewMap()
	)

	serveUsers := userserve.NewService(repoUsers, repoSessions, repoPromotions)

	authServeUsers := userserve.New(serveUsers)

	users := users.New(authServeUsers)

	resources := map[string]http.Handler{
		"users": users,
	}

	for name, handler := range resources {
		r.Handle("/api/v1/"+name+"/", http.StripPrefix("/api/v1", handler))
	}

	r.attach(repoPromotions)
	r.attach(repoSessions)
	r.attach(repoUsers)
	r.attach(serveUsers)
	r.attach(authServeUsers)
	r.attach(users)

	return &r, nil
}

func (h *Handler) Close() error {
	errs := make([]error, 0, len(h.cleanup))

	for _, closer := range h.cleanup {
		errs = append(errs, closer.Close())
	}

	return errors.Join(errs...)
}

type closer interface{ Close() error }

func (h *Handler) attach(a any) bool {
	closer, ok := a.(closer)
	if !ok {
		return false
	}

	h.cleanup = append(h.cleanup, closer)
	return true
}
