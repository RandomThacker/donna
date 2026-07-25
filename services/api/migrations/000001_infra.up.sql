-- Infra bootstrap: enable pgcrypto for future UUID/crypto helpers.
-- No domain tables in M1.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
