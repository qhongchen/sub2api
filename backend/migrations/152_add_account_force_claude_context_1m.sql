-- Add force_claude_context_1m column to accounts table
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS force_claude_context_1m BOOLEAN NOT NULL DEFAULT FALSE;
COMMENT ON COLUMN accounts.force_claude_context_1m IS 'Force enable Claude 1M context for this account';
