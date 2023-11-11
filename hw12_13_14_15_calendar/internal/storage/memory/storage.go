package memorystorage

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
)

type mapEvent map[int64]*storage.Event

type Storage struct {
	data mapEvent
	mu   sync.RWMutex
}

var GenID int64 = 0

func getNewIDSafe() int64 {
	return atomic.AddInt64(&GenID, 1)
}

func New() *Storage {
	return &Storage{data: make(mapEvent), mu: sync.RWMutex{}}
}

func (s *Storage) Connect(ctx context.Context) error {
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	return nil
}

func (s *Storage) InsertEvent(ctx context.Context, e *storage.Event) error {
	if err := app.CheckingEvent(e); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e.ID = getNewIDSafe()
	s.data[e.ID] = e
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, e *storage.Event) error {
	if err := app.CheckingEvent(e); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[e.ID]; !ok {
		return app.ErrEventNotFound
	}
	s.data[e.ID] = e
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, e *storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, e.ID)
	return nil
}

func (s *Storage) GetEventByID(ctx context.Context, eID int64) (storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.data[eID]; ok {
		return *e, nil
	}
	return storage.Event{}, nil
}

func (s *Storage) GetAll(ctx context.Context, userID int64) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sliceE := []storage.Event{}
	for _, v := range s.data {
		if v.UserID == userID {
			sliceE = append(sliceE, *v)
		}
	}
	return sliceE, nil
}
