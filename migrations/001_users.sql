create table users
(
    id       uuid        not null unique primary key,
    username text unique not null,
    email    text unique not null,
    pass     text        not null
);
create unique index users_username_idx on users (username);
---- create above / drop below ----
drop index users_username_idx;
drop table users;
