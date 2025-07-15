-- migrate:up

create table email_verification_tokens (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade on update cascade,
  secret bytea not null,
  expires_at timestamptz not null,
  sent_at timestamptz,
  verified_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index email_verification_tokens_user_id_idx on email_verification_tokens (user_id);

create table password_reset_tokens (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade on update cascade,
  secret bytea not null,
  expires_at timestamptz not null,
  sent_at timestamptz,
  reset_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index password_reset_tokens_user_id_idx on password_reset_tokens (user_id);

alter table users add column email_verification_rate_limited_until timestamptz;
alter table users add column email_verified_at timestamptz;
alter table users add column password_reset_rate_limited_until timestamptz;

-- migrate:down

alter table users drop column email_verified_at;
alter table users drop column email_verification_rate_limited_until;

drop table password_reset_tokens;
drop table email_verification_tokens;
