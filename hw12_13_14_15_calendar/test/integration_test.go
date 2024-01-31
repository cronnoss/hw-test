//go:build integration
// +build integration

package test_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/pkg/event_service_v1"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	querySelectByID = `
		SELECT id, userid, title, description, ontime, offtime, notifytime
		FROM events WHERE id = $1
	`
	queryInsert = `
		INSERT INTO events (userid, title, description, ontime, offtime, notifytime)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`
)

type CalendarSuite struct {
	suite.Suite
	ctx          context.Context
	calendarConn *grpc.ClientConn
	client       event_service_v1.EventServiceV1Client
	db           *sqlx.DB
}

func (s *CalendarSuite) SetupSuite() { // general setting for the entire suite
	s.ctx = context.Background()

	host := os.Getenv("GRPC_HOST")
	port := os.Getenv("GRPC_PORT")
	calendarHost := host + ":" + port
	// calendarHost := ""

	if calendarHost == ":" {
		calendarHost = "localhost:50000"
	}
	var err error
	s.calendarConn, err = grpc.Dial(calendarHost, grpc.WithInsecure()) //nolint:staticcheck
	s.Require().NoError(err)
	s.client = event_service_v1.NewEventServiceV1Client(s.calendarConn)

	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		// "postgres", "postgres", "0.0.0.0", 5432, "calendar")
		"postgres", "postgres", "0.0.0.0", 5432, "calendar")
	s.db, err = sqlx.Open("pgx", connectionString)
	s.Require().NoError(err)
}

func (s *CalendarSuite) SetupTest() { // setting for a specific test
}

func (s *CalendarSuite) TearDownTest() { // cleaning for a specific test
	query := `TRUNCATE TABLE events`
	_, err := s.db.Exec(query)
	s.Require().NoError(err)
}

func (s *CalendarSuite) TearDownSuite() { // cleaning for the entire suite
	s.calendarConn.Close()
	defer s.db.Close()
}

func TestCalendarPost(t *testing.T) {
	suite.Run(t, new(CalendarSuite))
}

func (s *CalendarSuite) TestCalendar_GetEventByID() {
	// insert event to DB
	now := time.Now().Format("2006-01-02T15:04:05Z")
	row := s.db.QueryRow(
		queryInsert,
		1,
		"title 1",
		"description 1",
		now,
		now,
		now,
	)

	// get event ID from DB
	var ID int64
	_ = row.Scan(&ID)

	// get event by ID through gRPC
	request := &event_service_v1.ReqByID{
		ID: &ID,
	}
	response, err := s.client.GetEventByID(s.ctx, request)

	// compare event from DB with event from gRPC request
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil GetEventByID")
	s.Require().NotNil(response.GetEvent(), "expected a non-nil event from response")
	s.Require().Equal(int64(1), *response.Event[0].UserID, "event user ID should match")
	s.Require().Equal("title 1", *response.Event[0].Title, "event title should match")
	s.Require().Equal("description 1", *response.Event[0].Description,
		"event description should match")
	s.Require().Equal(now, response.Event[0].OnTime.AsTime().Format("2006-01-02T15:04:05Z"),
		"event onTime should match")
	s.Require().Equal(now, response.Event[0].OffTime.AsTime().Format("2006-01-02T15:04:05Z"),
		"event offTime should match")
	s.Require().Equal(now, response.Event[0].NotifyTime.AsTime().Format("2006-01-02T15:04:05Z"),
		"event notifyTime should match")
}

func (s *CalendarSuite) TestCalendar_InsertEvent() {
	UserID := int64(2)
	ID := int64(2)
	title := "title 2"
	description := "description 2"
	onTime, _ := time.Parse(time.RFC3339, "2024-02-21T00:00:00Z")
	offTime, _ := time.Parse(time.RFC3339, "2024-02-23T00:00:00Z")
	notifyTime, _ := time.Parse(time.RFC3339, "2024-02-22T01:01:00Z")

	request := &event_service_v1.ReqByEvent{
		Event: &event_service_v1.Event{
			UserID:      &UserID,
			ID:          &ID,
			Title:       &title,
			Description: &description,
			OnTime:      timestamppb.New(onTime),
			OffTime:     timestamppb.New(offTime),
			NotifyTime:  timestamppb.New(notifyTime),
		},
	}

	// insert event through gRPC
	response, err := s.client.InsertEvent(s.ctx, request)
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil InsertEvent")
	s.Require().NotNil(response.GetID(), "expected a non-nil event ID from response")
	rID := response.GetID()

	// get event from DB
	row := s.db.QueryRow(querySelectByID, rID)
	event := &event_service_v1.Event{}
	var onTimeDatetime, offTimeDateTime, notifyTimeDateTime time.Time
	err = row.Scan(
		&event.ID,
		&event.UserID,
		&event.Title,
		&event.Description,
		&onTimeDatetime,
		&offTimeDateTime,
		&notifyTimeDateTime,
	)

	// compare event from DB with event from gRPC request
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil InsertEvent")
	s.Require().Equal(UserID, *event.UserID, "event user ID should match")
	s.Require().Equal(title, *event.Title, "event title should match")
	s.Require().Equal(description, *event.Description, "event description should match")
	s.Require().Equal(onTime, onTimeDatetime, "event onTime should match")
	s.Require().Equal(offTime, offTimeDateTime, "event offTime should match")
	s.Require().Equal(notifyTime, notifyTimeDateTime, "event notifyTime should match")
	s.Nil(err, "expected no error from scan, but got %v", err)
}

func (s *CalendarSuite) TestCalendar_CreateEvent_Error() {
	// Attempted to create an event with invalid data
	UserID := int64(3)
	ID := int64(3)
	title := "" // Empty title
	description := "description 3"
	onTime, _ := time.Parse(time.RFC3339, "2024-02-21T00:00:00Z")
	offTime, _ := time.Parse(time.RFC3339, "2024-02-23T00:00:00Z")
	notifyTime, _ := time.Parse(time.RFC3339, "2024-02-22T01:01:00Z")

	request := &event_service_v1.ReqByEvent{
		Event: &event_service_v1.Event{
			UserID:      &UserID,
			ID:          &ID,
			Title:       &title,
			Description: &description,
			OnTime:      timestamppb.New(onTime),
			OffTime:     timestamppb.New(offTime),
			NotifyTime:  timestamppb.New(notifyTime),
		},
	}

	// insert event through gRPC
	response, err := s.client.InsertEvent(s.ctx, request)
	s.Require().Error(err)
	s.Require().Nil(response, "expected a nil InsertEvent")
	s.Require().Equal("rpc error: code = Unknown desc = failed to rows.Err: "+
		"ERROR: null value in column \"title\" of relation \"events\" violates not-null constraint (SQLSTATE 23502)",
		err.Error(), "error message should match")
}

func (s *CalendarSuite) TestCalendar_UpdateEvent() {
	// insert event to DB
	now := time.Now().Format("2006-01-02T15:04:05Z")
	row := s.db.QueryRow(
		queryInsert,
		4,
		"title 4",
		"description 4",
		now,
		now,
		now,
	)

	// get event ID from DB
	var ID int64
	_ = row.Scan(&ID)

	// update event through gRPC
	UserID := int64(4)
	title := "title 4 updated"
	description := "description 4 updated"
	onTime, _ := time.Parse(time.RFC3339, "2025-03-11T00:00:00Z")
	offTime, _ := time.Parse(time.RFC3339, "2025-03-13T00:00:00Z")
	notifyTime, _ := time.Parse(time.RFC3339, "2025-03-12T01:01:00Z")

	request := &event_service_v1.ReqByEvent{
		Event: &event_service_v1.Event{
			UserID:      &UserID,
			ID:          &ID,
			Title:       &title,
			Description: &description,
			OnTime:      timestamppb.New(onTime),
			OffTime:     timestamppb.New(offTime),
			NotifyTime:  timestamppb.New(notifyTime),
		},
	}
	response, err := s.client.UpdateEvent(s.ctx, request)
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil UpdateEvent")

	// get updated event from DB
	row = s.db.QueryRow(querySelectByID, ID)
	event := &event_service_v1.Event{}
	var onTimeDatetime, offTimeDateTime, notifyTimeDateTime time.Time
	err = row.Scan(
		&event.ID,
		&event.UserID,
		&event.Title,
		&event.Description,
		&onTimeDatetime,
		&offTimeDateTime,
		&notifyTimeDateTime,
	)

	// compare updated event from DB with event from gRPC request
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil UpdateEvent")
	s.Require().Equal(UserID, *event.UserID, "event user ID should match")
	s.Require().Equal(title, *event.Title, "event title should match")
	s.Require().Equal(description, *event.Description, "event description should match")
	s.Require().Equal(onTime, onTimeDatetime, "event onTime should match")
	s.Require().Equal(offTime, offTimeDateTime, "event offTime should match")
	s.Require().Equal(notifyTime, notifyTimeDateTime, "event notifyTime should match")
	s.Nil(err, "expected no error from scan, but got %v", err)
}

func (s *CalendarSuite) TestCalendar_DeleteEvent() {
	// insert event to DB
	now := time.Now().Format("2006-01-02T15:04:05Z")
	row := s.db.QueryRow(
		queryInsert,
		5,
		"title 5",
		"description 5",
		now,
		now,
		now,
	)

	// get event ID from DB
	var ID int64
	err := row.Scan(&ID)
	s.Require().NoError(err)

	// delete event through gRPC
	request := &event_service_v1.ReqByID{
		ID: &ID,
	}
	response, err := s.client.DeleteEvent(s.ctx, request)
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil DeleteEvent")

	// get event from DB
	row = s.db.QueryRow(querySelectByID, ID)
	event := &event_service_v1.Event{}
	var onTimeDatetime, offTimeDateTime, notifyTimeDateTime time.Time
	err = row.Scan(
		&event.ID,
		&event.UserID,
		&event.Title,
		&event.Description,
		&onTimeDatetime,
		&offTimeDateTime,
		&notifyTimeDateTime,
	)

	// compare event from DB with event from gRPC request
	s.Require().Error(err)
	s.Require().NotNil(response, "expected a non-nil DeleteEvent")
	s.Require().Equal("sql: no rows in result set", err.Error(), "error message should match")
}

func (s *CalendarSuite) TestCalendar_DeleteNonExistingEvent() {
	// delete non-existing event through gRPC
	ID := int64(9999) // non-existing event ID
	request := &event_service_v1.ReqByID{
		ID: &ID,
	}

	// Checking for error and not existing event with fake ID through gRPC
	response, err := s.client.GetEventByID(s.ctx, request)
	s.Require().Error(err)
	s.Require().Nil(response, "expected a nil GetEventByID")
	s.Require().Equal("rpc error: code = Unknown desc = event not found",
		err.Error(), "error message should match")

	// After checking try to delete non-existing event through gRPC
	response1, err := s.client.DeleteEvent(s.ctx, request)

	// Checking for no errors and expected response with nil event
	s.Require().NoError(err)
	s.Require().NotNil(response1, "expected a non-nil DeleteEvent")
	s.Require().Equal("", response1.String(), "expected a nil event from response")
}

func (s *CalendarSuite) TestCalendar_GetAllEvents() {
	// insert one event to DB
	UserID := int64(6)
	now := time.Now().Format("2006-01-02T15:04:05Z")
	_, err := s.db.Exec(
		queryInsert,
		UserID,
		"title 6",
		"description 6",
		now,
		now,
		now,
	)
	s.Require().NoError(err)

	// insert another event to DB with the same user ID but next month
	now1 := time.Now().AddDate(0, 1, 0).Format("2006-01-02T15:04:05Z")
	_, err = s.db.Exec(
		queryInsert,
		UserID,
		"title 6 other",
		"description 6 other",
		now1,
		now1,
		now1,
	)
	s.Require().NoError(err)

	// get all events through gRPC with the same user ID
	request := &event_service_v1.ReqByUser{
		UserID: &UserID,
	}
	response, err := s.client.GetAllEvents(s.ctx, request)

	// compare events from DB with events from gRPC request
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil GetAllEvents")
	s.Require().NotNil(response.GetEvent(), "expected a non-nil event from response")
	s.Require().Equal(2, len(response.Event), "expected 2 events from response")
}

func (s *CalendarSuite) TestCalendar_GetAllEventsDay() {
	UserID := int64(7)
	onTime := time.Date(2024, 01, 21, 0, 0, 0, 0, time.UTC) //nolint:gofumpt
	offTime := time.Date(2024, 01, 23, 0, 0, 0, 0, time.UTC)
	notifyTime := time.Date(2024, 01, 22, 1, 1, 0, 0, time.UTC)

	// insert one event to DB
	_, err := s.db.Exec(
		queryInsert,
		UserID,
		"title 7",
		"description 7",
		onTime,
		offTime,
		notifyTime,
	)
	s.Require().NoError(err)

	// get all events through gRPC with the same user ID and same day
	request := &event_service_v1.ReqByUserByDate{
		UserID: &UserID,
		Date:   timestamppb.New(offTime),
	}
	response, err := s.client.GetAllEventsDay(s.ctx, request)

	// compare events from DB with events from gRPC request
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil GetAllEventsDay")
	s.Require().NotNil(response.GetEvent(), "expected a non-nil event from response")
	s.Require().Equal(1, len(response.Event), "expected 1 event from response")
}

func (s *CalendarSuite) TestCalendar_GetAllEventsDay1() {
	UserID := int64(8)
	onTime := time.Date(2024, 01, 21, 0, 0, 0, 0, time.UTC) //nolint:gofumpt
	offTime := time.Date(2024, 01, 21, 15, 0, 0, 0, time.UTC)
	notifyTime := time.Date(2024, 01, 21, 12, 1, 0, 0, time.UTC)

	// insert one event to DB
	_, err := s.db.Exec(
		queryInsert,
		UserID,
		"title 8",
		"description 8",
		onTime,
		offTime,
		notifyTime,
	)
	s.Require().NoError(err)

	// get all events through gRPC with the same user ID and same day
	request := &event_service_v1.ReqByUserByDate{
		UserID: &UserID,
		Date:   timestamppb.New(onTime),
	}
	response, err := s.client.GetAllEventsDay(s.ctx, request)

	// compare events from DB with events from gRPC request
	s.Require().NoError(err)
	s.Require().NotNil(response, "expected a non-nil GetAllEventsDay")
	s.Require().NotNil(response.GetEvent(), "expected a non-nil event from response")
	s.Require().Len(response.Event, 1, "expected 1 event from response")
}

func (s *CalendarSuite) TestCalendar_GetAllEventsWeek() {
	layout := "2006-01-02 15:04:05"
	currTime, err := time.Parse(layout, "2024-10-07 00:00:01")
	s.Require().NoError(err)
	currTime2 := currTime

	UserID := int64(10)
	ID := int64(10)
	title := "title 10"
	description := "description 10"

	for i := 0; i < 7; i++ {
		request := &event_service_v1.ReqByEvent{
			Event: &event_service_v1.Event{
				UserID:      &UserID,
				ID:          &ID,
				Title:       &title,
				Description: &description,
				OnTime:      timestamppb.New(currTime2),
				OffTime:     timestamppb.New(currTime2.Add(86399 * time.Second)),
				NotifyTime:  timestamppb.New(currTime2),
			},
		}
		currTime2 = currTime2.Add(86400 * time.Second)

		_, err := s.client.InsertEvent(s.ctx, request)
		s.Require().NoError(err)
	}

	founds, err := s.client.GetAllEventsWeek(s.ctx,
		&event_service_v1.ReqByUserByDate{
			UserID: &UserID,
			Date:   timestamppb.New(currTime),
		})
	require.NoError(s.T(), err)
	// Should be 7 events in week from 2024-10-07 Monday to 2024-10-13 Sunday
	require.Len(s.T(), founds.GetEvent(), 7)
}

func (s *CalendarSuite) TestCalendar_GetAllEventsMonth() {
	layout := "2006-01-02 15:04:05"
	currTime, err := time.Parse(layout, "2024-10-01 00:00:01")
	s.Require().NoError(err)
	currTime2 := currTime

	UserID := int64(11)
	ID := int64(11)
	title := "title 11"
	description := "description 11"

	for i := 0; i < 31; i++ {
		request := &event_service_v1.ReqByEvent{
			Event: &event_service_v1.Event{
				UserID:      &UserID,
				ID:          &ID,
				Title:       &title,
				Description: &description,
				OnTime:      timestamppb.New(currTime2),
				OffTime:     timestamppb.New(currTime2.Add(86399 * time.Second)),
				NotifyTime:  timestamppb.New(currTime2),
			},
		}
		currTime2 = currTime2.Add(86400 * time.Second)

		_, err := s.client.InsertEvent(s.ctx, request)
		s.Require().NoError(err)
	}

	founds, err := s.client.GetAllEventsMonth(s.ctx,
		&event_service_v1.ReqByUserByDate{
			UserID: &UserID,
			Date:   timestamppb.New(currTime),
		})
	require.NoError(s.T(), err)
	// Should be 31 events in month from 2024-10-01 Tuesday to 2024-10-31 Thursday
	require.Len(s.T(), founds.GetEvent(), 31)
}
