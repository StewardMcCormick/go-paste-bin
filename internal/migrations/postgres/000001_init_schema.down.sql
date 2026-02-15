DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS privacy_policy;

DROP TABLE IF EXISTS paste_info;

DROP INDEX IF EXISTS paste_info_user_id_idx;

DROP INDEX IF EXISTS paste_info_expire_at_idx;

DROP TABLE IF EXISTS paste_content;

DROP TABLE IF EXISTS api_key;

DROP INDEX IF EXISTS api_key_hash_idx;

DROP INDEX IF EXISTS api_key_expire_at_idx;
