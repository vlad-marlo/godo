create table tasks
(
    id          uuid not null unique primary key,
    "name"      text not null,
    description text,
    created_at  timestamp default current_timestamp,
    created_by  uuid not null,
    status text not null default 'NEW',
    constraint created_by_fk foreign key (created_by) references users (id) match full on delete cascade
);
create table task_group
(
    id       bigserial unique primary key not null,
    task_id  uuid                         not null,
    group_id uuid                         not null,
    constraint task_id_fk foreign key (task_id) references tasks (id) match full on delete cascade,
    constraint group_id_fk foreign key (group_id) references groups (id) match full on delete cascade,
    constraint task_group_unique unique (task_id, group_id)
);
create unique index task_group_idx on task_group (task_id, group_id);

create table task_user
(
    id      bigserial unique primary key,
    user_id uuid not null,
    task_id uuid not null,
    constraint user_id_fk foreign key (user_id) references users (id) match full on delete cascade,
    constraint task_id_fk foreign key (task_id) references tasks (id) match full on delete cascade
);

---- create above / drop below ----
drop table task_user;
drop index task_group_idx;
drop table task_group;
drop table tasks;