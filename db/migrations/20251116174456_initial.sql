-- migrate:up

create table users (
  id uuid primary key default gen_random_uuid(),
  email text not null,
  password_digest text not null,

  -- Email verification
  email_verified_at timestamptz,
  email_verification_otp_hmac bytea,
  email_verification_expires_at timestamptz,
  email_verification_otp_attempts int not null default 0,
  email_verification_cooldown_resets_at timestamptz,
  email_verification_last_requested_at timestamptz,

  -- Password reset
  password_reset_otp_hmac bytea,
  password_reset_expires_at timestamptz,
  password_reset_otp_attempts int not null default 0,
  password_reset_cooldown_resets_at timestamptz,
  password_reset_last_requested_at timestamptz,

  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- Unique case-insensitive email
create unique index users_email_unique_idx on users (lower(email));

create table sessions (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on update cascade on delete cascade,
  terminated_at timestamptz,
  expires_at timestamptz not null default now(),
  user_agent text,
  ip_address text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- Index sessions on user_id
create index idx_sessions_user_id on "sessions" (user_id);

-- Create a table for storing refresh tokens
create table refresh_tokens (
  id uuid primary key default gen_random_uuid() not null,
  session_id uuid not null references sessions(id) on delete cascade on update cascade,
  parent_id uuid references refresh_tokens(id) on delete cascade on update cascade,
  public_key bytea not null,
  expires_at timestamptz not null,
  revoked_at timestamptz,
  leeway_expires_at timestamptz,
  created_at timestamptz default now() not null,
  updated_at timestamptz default now() not null,

  -- Make sure that leeway_expires_at and revoked_at are either both set or not
  -- set
  constraint revoked_at_and_leeway_expires_at_check check (
    (revoked_at is null and leeway_expires_at is null) or
    (revoked_at is not null and leeway_expires_at is not null)
  )
);

-- Index refresh tokens on session_id and parent_id
create index refresh_tokens_session_id_idx ON refresh_tokens(session_id);
create index refresh_tokens_parent_id_idx ON refresh_tokens(parent_id);

-- Create a table for storing access tokens
create table access_tokens (
  id uuid primary key default gen_random_uuid() not null,
  refresh_token_id uuid not null references refresh_tokens(id) on delete cascade on update cascade,
  public_key bytea not null,
  expires_at timestamptz not null,
  created_at timestamptz default now() not null,
  updated_at timestamptz default now() not null
);

-- Index access tokens on refresh_token_id
create index access_tokens_refresh_token_id_idx on access_tokens(refresh_token_id);

create table email_send_attempts (
  id serial primary key,
  attempted_at timestamptz not null default now()
);

create index email_send_attempts_attempted_at_idx on email_send_attempts (attempted_at);

-- migrate:down
drop table email_send_attempts;
drop table access_tokens;
drop table refresh_tokens;
drop table "sessions";
drop table users;
