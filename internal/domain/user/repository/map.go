package userrepo

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"unsafe"

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

	datapath string
}

func NewMap() user.Repository {
	repo := Map{
		uuidIndex:  make(map[uuid.UUID]int),
		siapeIndex: make(map[int]int),
	}

	return &repo
}

func NewPersistantMap(datapath string) (user.Repository, error) {
	repo := Map{
		uuidIndex:  make(map[uuid.UUID]int),
		siapeIndex: make(map[int]int),
		datapath:   datapath,
	}

	if err := repo.init(); err != nil {
		return nil, err
	}

	return &repo, nil
}

func (m *Map) init() error {
	f, err := os.Open(m.datapath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var repo []entity
	if err := json.NewDecoder(f).Decode(&repo); err != nil {
		return err
	}

	m.repo = entity_from_json(repo)
	for i, record := range m.repo {
		m.uuidIndex[record.UUID] = i
		m.siapeIndex[record.SIAPE] = i
	}

	return nil
}

func (m *Map) Close() error {
	defer m.mu.Unlock()
	m.mu.Lock()

	if m.datapath == "" {
		return nil
	}

	f, err := os.OpenFile(m.datapath, os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := f.Truncate(0); err != nil {
		return err
	}

	repo := json_for_entity(m.repo)
	if err := json.NewEncoder(f).Encode(repo); err != nil {
		return err
	}

	return nil
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
	if ok && role != u.Role && u.Role == auth.Chief && !enough_chiefs(m) {
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
	if u.Role == auth.Chief && !enough_chiefs(m) {
		return xerrors.ErrNotEnoughChiefs
	}

	delete(m.uuidIndex, u.UUID)
	delete(m.siapeIndex, u.SIAPE)

	m.repo[index] = m.repo[len(m.repo)-1]
	m.repo = m.repo[:len(m.repo)-1]

	return nil
}

func enough_chiefs(m *Map) bool {
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

type entity struct {
	UUID     uuid.UUID `json:"uuid"`
	SIAPE    int       `json:"siape"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password pwd       `json:"password"`
	Role     role      `json:"role"`
}

type (
	pwd  [60]byte
	role auth.Role
)

func json_for_entity(entities []user.Entity) []entity {
	return unsafe.Slice((*entity)(unsafe.Pointer(unsafe.SliceData(entities))), len(entities))
}

func entity_from_json(entities []entity) []user.Entity {
	return unsafe.Slice((*user.Entity)(unsafe.Pointer(unsafe.SliceData(entities))), len(entities))
}

func (v pwd) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%+q", v[:]), nil
}

func (v *pwd) UnmarshalJSON(buf []byte) error {
	*v = pwd(buf[1 : len(buf)-1])
	return nil
}

func (v role) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%+q", auth.Role(v).String()), nil
}

func (v *role) UnmarshalJSON(buf []byte) error {
	r, _ := auth.FromString(string(buf[1 : len(buf)-1]))
	*v = role(r)
	return nil
}
