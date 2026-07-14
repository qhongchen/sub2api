<template>
  <span v-if="quotas === undefined" class="text-xs text-gray-400 dark:text-gray-500">
    ...
  </span>
  <span
    v-else-if="configured.length === 0"
    class="text-xs text-gray-400 dark:text-gray-500"
  >
    {{ t('admin.users.platformQuota.cellNotConfigured') }}
  </span>
  <div v-else class="space-y-0.5 text-xs">
    <div
      v-for="row in configured"
      :key="row.platform"
      class="flex items-center gap-2 whitespace-nowrap"
    >
      <span class="w-20 shrink-0 font-mono text-gray-700 dark:text-gray-300">
        {{ row.platform }}
      </span>
      <span class="text-gray-500 dark:text-gray-400">
        {{ t('admin.users.platformQuota.windowDaily') }}
        <span class="text-gray-900 dark:text-white">
          {{ formatUsd(row.daily_usage_usd) }}/{{ formatLimit(row.daily_limit_usd) }}
        </span>
      </span>
      <span class="text-gray-500 dark:text-gray-400">
        {{ t('admin.users.platformQuota.windowWeekly') }}
        <span class="text-gray-900 dark:text-white">
          {{ formatUsd(row.weekly_usage_usd) }}/{{ formatLimit(row.weekly_limit_usd) }}
        </span>
      </span>
      <span class="text-gray-500 dark:text-gray-400">
        {{ t('admin.users.platformQuota.windowMonthly') }}
        <span class="text-gray-900 dark:text-white">
          {{ formatUsd(row.monthly_usage_usd) }}/{{ formatLimit(row.monthly_limit_usd) }}
        </span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PlatformQuotaItem, PlatformQuotaPlatform } from '@/api/admin/users'

const props = defineProps<{ quotas?: PlatformQuotaItem[] }>()
const { t } = useI18n()
const PLATFORM_ORDER: PlatformQuotaPlatform[] = [
  'anthropic',
  'openai',
  'gemini',
  'antigravity',
  'grok'
]

const configured = computed(() => {
  if (!props.quotas) return []
  return props.quotas
    .filter(
      (quota) =>
        quota.daily_limit_usd != null ||
        quota.weekly_limit_usd != null ||
        quota.monthly_limit_usd != null
    )
    .slice()
    .sort(
      (left, right) =>
        PLATFORM_ORDER.indexOf(left.platform) - PLATFORM_ORDER.indexOf(right.platform)
    )
})

function formatUsd(value: number): string {
  if (!Number.isFinite(value)) return '0'
  return String(Math.round(value * 100) / 100)
}

function formatLimit(value: number | null): string {
  return value == null ? '-' : formatUsd(value)
}
</script>
