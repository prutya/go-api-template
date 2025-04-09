-- migrate:up

-- Create a table for storing refresh tokens
create table refresh_tokens (
  id uuid primary key default gen_random_uuid() not null,
  session_id uuid not null references sessions(id) on delete cascade on update cascade,
  parent_id uuid references refresh_tokens(id) on delete cascade on update cascade,
  secret bytea not null,
  expires_at timestamptz not null,
  revoked_at timestamptz,
  created_at timestamptz default now() not null,
  updated_at timestamptz default now() not null
);

-- Index refresh tokens on session_id and parent_id
create index refresh_tokens_session_id_idx ON refresh_tokens(session_id);
create index refresh_tokens_parent_id_idx ON refresh_tokens(parent_id);

-- Create a table for storing access tokens
create table access_tokens (
  id uuid primary key default gen_random_uuid() not null,
  refresh_token_id uuid not null references refresh_tokens(id) on delete cascade on update cascade,
  secret bytea not null,
  expires_at timestamptz not null,
  created_at timestamptz default now() not null,
  updated_at timestamptz default now() not null
);

-- Index access tokens on refresh_token_id
create index access_tokens_refresh_token_id_idx on access_tokens(refresh_token_id);

-- Drop the secret and expires_at from the sessions table
alter table sessions
  drop column secret,
  drop column expires_at;

-- migrate:down

-- Drop the access_tokens table
drop table if exists access_tokens;

-- Drop the refresh_tokens table
drop table if exists refresh_tokens;

-- Delete all sessions
delete from sessions;

-- Recreate the secret and expires_at columns in the sessions table
alter table sessions
  add column secret bytea not null,
  add column expires_at timestamp with time zone not null;
