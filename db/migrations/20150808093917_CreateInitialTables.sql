
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE channels (
  id INTEGER PRIMARY KEY ASC,
  Name VARCHAR(64) NOT NULL UNIQUE,
  ChannelSubscribers TEXT,
  Notifiers TEXT,
  NotificationTime TEXT
);

CREATE TABLE notifications (
  id INTEGER PRIMARY KEY ASC,
  ChannelName VARCHAR(64) NOT NULL,
  Subject TEXT NOT NULL,
  Content TEXT,
  Tags TEXT,
  Priority INTEGER
);

CREATE TABLE api_tokens (
  id INTEGER PRIMARY KEY ASC,
  Token VARCHAR(64)
);

CREATE TABLE users (
  id INTEGER PRIMARY KEY ASC,
  Email VARCHAR(255) UNIQUE,
  Name VARCHAR(255)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE channels;
DROP TABLE notifications;
DROP TABLE api_tokens;
DROP TABLE users;

