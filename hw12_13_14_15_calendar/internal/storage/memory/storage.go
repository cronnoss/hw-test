package memorystorage

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

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

func (s *Storage) inTimeSpan(start, end, check time.Time) bool {
	switch {
	case check.Equal(start):
		return true
	case check.Equal(end):
		return true
	case check.After(start) && check.Before(end):
		return true
	}
	return false
}

func (s *Storage) IsBusyDateTimeRange(ctx context.Context, id, userID int64, onTime, offTime time.Time) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.data {
		if v.UserID == userID && v.ID != id &&
			(s.inTimeSpan(v.OnTime, v.OffTime, onTime) ||
				s.inTimeSpan(v.OnTime, v.OffTime, offTime)) {
			return app.ErrDateBusy
		}
	}
	return nil
}

func (s *Storage) InsertEvent(ctx context.Context, e *storage.Event) error {
	if err := app.CheckingEvent(e, false); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e.ID = getNewIDSafe()
	s.data[e.ID] = e
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, e *storage.Event) error {
	if err := app.CheckingEvent(e, true); err != nil {
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

func (s *Storage) DeleteEvent(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
	return nil
}

func (s *Storage) GetAllEvents(ctx context.Context, userID int64) ([]storage.Event, error) {
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

func (s *Storage) GetAllRange(ctx context.Context, userID int64, begin, end time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sliceE := []storage.Event{}
	for _, v := range s.data {
		if v.UserID == userID &&
			(s.inTimeSpan(begin, end, v.OnTime) ||
				s.inTimeSpan(begin, end, v.OffTime)) {
			sliceE = append(sliceE, *v)
		}
	}
	return sliceE, nil
}

func (s *Storage) GetEventByID(ctx context.Context, eID int64) (storage.Event, error) {
	var event storage.Event
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.data[eID]; ok {
		event = *e
		return event, nil
	}
	return event, app.ErrEventNotFound
}
