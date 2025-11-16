package api

import (
	"errors"
	"net/http"

	sessionrepo "github.com/alan-b-lima/almodon/internal/domain/session/repository"
	userrepo "github.com/alan-b-lima/almodon/internal/domain/user/repository"
	users "github.com/alan-b-lima/almodon/internal/domain/user/resource"
	userserve "github.com/alan-b-lima/almodon/internal/domain/user/service"
)

type closer interface {
	Close() error
}

type Handler struct {
	mux     http.ServeMux
	cleanup []closer
}

func New() (*Handler, error) {
	var r Handler

	var (
		repoSessions = sessionrepo.NewMap()
		repoUsers    = userrepo.NewMap()
	)

	serveUsers := userserve.NewService(repoUsers, repoSessions)

	authServeUsers := userserve.New(serveUsers)

	users := users.New(serveUsers)

	resources := map[string]http.Handler{
		"users": users,
	}

	for name, handler := range resources {
		r.mux.Handle("/api/v1/"+name+"/", http.StripPrefix("/api/v1", handler))
	}

	r.attach(repoSessions)
	r.attach(repoUsers)
	r.attach(serveUsers)
	r.attach(authServeUsers)
	r.attach(users)

	return &r, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) Close() error {
	errs := make([]error, 0, len(h.cleanup))

	for _, closer := range h.cleanup {
		errs = append(errs, closer.Close())
	}

	return errors.Join(errs...)
}

func (h *Handler) attach(a any) bool {
	closer, ok := a.(closer)
	if !ok {
		return false
	}

	h.cleanup = append(h.cleanup, closer)
	return true
}
