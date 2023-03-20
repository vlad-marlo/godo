set time zone 'Europe/London';
create table users
(
    id    uuid        not null unique primary key,
    email text unique not null,
    pass  text        not null
);
create unique index users_username_idx on users (email);
---- create above / drop below ----
drop index users_username_idx;
drop table users;
