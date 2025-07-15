-- migrate:up

create table email_send_attempts (
  id serial primary key,
  attempted_at timestamptz not null default now()
);

create index email_send_attempts_attempted_at_idx on email_send_attempts (attempted_at);

-- migrate:down

drop table email_send_attempts;
