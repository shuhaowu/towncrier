
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE notifications (
  id INTEGER PRIMARY KEY ASC,
  Channel VARCHAR(64) NOT NULL,
  Subject TEXT NOT NULL,
  Content TEXT,
  Origin TEXT,
  TagsString TEXT,
  PriorityInt INTEGER
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE notifications;

