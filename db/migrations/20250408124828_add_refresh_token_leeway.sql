-- migrate:up

alter table refresh_tokens add column leeway_expires_at timestamptz;

update refresh_tokens set leeway_expires_at = revoked_at;

-- Make sure that leeway_expires_at and revoked_at are either both set or not
-- set
alter table refresh_tokens add constraint revoked_at_and_leeway_expires_at_check check (
  (revoked_at is null and leeway_expires_at is null) or
  (revoked_at is not null and leeway_expires_at is not null)
);

-- migrate:down

alter table refresh_tokens drop constraint revoked_at_and_leeway_expires_at_check;

alter table refresh_tokens drop column leeway_expires_at;
