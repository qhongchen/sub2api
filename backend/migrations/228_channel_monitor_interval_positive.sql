-- Migration: 228_channel_monitor_interval_positive
-- 渠道监控检测间隔改为任意正整数秒。125 号迁移的 15--3600 秒约束会让
-- 已通过 API 校验的小间隔写入失败，并被统一错误处理显示为 internal error。

ALTER TABLE channel_monitors
    DROP CONSTRAINT IF EXISTS channel_monitors_interval_check;

ALTER TABLE channel_monitors
    ADD CONSTRAINT channel_monitors_interval_check
    CHECK (interval_seconds > 0);
