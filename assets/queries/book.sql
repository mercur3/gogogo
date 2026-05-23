-- name: CreateBook :one
insert into books (title, published_at)
values(sqlc.arg('title'), sqlc.arg('published_at'))
returning *;

-- name: PublishBook :exec
insert into book_author (book_id, author_id)
values (sqlc.arg('book_id'), sqlc.arg('author_id'));

-- name: GetBook :one
select * from books where id = sqlc.arg('id');
