-- name: GetAuthor :one
SELECT * FROM authors
WHERE id = sqlc.arg('id') LIMIT 1;

-- name: ListAuthors :many
SELECT * FROM authors
ORDER BY name;

-- name: CreateAuthor :one
INSERT INTO authors (
  name, bio
) VALUES (
  sqlc.arg('name'), sqlc.arg('bio')
)
RETURNING *;

-- name: UpdateAuthor :execrows
UPDATE authors
  set name = sqlc.arg('name'),
  bio = sqlc.arg('bio')
WHERE id = sqlc.arg('id');

-- name: DeleteAuthor :execrows
DELETE FROM authors
WHERE id = sqlc.arg('id');
