-- Create a dev user
insert into users (email, password_digest) values (
  'user@example.com',
  '$2a$12$S5bHLVWs9NyOPxadkRYBeuJnPWQQ86Rm/UZJ3S4tS0Whv7FFrBk6a' -- P@ssw0rd!
) returning *;
