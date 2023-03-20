create table reviews
(
    id         uuid not null unique primary key,
    task       uuid not null,
    user_id    uuid not null,
    msg        text not null,
    status     text not null,
    created_at timestamp default current_timestamp,
    constraint task_fk foreign key (task) references tasks (id) match full on delete cascade on update cascade,
    constraint user_id_fk foreign key (user_id) references users (id) match full on delete cascade on update cascade
);
---- create above / drop below ----
drop table reviews;