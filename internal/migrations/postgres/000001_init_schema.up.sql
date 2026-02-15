CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    username varchar(30) UNIQUE NOT NULL,
    password_hash varchar(60) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TYPE privacy_policy AS ENUM('public', 'private', 'protected');

CREATE TABLE IF NOT EXISTS paste_info(
    id SERIAL primary key,
    user_id int not null references users(id) on delete cascade,
    paste_hash varchar(20) unique not null,
    views int NOT NULL DEFAULT 0,
    privacy privacy_policy NOT NULL DEFAULT 'public',
    password_hash varchar(60),
    created_at timestamptz NOT NULL DEFAULT now(),
    expire_at timestamptz
);

CREATE INDEX IF NOT EXISTS paste_info_user_id_idx ON paste_info(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS paste_info_expire_at_idx ON paste_info(expire_at) WHERE expire_at is not null;

CREATE TABLE IF NOT EXISTS paste_content(
    paste_id int primary key references paste_info(id) on delete cascade,
    content text
);

CREATE TABLE IF NOT EXISTS api_key(
    key_hash varchar(64) primary key,
    user_id int not null references users(id) on delete cascade,
    created_at timestamptz NOT NULL DEFAULT now(),
    expire_at timestamptz NOT NULL
    );

CREATE INDEX IF NOT EXISTS api_key_hash_idx ON api_key USING HASH (key_hash);

CREATE INDEX IF NOT EXISTS api_key_expire_at_idx ON api_key(expire_at);
