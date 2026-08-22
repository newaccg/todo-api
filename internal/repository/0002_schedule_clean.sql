-- scheduling cleanup for expired tokens
CREATE EVENT IF NOT EXISTS daily_refresh_tokens_cleanup
ON SCHEDULE EVERY 1 DAY
STARTS CURRENT_TIMESTAMP
DO
  DELETE FROM refresh_tokens
  WHERE expires_at < UNIX_TIMESTAMP()
  LIMIT 5000;
