ALTER TABLE usage_logs
  DROP COLUMN IF EXISTS claude_context_1m;

ALTER TABLE accounts
  DROP COLUMN IF EXISTS force_claude_context_1m;

DELETE FROM settings
WHERE key = 'enable_claude_context_1m_force';
