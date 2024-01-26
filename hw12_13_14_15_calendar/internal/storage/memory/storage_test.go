package memorystorage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Error("NewCalendar() should not return nil")
	}
}

func TestConnect(t *testing.T) {
	s := New()
	err := s.Connect(context.Background())
	if err != nil {
		t.Errorf("Connect() error = %v, wantErr nil", err)
	}
}

func TestClose(t *testing.T) {
	s := New()
	err := s.Close(context.Background())
	if err != nil {
		t.Errorf("Close() error = %v, wantErr nil", err)
	}
}

func TestInsertEvent(t *testing.T) {
	s := New()
	e := &model.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	err := s.InsertEvent(context.Background(), e)
	if err != nil {
		t.Errorf("InsertEvent() error = %v, wantErr nil", err)
	}
	require.Equal(t, nil, err)
	require.NotEqual(t, int64(0), e.ID)
	require.NoError(t, err)
}

func TestUpdateEvent(t *testing.T) {
	s := New()
	e := &model.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	_ = s.InsertEvent(context.Background(), e)
	err := s.UpdateEvent(context.Background(), e)
	if err != nil {
		t.Errorf("UpdateEvent() error = %v, wantErr nil", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	s := New()
	e := &model.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	_ = s.InsertEvent(context.Background(), e)
	err := s.DeleteEvent(context.Background(), e.ID)
	if err != nil {
		t.Errorf("DeleteEvent() error = %v, wantErr nil", err)
	}
	e2, err := s.GetEventByID(context.Background(), 1)
	require.ErrorIs(t, err, ErrEventNotFound)
	require.Equal(t, int64(0), e2.ID)
}

func TestGetEventById(t *testing.T) {
	s := New()
	e := &model.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	_ = s.InsertEvent(context.Background(), e)
	_, err := s.GetEventByID(context.Background(), e.ID)
	if err != nil {
		t.Errorf("GetEventByID() error = %v, wantErr nil", err)
	}
}

func TestGetAll(t *testing.T) {
	s := New()
	e := &model.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	e2 := &model.Event{
		ID:          2,
		UserID:      1,
		Title:       "test2",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test2",
	}
	_ = s.InsertEvent(context.Background(), e)
	_ = s.InsertEvent(context.Background(), e2)
	events, err := s.GetAllEvents(context.Background(), 1)
	if err != nil {
		t.Errorf("GetAllEvents() error = %v, wantErr nil", err)
	}
	require.Equal(t, 2, len(events), "GetAllEvents() did not return all the events")
	for _, event := range events {
		require.True(t, event.ID == e.ID || event.ID == e2.ID, "GetAllEvents() returned an unknown event")
	}
}

func TestWrongUpdateEvent(t *testing.T) {
	s := New()
	e := &model.Event{
		ID:          -1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	err := s.UpdateEvent(context.Background(), e)
	require.ErrorIs(t, err, ErrEventNotFound)
}

func TestUpdateEventNotFound(t *testing.T) {
	s := New()
	e := &model.Event{
		ID:          100,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	err := s.UpdateEvent(context.Background(), e)
	require.ErrorIs(t, err, ErrEventNotFound)
}

func TestUpdateEventNotValid(t *testing.T) {
	s := New()
	invalidEvent := &model.Event{}
	err := s.UpdateEvent(context.Background(), invalidEvent)
	require.Error(t, err)
}

func TestInsertEventThreadSafe(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	eventsNum := 10000
	for i := 0; i < eventsNum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			e := &model.Event{
				UserID:      int64(i + 1),
				Title:       fmt.Sprintf("Title_N%v", i+1),
				Description: fmt.Sprintf("Description_N%v", i+1),
				OnTime:      time.Now(),
				OffTime:     time.Now().AddDate(0, 0, 5),
				NotifyTime:  time.Now().AddDate(0, 0, 1),
			}
			err := s.InsertEvent(context.Background(), e)
			assert.Nil(t, err)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, len(s.data), eventsNum)
}

func TestConcurrency_Insert(t *testing.T) {
	s := New()
	event := &model.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ev := *event // Copy original event
			ev.ID = int64(i)
			err := s.InsertEvent(context.Background(), &ev)
			require.NoError(t, err)
		}(i)
	}
	wg.Wait()

	events, err := s.GetAllEvents(context.Background(), event.UserID)
	require.NoError(t, err)
	require.Equal(t, 100, len(events), "Inserts are not concurrently safe")
}

func TestConcurrency_Update(t *testing.T) {
	s := New()
	event := &model.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}

	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
		_ = s.InsertEvent(context.Background(), event)

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ev := *event // Copy original event
			ev.Title = fmt.Sprintf("title_%v", i)
			err := s.UpdateEvent(context.Background(), &ev)
			require.NoError(t, err)
		}(i)
	}
	wg.Wait()

	events, err := s.GetAllEvents(context.Background(), event.UserID)
	require.NoError(t, err)
	for _, ev := range events {
		require.Contains(t, ev.Title, "title_", "Updates are not concurrently safe")
	}
}
