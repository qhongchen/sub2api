<template>
  <div v-if="snapshot" class="space-y-1" data-testid="monitor-quota-view">
    <div v-if="snapshot.plan_level" class="flex flex-wrap items-center gap-1.5">
      <span class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-600 dark:bg-dark-600 dark:text-gray-300">
        {{ snapshot.plan_level }}
      </span>
    </div>

    <div v-if="snapshot.success && tierRows.length" class="space-y-1">
      <div v-for="row in tierRows" :key="row.key" class="flex items-center gap-1.5 text-[10px]">
        <span class="w-14 shrink-0 truncate text-gray-500 dark:text-gray-400" :title="row.label">
          {{ row.label }}
        </span>
        <div class="h-1.5 w-16 shrink-0 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
          <div
            class="h-full rounded-full transition-all"
            :class="utilizationColor(row.tier.used_percent)"
            :style="{ width: `${Math.min(100, Math.max(0, row.tier.used_percent))}%` }"
          />
        </div>
        <span :class="['shrink-0 font-medium', utilizationTextColor(row.tier.used_percent)]">
          {{ Math.round(row.tier.used_percent) }}%
        </span>
        <span v-if="row.tier.reset_at" class="truncate text-gray-400 dark:text-gray-500" :title="row.tier.reset_at">
          &middot; {{ formatReset(row.tier.reset_at) }}
        </span>
      </div>
    </div>

    <div v-if="snapshot.success && balanceRows.length" class="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-[10px]">
      <span
        v-for="balance in balanceRows"
        :key="balance.currency"
        :class="['font-medium', balance.balance <= 0 ? 'text-red-600 dark:text-red-400' : 'text-gray-600 dark:text-gray-300']"
      >
        {{ balance.balance.toFixed(2) }} {{ balance.currency }}
      </span>
    </div>

    <div
      v-if="!snapshot.success"
      class="truncate text-[10px] text-red-600 dark:text-red-400"
      :title="snapshot.error"
      data-testid="monitor-quota-error"
    >
      {{ truncatedError }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorQuotaSnapshot, MonitorQuotaTier } from '@/api/admin/channelMonitor'

const props = defineProps<{
  snapshot?: MonitorQuotaSnapshot | null
}>()

const { t, te } = useI18n()

const windowI18nKeys: Record<string, string> = {
  '5h': 'monitorCommon.quota.windows.5h',
  '7d': 'monitorCommon.quota.windows.7d',
  '7d-sonnet': 'monitorCommon.quota.windows.7dSonnet',
  '7d-fable': 'monitorCommon.quota.windows.7dFable',
  weekly: 'monitorCommon.quota.windows.weekly',
  daily: 'monitorCommon.quota.windows.daily',
  '30d': 'monitorCommon.quota.windows.30d',
  total: 'monitorCommon.quota.windows.total',
}

const labelI18nKeys: Record<string, string> = {
  requests: 'monitorCommon.quota.labels.requests',
  tokens: 'monitorCommon.quota.labels.tokens',
  shared: 'monitorCommon.quota.labels.shared',
  pro: 'monitorCommon.quota.labels.pro',
  flash: 'monitorCommon.quota.labels.flash',
}

function localizedToken(value: string, keys: Record<string, string>): string {
  const key = keys[value]
  return key && te(key) ? t(key) : value
}

function tierLabel(tier: MonitorQuotaTier): string {
  const window = localizedToken(tier.window, windowI18nKeys)
  if (!tier.label) return window
  return `${localizedToken(tier.label, labelI18nKeys)}/${window}`
}

const tierRows = computed(() =>
  (props.snapshot?.tiers || []).map((tier, index) => ({
    key: `${tier.window}-${tier.label || ''}-${index}`,
    label: tierLabel(tier),
    tier,
  }))
)

const balanceRows = computed(() => {
  const snapshot = props.snapshot
  if (!snapshot) return []
  if (snapshot.balances?.length) return snapshot.balances
  if (snapshot.balance != null) {
    return [{ currency: snapshot.currency || '?', balance: snapshot.balance }]
  }
  return []
})

const truncatedError = computed(() => {
  const error = props.snapshot?.error || t('monitorCommon.quota.unavailable')
  return error.length > 48 ? `${error.slice(0, 48)}...` : error
})

const utilizationColor = (percent: number) => {
  if (percent >= 90) return 'bg-red-500'
  if (percent >= 75) return 'bg-amber-500'
  return 'bg-emerald-500'
}

const utilizationTextColor = (percent: number) => {
  if (percent >= 90) return 'text-red-600 dark:text-red-400'
  if (percent >= 75) return 'text-amber-600 dark:text-amber-400'
  return 'text-emerald-600 dark:text-emerald-400'
}

const formatReset = (iso: string) => {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return iso
  const diffMs = date.getTime() - Date.now()
  if (diffMs <= 0) return t('monitorCommon.quota.resetSoon')
  if (diffMs < 3_600_000) return `${Math.max(1, Math.round(diffMs / 60_000))}m`
  const hours = Math.round(diffMs / 3_600_000)
  if (hours < 48) return `${hours}h`
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${month}-${day}`
}
</script>
