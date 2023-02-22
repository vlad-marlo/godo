create table groups
(
    id          text not null unique primary key,
    "name"      text not null unique,
    description text,
    created_at  timestamp default current_timestamp,
    "owner"     text not null,
    foreign key ("owner") references users (id) MATCH FULL ON DELETE CASCADE
);
create unique index groups_name on groups ("name");
---- create above / drop below ----
drop index groups_name;
drop table groups;