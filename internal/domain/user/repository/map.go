package userrepo

import (
	"cmp"
	"sync"

	"github.com/alan-b-lima/almodon/internal/auth"
	"github.com/alan-b-lima/almodon/internal/domain/user"
	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/opt"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

type Map struct {
	uuidIndex  map[uuid.UUID]int
	siapeIndex map[int]int

	repo []user.Entity
	mu   sync.RWMutex
}

func NewMap() user.Repository {
	repo := Map{
		uuidIndex:  make(map[uuid.UUID]int),
		siapeIndex: make(map[int]int),
	}

	return &repo
}

func (m *Map) List(offset, limit int) (user.Entities, error) {
	defer m.mu.RUnlock()
	m.mu.RLock()

	lo := clamp(0, offset, len(m.repo))
	hi := clamp(0, offset+limit, len(m.repo))

	if lo >= hi {
		return user.Entities{
			Records:      []user.Entity{},
			TotalRecords: len(m.repo),
		}, nil
	}

	res := make([]user.Entity, hi-lo)
	copy(res, m.repo[lo:hi])

	return user.Entities{
		Offset:       lo,
		Length:       len(res),
		Records:      res,
		TotalRecords: len(m.repo),
	}, nil
}

func (m *Map) Get(uuid uuid.UUID) (user.Entity, error) {
	defer m.mu.RUnlock()
	m.mu.RLock()

	index, in := m.uuidIndex[uuid]
	if !in {
		return user.Entity{}, xerrors.ErrUserNotFound
	}

	return m.repo[index], nil
}

func (m *Map) GetBySIAPE(siape int) (user.Entity, error) {
	defer m.mu.RUnlock()
	m.mu.RLock()

	index, in := m.siapeIndex[siape]
	if !in {
		return user.Entity{}, xerrors.ErrUserNotFound
	}

	return m.repo[index], nil
}

func (m *Map) Create(user user.Entity) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	if _, in := m.siapeIndex[user.SIAPE]; in {
		return xerrors.ErrSiapeTaken
	}

	m.uuidIndex[user.UUID] = len(m.repo)
	m.siapeIndex[user.SIAPE] = len(m.repo)
	m.repo = append(m.repo, user)

	return nil
}

func (m *Map) Patch(uuid uuid.UUID, user user.PartialEntity) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	index, in := m.uuidIndex[uuid]
	if !in {
		return xerrors.ErrUserNotFound
	}

	u := &m.repo[index]

	role, ok := user.Role.Unwrap()
	if ok && role != u.Role && u.Role == auth.Chief && !m.enough_chiefs() {
		return xerrors.ErrNotEnoughChiefs
	} else {
		u.Role = role
	}

	some_then(&u.Name, user.Name)
	some_then(&u.Email, user.Email)
	some_then(&u.Password, user.Password)

	return nil
}

func (m *Map) Delete(uuid uuid.UUID) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	index, in := m.uuidIndex[uuid]
	if !in {
		return nil
	}

	u := &m.repo[index]
	if u.Role == auth.Chief && !m.enough_chiefs() {
		return xerrors.ErrNotEnoughChiefs
	}

	delete(m.uuidIndex, u.UUID)
	delete(m.siapeIndex, u.SIAPE)

	m.repo[index] = m.repo[len(m.repo)-1]
	m.repo = m.repo[:len(m.repo)-1]

	return nil
}

func (m *Map) enough_chiefs() bool {
	var count int
	for _, user := range m.repo {
		if user.Role == auth.Chief {
			count++
		}
	}

	if count < 2 {
		return false
	}

	return true
}

func some_then[F any](dst *F, src opt.Opt[F]) {
	val, ok := src.Unwrap()
	if !ok {
		return
	}

	*dst = val
}

func clamp[T cmp.Ordered](mn, val, mx T) T {
	return min(max(mn, val), mx)
}
