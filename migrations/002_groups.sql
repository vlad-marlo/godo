create table groups
(
    id          uuid not null unique primary key,
    "name"      text not null unique,
    description text,
    created_at  timestamp default current_timestamp,
    "owner"     uuid not null,
    foreign key ("owner") references users (id) MATCH FULL ON DELETE CASCADE
);
create unique index groups_name on groups ("name");

create table invites
(
    id        uuid not null unique primary key,
    group_id  uuid not null,
    use_count integer,
    check ( use_count >= 0 ),
    constraint group_id_fk foreign key (group_id) references groups (id) match full on delete cascade
);
---- create above / drop below ----
drop table invites;
drop index groups_name;
drop table groups;
