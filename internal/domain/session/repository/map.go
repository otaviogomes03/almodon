package sessionrepo

import (
	"sync"
	"time"

	"github.com/alan-b-lima/almodon/internal/domain/session"
	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/heap"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

type Map struct {
	uuidIndex   map[uuid.UUID]int
	userIndex   map[uuid.UUID]int
	expiresHeap sleepqueue

	repo []session.Entity
	mu   sync.RWMutex
}

func NewMap() session.Repository {
	repo := Map{
		uuidIndex: make(map[uuid.UUID]int),
		userIndex: make(map[uuid.UUID]int),
		expiresHeap: sleepqueue{
			new:    make(chan ess, 64),
			cancel: make(chan struct{}, 1),
		},
	}

	go flush(&repo)

	return &repo
}

func (m *Map) Get(uuid uuid.UUID) (session.Entity, error) {
	defer m.mu.RUnlock()
	m.mu.RLock()

	index, in := m.uuidIndex[uuid]
	if !in {
		return session.Entity{}, xerrors.ErrSessionNotFound
	}

	s := m.repo[index]
	if time.Now().After(s.Expires) {
		m.delete(s.UUID)
		return session.Entity{}, xerrors.ErrSessionNotFound
	}

	return m.repo[index], nil
}

func (m *Map) Create(session session.Entity) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	if index, in := m.userIndex[session.User]; in {
		s := m.repo[index]
		m.delete(s.UUID)
	}

	m.uuidIndex[session.UUID] = len(m.repo)
	m.userIndex[session.User] = len(m.repo)
	m.repo = append(m.repo, session)

	m.expiresHeap.new <- ess{session.UUID, session.Expires}

	return nil
}

func (m *Map) Update(uuid uuid.UUID, expires time.Time) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	index, in := m.uuidIndex[uuid]
	if !in {
		return xerrors.ErrSessionNotFound
	}

	s := &m.repo[index]
	s.Expires = expires

	m.expiresHeap.new <- ess{s.UUID, expires}

	return nil
}

func (m *Map) Delete(uuid uuid.UUID) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	return m.delete(uuid)
}

func (m *Map) delete(uuid uuid.UUID) error {
	index, in := m.uuidIndex[uuid]
	if !in {
		return nil
	}

	s := &m.repo[index]

	delete(m.uuidIndex, s.UUID)
	delete(m.userIndex, s.User)

	m.repo[index] = m.repo[len(m.repo)-1]
	m.repo = m.repo[:len(m.repo)-1]
	return nil
}

func flush(m *Map) {
	h := m.expiresHeap

	for {
		var after <-chan time.Time
		if h.heap.Len() > 0 {
			delay := time.Until(h.heap.Peek().expires)
			after = time.After(delay)
		}

		select {
		case <-h.cancel:
			return

		case es := <-h.new:
			h.heap.Push(es)

		case <-after:
			es := h.heap.Pop()
			m.Delete(es.session)
		}
	}
}

type sleepqueue struct {
	heap   heap.Heap[ess]
	new    chan ess
	cancel chan struct{}
}

type ess struct {
	session uuid.UUID
	expires time.Time
}

func (o0 ess) Less(o1 ess) bool { return o0.expires.Before(o1.expires) }
