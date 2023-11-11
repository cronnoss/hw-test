-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events(
                                     id               SERIAL PRIMARY KEY,
                                     title            VARCHAR (150) NOT NULL,
                                     ontime           TIMESTAMP NOT NULL,
                                     offtime          TIMESTAMP,
                                     description      TEXT,
                                     userid           BIGINT NOT NULL,
                                     notifytime       TIMESTAMP
);

CREATE INDEX IF NOT EXISTS events_userid_idx ON events (userid);
CREATE INDEX IF NOT EXISTS events_ontime_idx ON events (ontime);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS events_userid_idx;
DROP INDEX IF EXISTS events_ontime_idx;
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
