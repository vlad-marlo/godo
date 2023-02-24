create table auth_tokens
(
    id         bigserial not null unique primary key,
    user_id    uuid      not null,
    token      text      not null unique,
    expires_at timestamp not null,
    expires    boolean,
    constraint user_id_fk foreign key (user_id) references users (id) match full on delete cascade
);
create index user_id_auth_tokens on auth_tokens (user_id, expires);
---- create above / drop below ----
drop index user_id_auth_tokens;
drop table auth_tokens;
