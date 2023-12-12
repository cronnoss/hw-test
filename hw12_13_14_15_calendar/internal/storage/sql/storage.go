package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	dsn string
	db  *sqlx.DB
}

type eventSQL struct {
	ID          sql.NullInt64  `db:"id"`
	UserID      sql.NullInt64  `db:"userid"`
	Title       sql.NullString `db:"title"`
	Description sql.NullString `db:"description"`
	OnTime      sql.NullTime   `db:"ontime"`
	OffTime     sql.NullTime   `db:"offtime"`
	NotifyTime  sql.NullTime   `db:"notifytime"`
}

func ConvertSQLEventToStorageEvent(e eventSQL) (event storage.Event) {
	if e.ID.Valid {
		event.ID = e.ID.Int64
	}

	if e.UserID.Valid {
		event.UserID = e.UserID.Int64
	}

	if e.Title.Valid {
		event.Title = e.Title.String
	}

	if e.Description.Valid {
		event.Description = e.Description.String
	}

	if e.OnTime.Valid {
		event.OnTime = e.OnTime.Time
	}

	if e.OffTime.Valid {
		event.OffTime = e.OffTime.Time
	}

	if e.NotifyTime.Valid {
		event.NotifyTime = e.NotifyTime.Time
	}
	return event
}

func New(dsn string) *Storage {
	return &Storage{dsn: dsn}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to load driver: %w", err)
	}
	s.db = db
	err = s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	s.db.Close()
	ctx.Done()
	return nil
}

func timeNull(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func stringNull(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func (s *Storage) InsertEvent(ctx context.Context, e *storage.Event) error {
	query := `INSERT INTO events (userid, title, description, ontime, offtime, notifytime)
							VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	rows, err := s.db.QueryxContext(ctx, query, e.UserID, stringNull(e.Title), stringNull(e.Description),
		timeNull(e.OnTime), timeNull(e.OffTime), timeNull(e.NotifyTime))
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&e.ID)
		if err != nil {
			return fmt.Errorf("failed to rows.Scan: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to rows.Err: %w", err)
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, e *storage.Event) error {
	query := `UPDATE events SET userid=$2, 
                  				title=$3, 
								description=$4, 
                  				ontime=$5, 
                  				offtime=$6, 
                  				notifytime=$7 
              WHERE id=$1`
	res, err := s.db.ExecContext(ctx, query, e.ID, e.UserID, stringNull(e.Title), stringNull(e.Description),
		timeNull(e.OnTime), timeNull(e.OffTime), timeNull(e.NotifyTime))
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get RowsAffected: %w", err)
	}

	if ra != 1 {
		return fmt.Errorf("failed to update event: %v", ra)
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id int64) error {
	query := `DELETE FROM events WHERE id=$1`
	if _, err := s.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

func (s *Storage) GetAllRange(ctx context.Context, userID int64, begin, end time.Time) ([]storage.Event, error) {
	var e storage.Event
	var events []storage.Event
	var eSQL eventSQL

	query := `SELECT id, userid, title, description, ontime, offtime, notifytime
	          FROM events
			  WHERE userid = $1 AND 
			  (ontime BETWEEN $2 AND $3 OR offtime BETWEEN $2 AND $3)`

	rows, err := s.db.QueryContext(ctx, query, userID, begin, end)
	if err != nil {
		return events, fmt.Errorf("failed lookup event: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&eSQL.ID, &eSQL.UserID, &eSQL.Title, &eSQL.Description,
			&eSQL.OnTime, &eSQL.OffTime, &eSQL.NotifyTime); err != nil {
			return events, fmt.Errorf("failed rows.Scan: %w", err)
		}
		e = ConvertSQLEventToStorageEvent(eSQL)
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return events, fmt.Errorf("failed lookup event: %w", err)
	}

	return events, nil
}

func (s *Storage) GetEventByID(ctx context.Context, id int64) (e storage.Event, err error) {
	var eventSQL eventSQL
	query := `SELECT id, userid, title, description, ontime, offtime, notifytime 
			  FROM events WHERE id=$1`
	rows := s.db.QueryRowxContext(ctx, query, id)

	if err := rows.StructScan(&eventSQL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e, app.ErrEventNotFound
		}
		return e, fmt.Errorf("failed rows.StructScan: %w", err)
	}

	if err := rows.Err(); err != nil {
		return e, fmt.Errorf("failed rows.Err: %w", err)
	}

	e = ConvertSQLEventToStorageEvent(eventSQL)
	return e, err
}

func (s Storage) GetAllEvents(ctx context.Context, userID int64) (events []storage.Event, err error) {
	var e storage.Event
	var eSQL eventSQL

	query := `SELECT id, userid, title, description, ontime, offtime, notifytime
			  FROM events WHERE userid=$1`
	rows, err := s.db.Queryx(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.StructScan(&eSQL); err != nil {
			return nil, fmt.Errorf("failed to rows.StructScan: %w", err)
		}
		e = ConvertSQLEventToStorageEvent(eSQL)
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to rows.Err: %w", err)
	}

	return events, nil
}

func (s *Storage) IsBusyDateTimeRange(ctx context.Context, id, userID int64, onTime, offTime time.Time) error {
	var eSQL eventSQL
	query := `SELECT id
	          FROM events
			  WHERE id != $1 AND userid = $2
			  (($3 BETWEEN event.ontime and event.offtime) OR
			   ($4 BETWEEN event.ontime and event.offtime))`

	rows := s.db.QueryRowContext(ctx, query, id, userID, onTime, offTime)

	if err := rows.Scan(&eSQL.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed rows.Scan: %w", err)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed rows.Next: %w", err)
	}

	return app.ErrDateBusy
}
