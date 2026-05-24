<template>
  <div class="cch-bento-grid">
    <div class="cch-metric-card cch-accent-blue">
      <div class="flex items-center justify-between gap-3">
        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('usage.totalRequests') }}</p>
        <span class="rounded-full bg-blue-50 p-2 text-blue-600 dark:bg-blue-500/10 dark:text-blue-300">
          <Icon name="document" size="sm" />
        </span>
      </div>
      <div>
        <p class="text-2xl font-bold tracking-tight text-gray-950 md:text-3xl dark:text-white">
          {{ stats?.total_requests?.toLocaleString() || '0' }}
        </p>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('usage.inSelectedRange') }}</p>
      </div>
    </div>

    <div class="cch-metric-card cch-accent-amber">
      <div class="flex items-center justify-between gap-3">
        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('usage.totalTokens') }}</p>
        <span class="rounded-full bg-amber-50 p-2 text-amber-600 dark:bg-amber-500/10 dark:text-amber-300">
          <Icon name="cube" size="sm" />
        </span>
      </div>
      <div>
        <p class="text-2xl font-bold tracking-tight text-gray-950 md:text-3xl dark:text-white">
          {{ formatTokens(stats?.total_tokens || 0) }}
        </p>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('usage.in') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} /
          {{ t('usage.out') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}
        </p>
      </div>
    </div>

    <div class="cch-metric-card cch-accent-emerald">
      <div class="flex items-center justify-between gap-3">
        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('usage.totalCost') }}</p>
        <span class="rounded-full bg-emerald-50 p-2 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-300">
          <Icon name="dollar" size="sm" />
        </span>
      </div>
      <div class="min-w-0">
        <p class="text-2xl font-bold tracking-tight text-emerald-600 md:text-3xl dark:text-emerald-300">
          ${{ (stats?.total_actual_cost || 0).toFixed(4) }}
        </p>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          <span class="text-orange-500">{{ t('usage.accountCost') }} ${{ (stats?.total_account_cost || 0).toFixed(4) }}</span>
          <span> · </span>
          <span>{{ t('usage.standardCost') }} ${{ (stats?.total_cost || 0).toFixed(4) }}</span>
        </p>
      </div>
    </div>

    <div class="cch-metric-card cch-accent-purple">
      <div class="flex items-center justify-between gap-3">
        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('usage.avgDuration') }}</p>
        <span class="rounded-full bg-purple-50 p-2 text-purple-600 dark:bg-purple-500/10 dark:text-purple-300">
          <Icon name="clock" size="sm" />
        </span>
      </div>
      <div>
        <p class="text-2xl font-bold tracking-tight text-gray-950 md:text-3xl dark:text-white">
          {{ formatDuration(stats?.average_duration_ms || 0) }}
        </p>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('usage.perRequest') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import Icon from '@/components/icons/Icon.vue'

defineProps<{ stats: AdminUsageStatsResponse | null }>()

const { t } = useI18n()

const formatDuration = (ms: number) =>
  ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`

const formatTokens = (value: number) => {
  if (value >= 1e9) return (value / 1e9).toFixed(2) + 'B'
  if (value >= 1e6) return (value / 1e6).toFixed(2) + 'M'
  if (value >= 1e3) return (value / 1e3).toFixed(2) + 'K'
  return value.toLocaleString()
}
</script>
