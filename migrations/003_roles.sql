create table roles
(
    id             bigserial primary key not null unique,
    create_task    boolean default false,
    read_task      boolean default false,
    update_task    boolean default false,
    delete_task    boolean default false,
    create_issue   boolean default false,
    read_issue     boolean default false,
    update_issue   boolean default false,
    review_task    boolean default false,
    read_members   boolean default false,
    invite_members boolean default false,
    delete_members boolean default false
);
create table user_in_group
(
    id       bigserial primary key not null unique,
    user_id  uuid                  not null,
    group_id uuid                  not null,
    role_id  bigint                not null,
    constraint role_id_fk foreign key (role_id) references roles (id) match full on delete cascade,
    constraint user_id_fk foreign key (user_id) references users (id) match full on delete cascade,
    constraint group_id_fk foreign key (group_id) references groups (id) match full on delete cascade,
    constraint group_user_uq unique (group_id, user_id)
);
create unique index group_user_idx on user_in_group (group_id, user_id);
---- create above / drop below ----
drop index group_user_idx;
drop table user_in_group;
drop table roles;