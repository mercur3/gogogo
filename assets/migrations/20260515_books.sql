-- +goose up
CREATE TABLE books (
  id           BIGSERIAL                not null,
  title        varchar(255)             not null,
  published_at timestamp with time zone not null,

  constraint PK_books primary key (id)
);

create table book_author (
    book_id   bigint not null,
    author_id bigint not null,

    constraint PK_book_author primary key (book_id, author_id),
    constraint FK_book_id_ON_book_author foreign key (book_id) references books(id),
    constraint FK_author_id_ON_book_author foreign key (author_id) references authors(id)
);

-- +goose down
drop table book_author;
drop table books;
