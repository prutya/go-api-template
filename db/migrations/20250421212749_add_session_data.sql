-- migrate:up
alter table "sessions"
add column user_agent text,
add column ip_address text;

-- migrate:down
alter table "sessions"
drop column user_agent,
drop column ip_address;
