<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <DashboardMetricGrid
          v-model:show-more="showMoreMetrics"
          :core-cards="coreMetricCards"
          :folded-cards="foldedMetricCards"
        />

        <DashboardUsageChart
          v-model:active-metric="activeUsageMetric"
          :time-range-options="timeRangeOptions"
          :selected-time-range="selectedTimeRange"
          :metric-options="usageMetricOptions"
          :active-metric-config="activeUsageMetricConfig"
          :loading="chartsLoading"
          :chart-data="usageChartData"
          :chart-options="usageChartOptions"
          @time-range-change="applyTimeRange"
        />

        <DashboardRankingGrid
          :user-entries="userRankingEntries"
          :account-entries="accountRankingEntries"
          :model-entries="modelRankingEntries"
          :active-sessions="activeSessionItems"
          :user-loading="rankingLoading"
          :account-loading="accountRankingLoading"
          :model-loading="chartsLoading"
          :sessions-loading="accountRuntimeLoading"
          @view-usage="goToUsageList"
          @view-accounts="goToAccounts"
          @select-user="goToUserUsage"
          @select-account="goToAccountUsage"
          @select-model="goToModelUsage"
        />
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { ChartData, ChartOptions, TooltipItem } from 'chart.js'
import type {
  Account,
  AdminUsageLog,
  DashboardStats,
  ModelStat,
  TrendDataPoint,
  UserSpendingRankingItem
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DashboardMetricGrid from '@/components/admin/dashboard/DashboardMetricGrid.vue'
import DashboardUsageChart from '@/components/admin/dashboard/DashboardUsageChart.vue'
import DashboardRankingGrid from '@/components/admin/dashboard/DashboardRankingGrid.vue'
import type {
  DashboardActiveSessionEntry,
  DashboardMetricCard,
  DashboardRankingEntry,
  DashboardTimeRange,
  UsageMetricKey
} from '@/components/admin/dashboard/types'

const { t } = useI18n()
const appStore = useAppStore()
const router = useRouter()

const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const rankingLoading = ref(false)
const accountRuntimeLoading = ref(false)
const accountRankingLoading = ref(false)
const showMoreMetrics = ref(false)

const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const accounts = ref<Account[]>([])
const accountRankingLogs = ref<AdminUsageLog[]>([])

let chartLoadSeq = 0
let rankingLoadSeq = 0
let accountRuntimeLoadSeq = 0
let accountRankingLoadSeq = 0
const rankingLimit = 12

const comparisonClass = {
  neutral: 'text-gray-600 dark:text-dark-300',
  positive: 'text-emerald-600 dark:text-emerald-400',
  warning: 'text-amber-600 dark:text-amber-400',
  danger: 'text-rose-600 dark:text-rose-400',
  blue: 'text-blue-600 dark:text-blue-400',
  purple: 'text-purple-600 dark:text-purple-400'
}

const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const startOfWeek = (date: Date): Date => {
  const next = new Date(date)
  const day = next.getDay()
  const diff = day === 0 ? -6 : 1 - day
  next.setDate(next.getDate() + diff)
  return next
}

const resolveTimeRange = (range: DashboardTimeRange): { start: string; end: string; granularity: 'day' | 'hour' } => {
  const today = new Date()
  const start = new Date(today)
  const end = new Date(today)

  if (range === 'yesterday') {
    start.setDate(today.getDate() - 1)
    end.setDate(today.getDate() - 1)
    return { start: formatLocalDate(start), end: formatLocalDate(end), granularity: 'hour' }
  }

  if (range === 'this_week') {
    return { start: formatLocalDate(startOfWeek(today)), end: formatLocalDate(end), granularity: 'day' }
  }

  if (range === 'last_7_days') {
    start.setDate(today.getDate() - 6)
    return { start: formatLocalDate(start), end: formatLocalDate(end), granularity: 'day' }
  }

  if (range === 'last_30_days') {
    start.setDate(today.getDate() - 29)
    return { start: formatLocalDate(start), end: formatLocalDate(end), granularity: 'day' }
  }

  if (range === 'this_month') {
    start.setDate(1)
    return { start: formatLocalDate(start), end: formatLocalDate(end), granularity: 'day' }
  }

  return { start: formatLocalDate(start), end: formatLocalDate(end), granularity: 'hour' }
}

const selectedTimeRange = ref<DashboardTimeRange>('today')
const defaultRange = resolveTimeRange(selectedTimeRange.value)
const granularity = ref<'day' | 'hour'>(defaultRange.granularity)
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)
const activeUsageMetric = ref<UsageMetricKey>('cost')

const timeRangeOptions = computed(() => [
  { key: 'today' as const, label: t('admin.dashboard.rangeToday') },
  { key: 'yesterday' as const, label: t('admin.dashboard.rangeYesterday') },
  { key: 'this_week' as const, label: t('admin.dashboard.rangeThisWeek') },
  { key: 'last_7_days' as const, label: t('admin.dashboard.rangeLast7Days') },
  { key: 'last_30_days' as const, label: t('admin.dashboard.rangeLast30Days') },
  { key: 'this_month' as const, label: t('admin.dashboard.rangeThisMonth') }
])

const activeSessionCount = computed(() => (
  accounts.value.reduce((sum, account) => sum + (account.active_sessions ?? 0), 0)
))

const activeSessionAccountCount = computed(() => (
  accounts.value.filter((account) => (account.active_sessions ?? 0) > 0).length
))

const adminMetricCards = computed<DashboardMetricCard[]>(() => {
  const s = stats.value
  if (!s) return []

  return [
    {
      key: 'active-sessions',
      title: t('admin.dashboard.activeSessions'),
      value: formatNumber(activeSessionCount.value),
      subtitle: `${activeSessionAccountCount.value} ${t('admin.dashboard.sessionAccounts')}`,
      badge: 'Live',
      icon: 'terminal',
      accentClass: 'cch-accent-emerald',
      iconClass: 'bg-emerald-500/10 dark:bg-emerald-500/15',
      iconTextClass: 'text-emerald-500',
      comparisons: [
        {
          label: 'RPM',
          value: formatTokens(s.rpm),
          valueClass: comparisonClass.positive
        }
      ]
    },
    {
      key: 'today-requests',
      title: t('admin.dashboard.todayRequests'),
      value: formatNumber(s.today_requests),
      subtitle: `${t('common.total')}: ${formatNumber(s.total_requests)}`,
      badge: '24h',
      icon: 'chart',
      accentClass: 'cch-accent-blue',
      iconClass: 'bg-blue-500/10 dark:bg-blue-500/15',
      iconTextClass: 'text-blue-500',
      comparisons: [
        {
          label: 'RPM',
          value: formatTokens(s.rpm),
          valueClass: comparisonClass.positive
        }
      ]
    },
    {
      key: 'today-tokens',
      title: t('admin.dashboard.todayTokens'),
      value: formatTokens(s.today_tokens),
      subtitle: `${t('common.total')}: ${formatTokens(s.total_tokens)}`,
      badge: 'Tokens',
      icon: 'cube',
      accentClass: 'cch-accent-amber',
      iconClass: 'bg-amber-500/10 dark:bg-amber-500/15',
      iconTextClass: 'text-amber-500',
      comparisons: [
        {
          label: t('admin.dashboard.input'),
          value: formatTokens(s.today_input_tokens),
          valueClass: comparisonClass.blue
        },
        {
          label: t('admin.dashboard.output'),
          value: formatTokens(s.today_output_tokens),
          valueClass: comparisonClass.purple
        }
      ]
    },
    {
      key: 'avg-response',
      title: t('admin.dashboard.avgResponse'),
      value: formatDuration(s.average_duration_ms),
      subtitle: `${s.active_users} ${t('admin.dashboard.activeUsers')}`,
      badge: 'Latency',
      icon: 'clock',
      accentClass: 'cch-accent-purple',
      iconClass: 'bg-purple-500/10 dark:bg-purple-500/15',
      iconTextClass: 'text-purple-500',
      comparisons: [
        {
          label: t('admin.dashboard.averageTime'),
          value: formatDuration(s.average_duration_ms),
          valueClass: comparisonClass.neutral
        }
      ]
    },
    {
      key: 'today-cost',
      title: t('admin.dashboard.todayCost'),
      value: `$${formatCost(s.today_actual_cost)}`,
      subtitle: `${t('common.total')}: $${formatCost(s.total_actual_cost)}`,
      badge: 'USD',
      icon: 'chart',
      accentClass: 'cch-accent-amber',
      iconClass: 'bg-amber-500/10 dark:bg-amber-500/15',
      iconTextClass: 'text-amber-500',
      valueClass: 'text-amber-600 dark:text-amber-400',
      comparisons: [
        {
          label: t('admin.dashboard.accountCost'),
          value: `$${formatCost(s.today_account_cost)}`,
          valueClass: comparisonClass.warning
        },
        {
          label: t('admin.dashboard.standard'),
          value: `$${formatCost(s.today_cost)}`,
          valueClass: comparisonClass.neutral
        }
      ]
    },
    {
      key: 'api-keys',
      title: t('admin.dashboard.apiKeys'),
      value: formatNumber(s.total_api_keys),
      subtitle: `${s.active_api_keys} ${t('common.active')}`,
      badge: 'Keys',
      icon: 'key',
      accentClass: 'cch-accent-blue',
      iconClass: 'bg-blue-500/10 dark:bg-blue-500/15',
      iconTextClass: 'text-blue-500',
      comparisons: [
        {
          label: t('common.active'),
          value: formatNumber(s.active_api_keys),
          valueClass: comparisonClass.positive
        }
      ]
    },
    {
      key: 'accounts',
      title: t('admin.dashboard.accounts'),
      value: formatNumber(s.total_accounts),
      subtitle: `${s.normal_accounts} ${t('common.active')}`,
      badge: 'Pool',
      icon: 'server',
      accentClass: 'cch-accent-purple',
      iconClass: 'bg-purple-500/10 dark:bg-purple-500/15',
      iconTextClass: 'text-purple-500',
      comparisons: [
        {
          label: t('common.active'),
          value: formatNumber(s.normal_accounts),
          valueClass: comparisonClass.positive
        },
        {
          label: t('common.error'),
          value: formatNumber(s.error_accounts),
          valueClass: s.error_accounts > 0 ? comparisonClass.danger : comparisonClass.neutral
        }
      ]
    },
    {
      key: 'users',
      title: t('admin.dashboard.users'),
      value: `+${formatNumber(s.today_new_users)}`,
      subtitle: `${t('common.total')}: ${formatNumber(s.total_users)}`,
      badge: 'Users',
      icon: 'userPlus',
      accentClass: 'cch-accent-emerald',
      iconClass: 'bg-emerald-500/10 dark:bg-emerald-500/15',
      iconTextClass: 'text-emerald-500',
      valueClass: 'text-emerald-600 dark:text-emerald-400',
      comparisons: [
        {
          label: t('admin.dashboard.activeUsers'),
          value: formatNumber(s.active_users),
          valueClass: comparisonClass.positive
        }
      ]
    },
    {
      key: 'total-tokens',
      title: t('admin.dashboard.totalTokens'),
      value: formatTokens(s.total_tokens),
      subtitle: `$${formatCost(s.total_actual_cost)} / $${formatCost(s.total_cost)}`,
      badge: 'All',
      icon: 'database',
      accentClass: 'cch-accent-blue',
      iconClass: 'bg-blue-500/10 dark:bg-blue-500/15',
      iconTextClass: 'text-blue-500',
      comparisons: [
        {
          label: t('admin.dashboard.input'),
          value: formatTokens(s.total_input_tokens),
          valueClass: comparisonClass.blue
        },
        {
          label: t('admin.dashboard.output'),
          value: formatTokens(s.total_output_tokens),
          valueClass: comparisonClass.purple
        }
      ]
    },
    {
      key: 'performance',
      title: t('admin.dashboard.performance'),
      value: formatTokens(s.rpm),
      subtitle: 'RPM',
      badge: 'Live',
      icon: 'bolt',
      accentClass: 'cch-accent-purple',
      iconClass: 'bg-purple-500/10 dark:bg-purple-500/15',
      iconTextClass: 'text-purple-500',
      comparisons: [
        {
          label: 'TPM',
          value: formatTokens(s.tpm),
          valueClass: comparisonClass.purple
        }
      ]
    }
  ]
})

const coreMetricCards = computed(() => adminMetricCards.value.slice(0, 4))
const foldedMetricCards = computed(() => adminMetricCards.value.slice(4))

const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#d1d5db' : '#374151',
  mutedText: isDarkMode.value ? '#9ca3af' : '#6b7280',
  grid: isDarkMode.value ? 'rgba(255,255,255,0.08)' : 'rgba(17,24,39,0.08)',
  cost: '#ff5a00',
  requests: '#2563eb',
  tokens: '#10b981'
}))

const usageTotals = computed(() => {
  return trendData.value.reduce(
    (acc, item) => {
      acc.cost += item.actual_cost || 0
      acc.requests += item.requests || 0
      acc.tokens += item.total_tokens || 0
      return acc
    },
    { cost: 0, requests: 0, tokens: 0 }
  )
})

const usageMetricOptions = computed(() => [
  {
    key: 'cost' as const,
    label: t('admin.dashboard.metricAmount'),
    value: `$${formatCost(usageTotals.value.cost)}`
  },
  {
    key: 'requests' as const,
    label: t('admin.dashboard.metricApiCalls'),
    value: formatNumber(usageTotals.value.requests)
  },
  {
    key: 'tokens' as const,
    label: t('admin.dashboard.metricTokenCount'),
    value: formatTokens(usageTotals.value.tokens)
  }
])

const activeUsageMetricConfig = computed(() => {
  const option = usageMetricOptions.value.find((item) => item.key === activeUsageMetric.value) || usageMetricOptions.value[0]
  const colorMap = {
    cost: chartColors.value.cost,
    requests: chartColors.value.requests,
    tokens: chartColors.value.tokens
  }
  return {
    ...option,
    color: colorMap[option.key]
  }
})

const usageChartData = computed<ChartData<'line', number[], string> | null>(() => {
  if (!trendData.value?.length) return null

  const metric = activeUsageMetric.value
  const color = activeUsageMetricConfig.value.color
  const values = trendData.value.map((item) => {
    if (metric === 'requests') return item.requests || 0
    if (metric === 'tokens') return item.total_tokens || 0
    return item.actual_cost || 0
  })

  return {
    labels: trendData.value.map((item) => formatTrendLabel(item.date)),
    datasets: [
      {
        label: activeUsageMetricConfig.value.label,
        data: values,
        borderColor: color,
        backgroundColor: `${color}30`,
        fill: true,
        tension: 0.42,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHitRadius: 16,
        borderWidth: 2.5
      }
    ]
  }
})

const usageChartOptions = computed<ChartOptions<'line'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      backgroundColor: isDarkMode.value ? '#111827' : '#ffffff',
      borderColor: isDarkMode.value ? 'rgba(255,255,255,0.12)' : 'rgba(17,24,39,0.12)',
      borderWidth: 1,
      titleColor: chartColors.value.text,
      bodyColor: chartColors.value.text,
      displayColors: true,
      callbacks: {
        label: (context: TooltipItem<'line'>) => {
          const value = Number(context.raw || 0)
          if (activeUsageMetric.value === 'cost') {
            return `${activeUsageMetricConfig.value.label}: $${formatCost(value)}`
          }
          if (activeUsageMetric.value === 'tokens') {
            return `${activeUsageMetricConfig.value.label}: ${formatTokens(value)}`
          }
          return `${activeUsageMetricConfig.value.label}: ${formatNumber(value)}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        display: false
      },
      border: {
        display: false
      },
      ticks: {
        color: chartColors.value.mutedText,
        maxRotation: 0,
        autoSkipPadding: 18,
        font: {
          size: 11
        }
      }
    },
    y: {
      beginAtZero: true,
      grid: {
        color: chartColors.value.grid,
        borderDash: [4, 4]
      },
      border: {
        display: false
      },
      ticks: {
        color: chartColors.value.mutedText,
        font: {
          size: 11
        },
        callback: (value: string | number) => {
          const numeric = Number(value)
          if (activeUsageMetric.value === 'cost') return `$${formatCost(numeric)}`
          if (activeUsageMetric.value === 'tokens') return formatTokens(numeric)
          return formatTokens(numeric)
        }
      }
    }
  }
}))

const userRankingEntries = computed<DashboardRankingEntry[]>(() => {
  return rankingItems.value.slice(0, 3).map((item) => ({
    id: `user-${item.user_id}`,
    rawId: item.user_id,
    name: item.email || t('admin.dashboard.unknownUser', { id: item.user_id }),
    requests: item.requests || 0,
    tokens: item.tokens || 0,
    cost: item.actual_cost || 0
  }))
})

const accountRankingEntries = computed<DashboardRankingEntry[]>(() => {
  const accountNameMap = new Map(accounts.value.map((account) => [account.id, account.name || `${account.platform} #${account.id}`]))
  const rankingMap = new Map<number, DashboardRankingEntry>()

  accountRankingLogs.value.forEach((log) => {
    const accountId = log.account_id ?? log.account?.id
    if (!accountId) return

    const entry = rankingMap.get(accountId) || {
      id: `account-${accountId}`,
      rawId: accountId,
      name: log.account?.name || accountNameMap.get(accountId) || `#${accountId}`,
      requests: 0,
      tokens: 0,
      cost: 0
    }

    entry.requests += 1
    entry.tokens += getLogTotalTokens(log)
    entry.cost += getLogAccountCost(log)
    rankingMap.set(accountId, entry)
  })

  return Array.from(rankingMap.values())
    .sort((a, b) => b.cost - a.cost || b.requests - a.requests)
    .slice(0, 3)
})

const getLogTotalTokens = (log: AdminUsageLog): number => {
  return (
    (log.input_tokens || 0) +
    (log.output_tokens || 0) +
    (log.cache_creation_tokens || 0) +
    (log.cache_read_tokens || 0)
  )
}

const getLogAccountCost = (log: AdminUsageLog): number => {
  if (Number.isFinite(log.account_stats_cost)) {
    return Number(log.account_stats_cost)
  }
  return (log.total_cost || 0) * (log.account_rate_multiplier ?? 1)
}

const modelRankingEntries = computed<DashboardRankingEntry[]>(() => {
  return modelStats.value
    .map((item) => ({
      id: `model-${item.model}`,
      rawId: 0,
      name: item.model || t('usage.unknown'),
      requests: item.requests || 0,
      tokens: item.total_tokens || 0,
      cost: item.actual_cost || 0
    }))
    .sort((a, b) => b.cost - a.cost || b.requests - a.requests)
    .slice(0, 3)
})

const activeSessionItems = computed<DashboardActiveSessionEntry[]>(() => {
  return accounts.value
    .filter((account) => (account.active_sessions ?? 0) > 0)
    .sort((a, b) => (b.active_sessions ?? 0) - (a.active_sessions ?? 0))
    .slice(0, 8)
    .map((account) => ({
      id: `session-account-${account.id}`,
      shortId: String(account.id).padStart(4, '0'),
      name: account.name || `${account.platform} #${account.id}`,
      activeSessions: account.active_sessions ?? 0
    }))
})

const formatTrendLabel = (value: string): string => {
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const date = new Date(normalized)
  if (!Number.isNaN(date.getTime())) {
    if (granularity.value === 'hour') {
      return `${String(date.getHours()).padStart(2, '0')}:00`
    }
    return `${date.getMonth() + 1}/${date.getDate()}`
  }

  const hourMatch = value.match(/(\d{2}):\d{2}/)
  if (hourMatch) return `${hourMatch[1]}:00`
  return value
}
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
  } else if (safeValue >= 1) {
    return safeValue.toFixed(2)
  } else if (safeValue >= 0.01) {
    return safeValue.toFixed(3)
  }
  return safeValue.toFixed(4)
}

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}

const applyTimeRange = (range: DashboardTimeRange) => {
  if (selectedTimeRange.value === range && trendData.value.length > 0) return
  const next = resolveTimeRange(range)
  selectedTimeRange.value = range
  startDate.value = next.start
  endDate.value = next.end
  granularity.value = next.granularity
  void loadChartData()
}

const goToUsageList = () => {
  void router.push({
    path: '/admin/usage',
    query: {
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

const goToAccounts = () => {
  void router.push('/admin/accounts')
}

const goToUserUsage = (userId: number) => {
  void router.push({
    path: '/admin/usage',
    query: {
      user_id: String(userId),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

const goToAccountUsage = (accountId: number) => {
  void router.push({
    path: '/admin/usage',
    query: {
      account_id: String(accountId),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

const goToModelUsage = (model: string) => {
  void router.push({
    path: '/admin/usage',
    query: {
      model,
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

const loadDashboardSnapshot = async (includeStats: boolean) => {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) {
    loading.value = true
  }
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) {
      stats.value = response.stats
    }
    trendData.value = response.trend || []
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadUserSpendingRanking = async () => {
  const currentSeq = ++rankingLoadSeq
  rankingLoading.value = true
  try {
    const response = await adminAPI.dashboard.getUserSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit
    })
    if (currentSeq !== rankingLoadSeq) return
    rankingItems.value = response.ranking || []
  } catch (error) {
    if (currentSeq !== rankingLoadSeq) return
    console.error('Error loading user spending ranking:', error)
    rankingItems.value = []
  } finally {
    if (currentSeq === rankingLoadSeq) {
      rankingLoading.value = false
    }
  }
}

const loadAccountRuntime = async () => {
  const currentSeq = ++accountRuntimeLoadSeq
  accountRuntimeLoading.value = true
  try {
    const response = await adminAPI.accounts.list(1, 100, {
      lite: 'true',
      sort_by: 'name',
      sort_order: 'asc'
    })
    if (currentSeq !== accountRuntimeLoadSeq) return
    accounts.value = response.items || []
  } catch (error) {
    if (currentSeq !== accountRuntimeLoadSeq) return
    console.error('Error loading account runtime data:', error)
    accounts.value = []
  } finally {
    if (currentSeq === accountRuntimeLoadSeq) {
      accountRuntimeLoading.value = false
    }
  }
}

const loadAccountRanking = async () => {
  const currentSeq = ++accountRankingLoadSeq
  accountRankingLoading.value = true
  try {
    const response = await adminAPI.usage.list({
      page: 1,
      page_size: 1000,
      start_date: startDate.value,
      end_date: endDate.value,
      sort_by: 'created_at',
      sort_order: 'desc'
    })
    if (currentSeq !== accountRankingLoadSeq) return
    accountRankingLogs.value = response.items || []
  } catch (error) {
    if (currentSeq !== accountRankingLoadSeq) return
    console.error('Error loading account ranking:', error)
    accountRankingLogs.value = []
  } finally {
    if (currentSeq === accountRankingLoadSeq) {
      accountRankingLoading.value = false
    }
  }
}

const loadDashboardStats = async () => {
  await Promise.all([
    loadDashboardSnapshot(true),
    loadUserSpendingRanking(),
    loadAccountRuntime(),
    loadAccountRanking()
  ])
}

const loadChartData = async () => {
  await Promise.all([
    loadDashboardSnapshot(false),
    loadUserSpendingRanking(),
    loadAccountRanking()
  ])
}

onMounted(() => {
  void loadDashboardStats()
})
</script>
