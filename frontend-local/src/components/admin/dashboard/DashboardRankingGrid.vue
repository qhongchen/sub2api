<template>
  <section class="grid grid-cols-1 gap-6 md:grid-cols-3">
    <div class="cch-panel-card flex min-h-[280px] min-w-0 flex-col p-5">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-sm font-semibold text-gray-950 dark:text-white">
          {{ t('admin.dashboard.userRanking') }}
        </h3>
        <button
          type="button"
          class="inline-flex items-center gap-0.5 text-xs font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-300"
          @click="emit('viewUsage')"
        >
          <span>{{ t('admin.dashboard.viewAll') }}</span>
          <Icon name="chevronRight" size="xs" :stroke-width="2" />
        </button>
      </div>

      <div class="flex flex-1 flex-col gap-2">
        <template v-if="userLoading">
          <div v-for="idx in 3" :key="idx" class="flex items-center gap-3 rounded-lg p-2">
            <div class="h-8 w-8 animate-pulse rounded-md bg-gray-200/70 dark:bg-white/[0.08]" />
            <div class="flex-1 space-y-2">
              <div class="h-3 w-2/3 animate-pulse rounded bg-gray-200/70 dark:bg-white/[0.08]" />
              <div class="h-1.5 animate-pulse rounded bg-gray-200/70 dark:bg-white/[0.08]" />
            </div>
          </div>
        </template>
        <button
          v-for="(entry, index) in userEntries"
          v-else-if="userEntries.length"
          :key="entry.id"
          type="button"
          class="group flex w-full min-w-0 items-center gap-3 rounded-lg p-2 text-left transition-colors hover:bg-gray-100/70 dark:hover:bg-white/[0.04]"
          @click="emit('selectUser', entry.rawId)"
        >
          <span class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md border text-xs font-semibold" :class="rankBadgeClass(index)">
            {{ index + 1 }}
          </span>
          <span class="min-w-0 flex-1">
            <span class="mb-1.5 flex min-w-0 items-center justify-between gap-3">
              <span class="min-w-0 truncate text-sm font-semibold text-gray-950 dark:text-white">{{ entry.name }}</span>
              <span class="flex-shrink-0 font-mono text-sm font-semibold text-gray-950 dark:text-white">${{ formatCost(entry.cost) }}</span>
            </span>
            <span class="block h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-white/[0.08]">
              <span class="block h-full rounded-full bg-primary-500 transition-all duration-500" :style="{ width: `${rankingPercent(entry.cost, userEntries)}%` }" />
            </span>
            <span class="mt-1 flex min-w-0 items-center justify-between gap-2 text-[10px] text-gray-500 dark:text-dark-400">
              <span class="min-w-0 truncate">{{ formatNumber(entry.requests) }} {{ t('admin.dashboard.requests') }}</span>
              <span class="flex-shrink-0">{{ formatTokens(entry.tokens) }} {{ t('admin.dashboard.tokens') }}</span>
            </span>
          </span>
        </button>
        <div v-else class="flex flex-1 items-center justify-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.dashboard.noDataAvailable') }}
        </div>
      </div>
    </div>

    <div class="cch-panel-card flex min-h-[280px] min-w-0 flex-col p-5">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-sm font-semibold text-gray-950 dark:text-white">
          {{ t('admin.dashboard.accountRanking') }}
        </h3>
        <button
          type="button"
          class="inline-flex items-center gap-0.5 text-xs font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-300"
          @click="emit('viewAccounts')"
        >
          <span>{{ t('admin.dashboard.viewAll') }}</span>
          <Icon name="chevronRight" size="xs" :stroke-width="2" />
        </button>
      </div>

      <div class="flex flex-1 flex-col gap-2">
        <template v-if="accountLoading">
          <div v-for="idx in 3" :key="idx" class="flex items-center gap-3 rounded-lg p-2">
            <div class="h-8 w-8 animate-pulse rounded-md bg-gray-200/70 dark:bg-white/[0.08]" />
            <div class="flex-1 space-y-2">
              <div class="h-3 w-2/3 animate-pulse rounded bg-gray-200/70 dark:bg-white/[0.08]" />
              <div class="h-1.5 animate-pulse rounded bg-gray-200/70 dark:bg-white/[0.08]" />
            </div>
          </div>
        </template>
        <button
          v-for="(entry, index) in accountEntries"
          v-else-if="accountEntries.length"
          :key="entry.id"
          type="button"
          class="group flex w-full min-w-0 items-center gap-3 rounded-lg p-2 text-left transition-colors hover:bg-gray-100/70 dark:hover:bg-white/[0.04]"
          @click="emit('selectAccount', entry.rawId)"
        >
          <span class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md border text-xs font-semibold" :class="rankBadgeClass(index)">
            {{ index + 1 }}
          </span>
          <span class="min-w-0 flex-1">
            <span class="mb-1.5 flex min-w-0 items-center justify-between gap-3">
              <span class="min-w-0 truncate text-sm font-semibold text-gray-950 dark:text-white">{{ entry.name }}</span>
              <span class="flex-shrink-0 font-mono text-sm font-semibold text-gray-950 dark:text-white">${{ formatCost(entry.cost) }}</span>
            </span>
            <span class="block h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-white/[0.08]">
              <span class="block h-full rounded-full bg-purple-500 transition-all duration-500" :style="{ width: `${rankingPercent(entry.cost, accountEntries)}%` }" />
            </span>
            <span class="mt-1 flex min-w-0 items-center justify-between gap-2 text-[10px] text-gray-500 dark:text-dark-400">
              <span class="min-w-0 truncate">{{ formatNumber(entry.requests) }} {{ t('admin.dashboard.requests') }}</span>
              <span class="flex-shrink-0">{{ formatTokens(entry.tokens) }} {{ t('admin.dashboard.tokens') }}</span>
            </span>
          </span>
        </button>
        <div v-else class="flex flex-1 items-center justify-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.dashboard.noDataAvailable') }}
        </div>
      </div>
    </div>

    <div class="cch-panel-card flex min-h-[280px] min-w-0 flex-col p-5">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-sm font-semibold text-gray-950 dark:text-white">
          {{ t('admin.dashboard.modelRanking') }}
        </h3>
        <button
          type="button"
          class="inline-flex items-center gap-0.5 text-xs font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-300"
          @click="emit('viewUsage')"
        >
          <span>{{ t('admin.dashboard.viewAll') }}</span>
          <Icon name="chevronRight" size="xs" :stroke-width="2" />
        </button>
      </div>

      <div class="flex flex-1 flex-col gap-2">
        <template v-if="modelLoading">
          <div v-for="idx in 3" :key="idx" class="flex items-center gap-3 rounded-lg p-2">
            <div class="h-8 w-8 animate-pulse rounded-md bg-gray-200/70 dark:bg-white/[0.08]" />
            <div class="flex-1 space-y-2">
              <div class="h-3 w-2/3 animate-pulse rounded bg-gray-200/70 dark:bg-white/[0.08]" />
              <div class="h-1.5 animate-pulse rounded bg-gray-200/70 dark:bg-white/[0.08]" />
            </div>
          </div>
        </template>
        <button
          v-for="(entry, index) in modelEntries"
          v-else-if="modelEntries.length"
          :key="entry.id"
          type="button"
          class="group flex w-full min-w-0 items-center gap-3 rounded-lg p-2 text-left transition-colors hover:bg-gray-100/70 dark:hover:bg-white/[0.04]"
          @click="emit('selectModel', entry.name)"
        >
          <span class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md border text-xs font-semibold" :class="rankBadgeClass(index)">
            {{ index + 1 }}
          </span>
          <span class="min-w-0 flex-1">
            <span class="mb-1.5 flex min-w-0 items-center justify-between gap-3">
              <span class="min-w-0 truncate text-sm font-semibold text-gray-950 dark:text-white">{{ entry.name }}</span>
              <span class="flex-shrink-0 font-mono text-sm font-semibold text-gray-950 dark:text-white">${{ formatCost(entry.cost) }}</span>
            </span>
            <span class="block h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-white/[0.08]">
              <span class="block h-full rounded-full bg-blue-500 transition-all duration-500" :style="{ width: `${rankingPercent(entry.cost, modelEntries)}%` }" />
            </span>
            <span class="mt-1 flex min-w-0 items-center justify-between gap-2 text-[10px] text-gray-500 dark:text-dark-400">
              <span class="min-w-0 truncate">{{ formatNumber(entry.requests) }} {{ t('admin.dashboard.requests') }}</span>
              <span class="flex-shrink-0">{{ formatTokens(entry.tokens) }} {{ t('admin.dashboard.tokens') }}</span>
            </span>
          </span>
        </button>
        <div v-else class="flex flex-1 items-center justify-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.dashboard.noDataAvailable') }}
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { DashboardRankingEntry } from './types'

defineProps<{
  userEntries: DashboardRankingEntry[]
  accountEntries: DashboardRankingEntry[]
  modelEntries: DashboardRankingEntry[]
  userLoading: boolean
  accountLoading: boolean
  modelLoading: boolean
}>()

const emit = defineEmits<{
  (event: 'viewUsage'): void
  (event: 'viewAccounts'): void
  (event: 'selectUser', id: number): void
  (event: 'selectAccount', id: number): void
  (event: 'selectModel', model: string): void
}>()

const { t } = useI18n()

const formatTokens = (value: number | undefined | null): string => {
  if (value === undefined || value === null) return '0'
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return Math.round(value).toLocaleString()
}

const formatNumber = (value: number | undefined | null): string => {
  return Number(value || 0).toLocaleString()
}

const formatCost = (value?: number | null): string => {
  const safeValue = Number.isFinite(value) ? Number(value) : 0
  if (safeValue >= 1000) {
    return (safeValue / 1000).toFixed(2) + 'K'
  }
  return safeValue.toFixed(2)
}

const rankBadgeClass = (index: number): string => {
  if (index === 0) return 'border-amber-200 bg-amber-50 text-amber-600 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-300'
  if (index === 1) return 'border-slate-200 bg-slate-50 text-slate-500 dark:border-slate-500/20 dark:bg-slate-500/10 dark:text-slate-300'
  return 'border-orange-200 bg-orange-50 text-orange-600 dark:border-orange-500/20 dark:bg-orange-500/10 dark:text-orange-300'
}

const rankingPercent = (cost: number, entries: DashboardRankingEntry[]): number => {
  const max = Math.max(...entries.map((item) => item.cost), 0)
  if (max <= 0) return 0
  return Math.max(4, Math.min(100, (cost / max) * 100))
}
</script>
