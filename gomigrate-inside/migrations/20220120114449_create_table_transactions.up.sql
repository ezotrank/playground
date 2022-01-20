create table transactions
(
    id         serial primary key,
    user_id    integer   not null,
    amount     integer   not null,
    created_at timestamp not null default now()
);