package memorystorage

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
)

type mapEvent map[int64]*model.Event

type Storage struct {
	data mapEvent
	mu   sync.RWMutex
}

var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("data is busy")
)

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
			return ErrDateBusy
		}
	}
	return nil
}

func (s *Storage) InsertEvent(ctx context.Context, e *model.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.ID = getNewIDSafe()
	s.data[e.ID] = e
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, e *model.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[e.ID]; !ok {
		return ErrEventNotFound
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

func (s *Storage) GetAllEvents(ctx context.Context, userID int64) ([]model.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sliceE := []model.Event{}
	for _, v := range s.data {
		if v.UserID == userID {
			sliceE = append(sliceE, *v)
		}
	}
	return sliceE, nil
}

func (s *Storage) GetAllRange(ctx context.Context, userID int64, begin, end time.Time) ([]model.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sliceE := []model.Event{}
	for _, v := range s.data {
		if v.UserID == userID &&
			(s.inTimeSpan(begin, end, v.OnTime) ||
				s.inTimeSpan(begin, end, v.OffTime)) {
			sliceE = append(sliceE, *v)
		}
	}
	return sliceE, nil
}

func (s *Storage) GetEventByID(ctx context.Context, eID int64) (model.Event, error) {
	var event model.Event
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.data[eID]; ok {
		event = *e
		return event, nil
	}
	return event, ErrEventNotFound
}

func (s *Storage) GetEventsDayOfNotice(ctx context.Context, date time.Time) ([]model.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sliceE := []model.Event{}

	for _, v := range s.data {
		if !v.Notified && (v.NotifyTime.Before(date) || v.NotifyTime.Equal(date)) {
			sliceE = append(sliceE, *v)
		}
	}
	return sliceE, nil
}

func (s *Storage) UpdateEventNotified(ctx context.Context, eventid int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[eventid]; !ok {
		return ErrEventNotFound
	}
	s.data[eventid].Notified = true
	return nil
}

func (s *Storage) DeleteEventsOlderDate(ctx context.Context, date time.Time) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	deleted := int64(0)
	for id, v := range s.data {
		if v.OffTime.Before(date) {
			delete(s.data, id)
			deleted++
		}
	}
	return deleted, nil
}
