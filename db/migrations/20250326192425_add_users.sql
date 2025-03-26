-- migrate:up

create table users (
  id uuid primary key default gen_random_uuid(),
  email text not null,
  password_digest text not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- Unique case-insensitive email
create unique index users_email_unique_idx on users (lower(email));

create table sessions (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id),
  "secret" bytea not null,
  expires_at timestamptz not null,
  terminated_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- migrate:down

drop table "sessions";
drop table users;
