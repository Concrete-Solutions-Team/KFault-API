CREATE TABLE expired_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  token TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_expired_tokens_expires_at ON expired_tokens(expires_at);

CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
RETURNS void AS $$BEGIN
    DELETE FROM expired_tokens WHERE expires_at < NOW();
END;$$ LANGUAGE plpgsql;