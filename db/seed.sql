-- Create a dev user
insert into users (email, password_digest) values (
  'user@example.com',
  '$argon2id$v=19$m=65536,t=3,p=2$xT4ZrSEE4bmLtnoo1nDjXw$KDWd62L27zZSjpnElxcVJ5EDv2am0qhwkIc4T0pb3yg' -- P@ssw0rd!
) returning *;
