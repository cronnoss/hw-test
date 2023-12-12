package memorystorage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func helperEvent(ev *storage.Event, i int) {
	ev.UserID = int64(i + 1)
	ev.Title = fmt.Sprintf("Title_N%v", i+1)
	ev.Description = fmt.Sprintf("Description_N%v", i+1)
	ev.OnTime = time.Now()
	ev.OffTime = time.Now().AddDate(0, 0, 7)
	ev.NotifyTime = time.Now().AddDate(0, 0, 6)
}

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Error("New() should not return nil")
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
	e := &storage.Event{
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
	e := &storage.Event{
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
	e := &storage.Event{
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
	require.ErrorIs(t, err, app.ErrEventNotFound)
	require.Equal(t, int64(0), e2.ID)
}

func TestGetEventById(t *testing.T) {
	s := New()
	e := &storage.Event{
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
	e := &storage.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	e2 := &storage.Event{
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
	e := &storage.Event{
		ID:          -1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	err := s.UpdateEvent(context.Background(), e)
	require.ErrorIs(t, err, app.ErrEventNotFound)
}

func TestUpdateEventNotFound(t *testing.T) {
	s := New()
	e := &storage.Event{
		ID:          100,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	err := s.UpdateEvent(context.Background(), e)
	require.ErrorIs(t, err, app.ErrEventNotFound)
}

func TestUpdateEventNotValid(t *testing.T) {
	s := New()
	invalidEvent := &storage.Event{}
	err := s.UpdateEvent(context.Background(), invalidEvent)
	require.Error(t, err)
}

func TestCheckingEvent(t *testing.T) {
	e := &storage.Event{
		ID:          1,
		UserID:      1,
		Title:       "test",
		OnTime:      time.Now(),
		OffTime:     time.Now().AddDate(0, 0, 5),
		NotifyTime:  time.Now().AddDate(0, 0, 1),
		Description: "test",
	}
	err := app.CheckingEvent(e, false)
	require.Equal(t, nil, err)
}

func TestStorageRules(t *testing.T) {
	s := New()

	t.Parallel()
	t.Run("Checking userID", func(t *testing.T) {
		var e storage.Event
		helperEvent(&e, 1)
		e.UserID = 0
		err := s.InsertEvent(context.Background(), &e)
		require.ErrorIs(t, err, app.ErrUserID, "expected err message")
	})

	t.Run("Checking title", func(t *testing.T) {
		var e storage.Event
		helperEvent(&e, 2)
		e.Title = string(make([]byte, 151))
		err := s.InsertEvent(context.Background(), &e)
		require.ErrorIs(t, err, app.ErrTitle, "expected err message")
	})

	t.Run("Checking onTime", func(t *testing.T) {
		var e storage.Event
		helperEvent(&e, 3)
		e.OnTime = time.Time{}
		err := s.InsertEvent(context.Background(), &e)
		require.ErrorIs(t, err, app.ErrOnTime, "expected err message")
	})

	t.Run("Checking offTime", func(t *testing.T) {
		var e storage.Event
		helperEvent(&e, 4)
		e.OnTime = time.Now()
		e.OffTime = time.Now().AddDate(0, 0, -1)
		err := s.InsertEvent(context.Background(), &e)
		require.ErrorIs(t, err, app.ErrOffTime, "expected err message")
	})

	t.Run("Checking notifyTime", func(t *testing.T) {
		var e storage.Event
		helperEvent(&e, 5)
		e.OnTime = time.Now()
		e.OffTime = time.Now().AddDate(0, 0, 3)
		e.NotifyTime = time.Now().AddDate(0, 0, 4)
		err := s.InsertEvent(context.Background(), &e)
		require.ErrorIs(t, err, app.ErrNotifyTime, "expected err message")

		e.OnTime = time.Now()
		e.OffTime = time.Now().AddDate(0, 0, 3)
		e.NotifyTime = time.Now().AddDate(0, 0, -1)
		err = s.InsertEvent(context.Background(), &e)
		require.ErrorIs(t, err, app.ErrNotifyTime, "expected err message")
	})
}

func TestInsertEventThreadSafe(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	eventsNum := 10000
	for i := 0; i < eventsNum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			e := &storage.Event{
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
	event := &storage.Event{
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
	event := &storage.Event{
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
