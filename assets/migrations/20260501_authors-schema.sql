-- +goose up
CREATE TABLE authors (
  id   BIGSERIAL PRIMARY KEY not null,
  name text      NOT NULL,
  bio  text
);

-- +goose down
drop table authors;
