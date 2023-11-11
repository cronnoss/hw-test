package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
	query := `INSERT INTO events (userid, title, description, ontime)
							VALUES ($1, $2, $3, $4) RETURNING id`
	rows, err := s.db.QueryxContext(ctx, query, e.UserID, stringNull(e.Title), stringNull(e.Description),
		timeNull(e.OnTime))
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
	res, err := s.db.ExecContext(ctx, query, e.UserID, stringNull(e.Title), stringNull(e.Description),
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

func (s *Storage) DeleteEvent(ctx context.Context, e *storage.Event) error {
	query := `DELETE FROM events WHERE id=$1`
	if _, err := s.db.ExecContext(ctx, query, e.ID); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

func (s *Storage) GetEventByID(ctx context.Context, id int64) (e storage.Event, err error) {
	var eventSQL eventSQL
	query := `SELECT id, userid, title, description, ontime, offtime, notifytime 
			  FROM events WHERE id=$1`
	if err := s.db.GetContext(ctx, &eventSQL, query, id); err != nil {
		return e, fmt.Errorf("failed to get event: %w", err)
	}
	event := ConvertSQLEventToStorageEvent(eventSQL)
	return event, nil
}

func (s Storage) GetAll(ctx context.Context, userID int64) (events []storage.Event, err error) {
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
