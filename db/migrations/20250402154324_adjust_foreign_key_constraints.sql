-- migrate:up

-- Migration to update foreign key constraints to cascade on update/delete

-- First, drop the existing foreign key constraints
alter table sessions drop constraint sessions_user_id_fkey;

-- then recreate them with cascade options
alter table sessions
  add constraint sessions_user_id_fkey
  foreign key (user_id)
  references users(id)
  on update cascade
  on delete cascade;

-- migrate:down

-- first, drop the cascade-enabled foreign key constraints
alter table sessions drop constraint sessions_user_id_fkey;

-- then recreate them with the original settings (no cascade options)
alter table sessions
  add constraint sessions_user_id_fkey
  foreign key (user_id)
  references users(id);
