package promotionrepo

import (
	"cmp"
	"sync"
	"time"

	"github.com/alan-b-lima/almodon/internal/domain/promotion"
	"github.com/alan-b-lima/almodon/internal/xerrors"
	"github.com/alan-b-lima/almodon/pkg/heap"
	"github.com/alan-b-lima/almodon/pkg/uuid"
)

type Map struct {
	uuidIndex   map[uuid.UUID]int
	userIndex   map[uuid.UUID]int
	expiresHeap sleepqueue

	repo []promotion.Entity
	mu   sync.RWMutex
}

func NewMap() promotion.Repository {
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

func (m *Map) List(offset int, limit int) (promotion.Entities, error) {
	defer m.mu.RUnlock()
	m.mu.RLock()

	lo := clamp(0, offset, len(m.repo))
	hi := clamp(0, offset+limit, len(m.repo))

	if lo >= hi {
		return promotion.Entities{
			Records:      []promotion.Entity{},
			TotalRecords: len(m.repo),
		}, nil
	}

	res := make([]promotion.Entity, hi-lo)
	copy(res, m.repo[lo:hi])

	return promotion.Entities{
		Offset:       lo,
		Length:       len(res),
		Records:      res,
		TotalRecords: len(m.repo),
	}, nil
}

func (m *Map) Get(uuid uuid.UUID) (promotion.Entity, error) {
	defer m.mu.RUnlock()
	m.mu.RLock()

	index, in := m.uuidIndex[uuid]
	if !in {
		return promotion.Entity{}, xerrors.ErrPromotionNotFound
	}

	s := m.repo[index]
	if time.Now().After(s.Expires) {
		m.delete(s.UUID)
		return promotion.Entity{}, xerrors.ErrPromotionNotFound
	}

	return s, nil
}

func (m *Map) GetByUser(user uuid.UUID) (promotion.Entity, error) {
	defer m.mu.RUnlock()
	m.mu.RLock()

	index, in := m.userIndex[user]
	if !in {
		return promotion.Entity{}, xerrors.ErrPromotionNotFound
	}

	s := m.repo[index]
	if time.Now().After(s.Expires) {
		m.delete(s.UUID)
		return promotion.Entity{}, xerrors.ErrPromotionNotFound
	}

	return s, nil
}

func (m *Map) Create(promotion promotion.Entity) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	if index, in := m.userIndex[promotion.User]; in {
		s := m.repo[index]
		m.delete(s.UUID)
	}

	m.uuidIndex[promotion.UUID] = len(m.repo)
	m.userIndex[promotion.User] = len(m.repo)
	m.repo = append(m.repo, promotion)

	m.expiresHeap.new <- ess{
		promotion: promotion.UUID,
		expires:   promotion.Expires,
	}

	return nil
}

func (m *Map) Update(uuid uuid.UUID, expires time.Time) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	index, in := m.uuidIndex[uuid]
	if !in {
		return xerrors.ErrPromotionNotFound
	}

	s := &m.repo[index]
	s.Expires = expires

	m.expiresHeap.new <- ess{
		promotion: s.UUID,
		expires:   s.Expires,
	}

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
			h.cancel <- struct{}{}
			return

		case es := <-h.new:
			h.heap.Push(es)

		case <-after:
			es := h.heap.Pop()
			m.Delete(es.promotion)
		}
	}
}

func clamp[T cmp.Ordered](mn, val, mx T) T {
	return min(max(mn, val), mx)
}

type sleepqueue struct {
	heap   heap.Heap[ess]
	new    chan ess
	cancel chan struct{}
}

type ess struct {
	promotion uuid.UUID
	expires   time.Time
}

func (o0 ess) Less(o1 ess) bool { return o0.expires.Before(o1.expires) }
