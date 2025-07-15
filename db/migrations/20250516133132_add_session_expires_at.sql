-- migrate:up
alter table "sessions" add column expires_at timestamp with time zone not null default now();
create index idx_sessions_user_id on "sessions" (user_id);

-- migrate:down

drop index idx_sessions_user_id;
alter table "sessions" drop column expires_at;
