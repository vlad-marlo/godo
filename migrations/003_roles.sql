create table roles
(
    id         bigserial primary key not null unique,
    members    int,
    tasks      int,
    reviews    int,
    "comments" int,
    constraint unique_role unique (members, tasks, reviews, "comments"),
    check ( members >= 0 and tasks >= 0 and reviews >= 0 and "comments" >= 0 )
);
create table user_in_group
(
    id       bigserial primary key not null unique,
    user_id  uuid                  not null,
    group_id uuid                  not null,
    is_admin boolean               not null default false,
    role_id  bigint                not null,
    constraint role_id_fk foreign key (role_id) references roles (id) match full on delete cascade,
    constraint user_id_fk foreign key (user_id) references users (id) match full on delete cascade,
    constraint group_id_fk foreign key (group_id) references groups (id) match full on delete cascade,
    constraint group_user_uq unique (group_id, user_id)
);
create unique index group_user_idx on user_in_group (group_id, user_id);
create table invites
(
    id        uuid not null unique primary key,
    group_id  uuid not null,
    role_id   int,
    use_count integer,
    check ( use_count >= 0 ),
    constraint role_id_fk foreign key (role_id) references roles (id) match full on delete cascade,
    constraint group_id_fk foreign key (group_id) references groups (id) match full on delete cascade
);
---- create above / drop below ----
drop table invites;
drop index group_user_idx;
drop table user_in_group;
drop table roles;