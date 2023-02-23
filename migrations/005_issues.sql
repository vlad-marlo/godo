create table issues
(
    id         uuid not null unique primary key,
    task       uuid not null,
    msg        text,
    created_by uuid not null,
    created_at timestamp default current_timestamp,
    constraint created_by_fk foreign key (created_by) references users (id) match full on delete cascade,
    constraint task_fk foreign key (task) references tasks (id) match full on delete cascade
);
---- create above / drop below ----
drop table issues;
