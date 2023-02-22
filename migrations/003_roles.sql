create table roles
(
    id             bigserial primary key not null unique,
    create_task    boolean default false,
    read_task      boolean default false,
    update_task    boolean default false,
    delete_task    boolean default false,
    create_issue   boolean default false,
    read_issue     boolean default false,
    review_task    boolean default false,
    read_members   boolean default false,
    invite_members boolean default false,
    delete_members boolean default false
);
create table role_user
(
    id       bigserial primary key not null unique,
    user_id  text,
    group_id text,
    constraint user_id_fk foreign key (user_id) references users (id) match full on delete cascade,
    constraint group_id_fk foreign key (group_id) references groups (id) match full on delete cascade,
    constraint group_user_uq unique (group_id, user_id)
);
create unique index group_user_idx on role_user (group_id, user_id);
---- create above / drop below ----
drop index group_user_idx;
drop table role_user;
drop table roles;