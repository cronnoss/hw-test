package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/logger"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

const (
	msgInserted string = "Inserted"
	msgUpdated  string = "Updated"
	msgDeleted  string = "Deleted"
)

type ReplayMsg struct {
	Msg string `json:"msg"`
	Err string `json:"error"`
}

func helperDecode(stream io.Reader, r interface{}) error {
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(r); err != nil {
		return err
	}
	return nil
}

func TestCalendarHTTPServer(t *testing.T) {
	userID400 := int64(400)
	bodyUserID := fmt.Sprintf(`{"userid": %d}`, userID400)

	body1 := `{
		"id": 1,
		"userid": 200,
		"title" : "Title_N200",
		"description" : "Description_N200",
		"ontime" : "2015-09-18T00:00:00Z",
		"offtime" : "2015-09-19T00:00:00Z",
		"notifytime" : "0001-01-01T00:00:00Z"
	}`

	body2 := fmt.Sprintf(`{
				"id": 1,
				"userid": %d,
				"title" : "Title_N400",
				"description" : "Description_N400",
				"ontime" : "2015-09-18T00:00:00Z",
				"offtime" : "2015-09-20T00:00:00Z",
				"notifytime" : "0001-01-01T00:00:00Z"
			}`, userID400)

	body3 := fmt.Sprintf(`{
				"id": 2,
				"userid": %d,
				"title" : "Title_N402",
				"description" : "Description_N402",
				"ontime" : "2015-10-18T00:00:00Z",
				"offtime" : "2015-10-20T00:00:00Z",
				"notifytime" : "0001-01-01T00:00:00Z"
			}`, userID400)

	db := memorystorage.New()
	log := logger.NewLogger("DEBUG", os.Stdout)
	calendar := &Calendar{log: log, storage: db}
	httpsrv := internalhttp.NewServer(log, calendar, "", "")
	httpcli := &http.Client{}

	t.Run("case_insert", func(t *testing.T) {
		var rep ReplayMsg

		ts := httptest.NewServer(http.HandlerFunc(httpsrv.InsertEvent))
		defer ts.Close()

		reader := strings.NewReader(body1)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL, reader)
		require.NoError(t, err)
		res, err := httpcli.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		err = helperDecode(res.Body, &rep)
		require.NoError(t, err)

		require.Empty(t, rep.Err)
		require.EqualValues(t, msgInserted, rep.Msg)

		reader = strings.NewReader(body3)
		req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL, reader)
		require.NoError(t, err)
		res, err = httpcli.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		err = helperDecode(res.Body, &rep)
		require.NoError(t, err)

		require.Empty(t, rep.Err)
		require.EqualValues(t, msgInserted, rep.Msg)
	})
	t.Run("case_update", func(t *testing.T) {
		var rep ReplayMsg
		ts := httptest.NewServer(http.HandlerFunc(httpsrv.UpdateEvent))
		defer ts.Close()

		reader := strings.NewReader(body2)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL, reader)
		require.NoError(t, err)
		res, err := httpcli.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		err = helperDecode(res.Body, &rep)
		require.NoError(t, err)

		require.Empty(t, rep.Err)
		require.EqualValues(t, msgUpdated, rep.Msg)
	})

	t.Run("case_lookup", func(t *testing.T) {
		var rep model.Event
		ts := httptest.NewServer(http.HandlerFunc(httpsrv.GetEventByID))
		defer ts.Close()

		reader := strings.NewReader(body1)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL, reader)
		require.NoError(t, err)
		res, err := httpcli.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		err = helperDecode(res.Body, &rep)
		require.NoError(t, err)
		require.EqualValues(t, userID400, rep.UserID)
	})

	t.Run("case_listevents", func(t *testing.T) {
		var rep []model.Event
		ts := httptest.NewServer(http.HandlerFunc(httpsrv.GetAllEvents))
		defer ts.Close()

		reader := strings.NewReader(bodyUserID)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL, reader)
		require.NoError(t, err)
		res, err := httpcli.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		err = helperDecode(res.Body, &rep)
		require.NoError(t, err)
		require.EqualValues(t, int(2), len(rep))
		require.EqualValues(t, userID400, rep[0].UserID)
		require.EqualValues(t, userID400, rep[1].UserID)
	})

	t.Run("case_delete", func(t *testing.T) {
		var rep ReplayMsg
		ts := httptest.NewServer(http.HandlerFunc(httpsrv.DeleteEvent))
		defer ts.Close()

		reader := strings.NewReader(body2)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL, reader)
		require.NoError(t, err)
		res, err := httpcli.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		err = helperDecode(res.Body, &rep)
		require.NoError(t, err)

		require.Empty(t, rep.Err)
		require.EqualValues(t, msgDeleted, rep.Msg)
	})
}
