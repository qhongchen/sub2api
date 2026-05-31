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
          :legend-items="usageLegendItems"
          :loading="chartsLoading"
          :chart-data="usageChartData"
          :chart-options="usageChartOptions"
          @time-range-change="applyTimeRange"
          @toggle-series="toggleUsageSeries"
        />

        <DashboardRankingGrid
          :user-entries="userRankingEntries"
          :account-entries="accountRankingEntries"
          :model-entries="modelRankingEntries"
          :user-loading="rankingLoading"
          :account-loading="accountRankingLoading"
          :model-loading="chartsLoading"
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
  DashboardStats,
  ModelStat,
  UserUsageTrendPoint,
  UserSpendingRankingItem,
  AccountSpendingRankingItem
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DashboardMetricGrid from '@/components/admin/dashboard/DashboardMetricGrid.vue'
import DashboardUsageChart from '@/components/admin/dashboard/DashboardUsageChart.vue'
import DashboardRankingGrid from '@/components/admin/dashboard/DashboardRankingGrid.vue'
import type {
  DashboardMetricCard,
  DashboardRankingEntry,
  DashboardTimeRange,
  DashboardUsageLegendItem,
  UsageMetricKey
} from '@/components/admin/dashboard/types'

const { t } = useI18n()
const appStore = useAppStore()
const router = useRouter()

const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const rankingLoading = ref(false)
const accountRankingLoading = ref(false)
const showMoreMetrics = ref(false)

const userTrend = ref<UserUsageTrendPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const accountRankingItems = ref<AccountSpendingRankingItem[]>([])

let snapshotLoadSeq = 0
let userTrendLoadSeq = 0
let rankingLoadSeq = 0
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
const activeUsageMetric = ref<UsageMetricKey>('tokens')
const hiddenUsageSeries = ref<Set<string>>(new Set())

const timeRangeOptions = computed(() => [
  { key: 'today' as const, label: t('admin.dashboard.rangeToday') },
  { key: 'yesterday' as const, label: t('admin.dashboard.rangeYesterday') },
  { key: 'this_week' as const, label: t('admin.dashboard.rangeThisWeek') },
  { key: 'last_7_days' as const, label: t('admin.dashboard.rangeLast7Days') },
  { key: 'last_30_days' as const, label: t('admin.dashboard.rangeLast30Days') },
  { key: 'this_month' as const, label: t('admin.dashboard.rangeThisMonth') }
])

const adminMetricCards = computed<DashboardMetricCard[]>(() => {
  const s = stats.value
  if (!s) return []

  return [
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

const usageSeriesPalette = [
  'var(--chart-1)',
  'var(--chart-2)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
  'hsl(15, 85%, 60%)',
  'hsl(195, 85%, 60%)',
  'hsl(285, 85%, 60%)',
  'hsl(135, 85%, 50%)',
  'hsl(45, 85%, 55%)',
  'hsl(315, 85%, 65%)',
  'hsl(165, 85%, 55%)',
  'hsl(35, 85%, 65%)',
  'hsl(255, 85%, 65%)',
  'hsl(75, 85%, 50%)',
  'hsl(345, 85%, 65%)',
  'hsl(105, 85%, 55%)',
  'hsl(225, 85%, 65%)',
  'hsl(55, 85%, 60%)',
  'hsl(275, 85%, 60%)',
  'hsl(25, 85%, 65%)',
  'hsl(185, 85%, 60%)',
  'hsl(125, 85%, 55%)',
  'hsl(295, 85%, 70%)'
] as const

const getUsageSeriesColor = (index: number): string => usageSeriesPalette[index % usageSeriesPalette.length]

const getCssVariableName = (color: string): string | null => {
  return color.match(/^var\((--[^),\s]+)\)$/)?.[1] || null
}

const resolveCssColor = (color: string): string => {
  const cssVariableName = getCssVariableName(color)
  if (!cssVariableName || typeof window === 'undefined') return color

  const resolved = window.getComputedStyle(document.documentElement).getPropertyValue(cssVariableName).trim()
  return resolved || color
}

const withColorAlpha = (color: string, alpha: number): string => {
  const normalized = color.trim()
  const safeAlpha = Math.max(0, Math.min(1, alpha))
  const alphaHex = Math.round(safeAlpha * 255).toString(16).padStart(2, '0')

  if (/^#[\da-f]{3}$/i.test(normalized)) {
    const [, r, g, b] = normalized
    return `#${r}${r}${g}${g}${b}${b}${alphaHex}`
  }

  if (/^#[\da-f]{6}$/i.test(normalized)) {
    return `${normalized}${alphaHex}`
  }

  if (/^rgb\(/i.test(normalized)) {
    return normalized.replace(/^rgb\((.*)\)$/i, `rgba($1, ${safeAlpha})`)
  }

  if (/^rgba\(/i.test(normalized)) return normalized

  if (/^hsl\(/i.test(normalized)) {
    const inner = normalized.slice(4, -1)
    if (inner.includes('/')) return normalized
    if (inner.includes(',')) return `hsla(${inner}, ${safeAlpha})`
    return `hsl(${inner} / ${safeAlpha})`
  }

  if (/^oklch\(/i.test(normalized)) {
    const inner = normalized.slice(6, -1)
    if (inner.includes('/')) return normalized
    return `oklch(${inner} / ${safeAlpha})`
  }

  return normalized
}

type UsageBucket = {
  key: string
  label: string
}

type UsageSeries = {
  id: number
  label: string
  color: string
  order: number
  values: number[]
  total: number
}

const cloneDate = (date: Date): Date => new Date(date.getTime())

const parseDateValue = (value: string): Date | null => {
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const date = new Date(normalized)
  return Number.isNaN(date.getTime()) ? null : date
}

const toBucketDate = (value: string, bucketGranularity: 'day' | 'hour'): Date | null => {
  const parsed = parseDateValue(value)
  if (parsed) {
    const next = cloneDate(parsed)
    next.setMinutes(0, 0, 0)
    if (bucketGranularity === 'day') {
      next.setHours(0, 0, 0, 0)
    }
    return next
  }

  const dayMatch = value.match(/^(\d{4})-(\d{2})-(\d{2})$/)
  if (dayMatch) {
    const [, year, month, day] = dayMatch
    return new Date(Number(year), Number(month) - 1, Number(day), 0, 0, 0, 0)
  }

  const hourMatch = value.match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2})/)
  if (hourMatch) {
    const [, year, month, day, hour] = hourMatch
    return new Date(Number(year), Number(month) - 1, Number(day), Number(hour), 0, 0, 0)
  }

  return null
}

const normalizeBucketKey = (date: Date, bucketGranularity: 'day' | 'hour'): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  if (bucketGranularity === 'hour') {
    return `${year}-${month}-${day} ${String(date.getHours()).padStart(2, '0')}:00`
  }
  return `${year}-${month}-${day}`
}

const formatUsageBucketLabel = (date: Date, bucketGranularity: 'day' | 'hour'): string => {
  if (bucketGranularity === 'hour') {
    return `${String(date.getHours()).padStart(2, '0')}:00`
  }
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const buildUsageBuckets = (
  range: DashboardTimeRange,
  start: string,
  end: string,
  bucketGranularity: 'day' | 'hour'
): UsageBucket[] => {
  const startAt = parseDateValue(`${start}T00:00:00`)
  const endAt = parseDateValue(`${end}T00:00:00`)
  if (!startAt || !endAt) return []

  if (bucketGranularity === 'hour' && (range === 'today' || range === 'yesterday')) {
    const base = cloneDate(startAt)
    base.setHours(0, 0, 0, 0)
    return Array.from({ length: 24 }, (_, hour) => {
      const item = cloneDate(base)
      item.setHours(hour, 0, 0, 0)
      return {
        key: normalizeBucketKey(item, 'hour'),
        label: formatUsageBucketLabel(item, 'hour')
      }
    })
  }

  const cursor = cloneDate(startAt)
  cursor.setHours(0, 0, 0, 0)
  const limit = cloneDate(endAt)
  limit.setHours(0, 0, 0, 0)

  const buckets: UsageBucket[] = []
  while (cursor.getTime() <= limit.getTime()) {
    buckets.push({
      key: normalizeBucketKey(cursor, 'day'),
      label: formatUsageBucketLabel(cursor, 'day')
    })
    cursor.setDate(cursor.getDate() + 1)
  }

  return buckets
}

const resolveUsageSeriesLabel = (item: UserUsageTrendPoint): string => {
  const username = item.username?.trim()
  if (username) return username
  const email = item.email?.trim()
  if (email) return email
  return t('admin.dashboard.unknownUser', { id: item.user_id })
}

const usageBuckets = computed(() => (
  buildUsageBuckets(selectedTimeRange.value, startDate.value, endDate.value, granularity.value)
))

const usageSeriesColorMap = computed(() => {
  const colorMap = new Map<number, { color: string; order: number }>()

  userTrend.value.forEach((item) => {
    if (colorMap.has(item.user_id)) return

    const order = colorMap.size
    colorMap.set(item.user_id, {
      color: getUsageSeriesColor(order),
      order
    })
  })

  return colorMap
})

const usageSeries = computed<UsageSeries[]>(() => {
  if (!usageBuckets.value.length || !userTrend.value.length) return []

  const bucketIndex = new Map(usageBuckets.value.map((bucket, index) => [bucket.key, index]))
  const seriesMap = new Map<number, UsageSeries>()

  userTrend.value.forEach((item) => {
    const bucketDate = toBucketDate(item.date, granularity.value)
    if (!bucketDate) return

    const index = bucketIndex.get(normalizeBucketKey(bucketDate, granularity.value))
    if (index === undefined) return

    let series = seriesMap.get(item.user_id)
    if (!series) {
      const userColor = usageSeriesColorMap.value.get(item.user_id) || {
        color: getUsageSeriesColor(seriesMap.size),
        order: seriesMap.size
      }

      series = {
        id: item.user_id,
        label: resolveUsageSeriesLabel(item),
        color: userColor.color,
        order: userColor.order,
        values: Array.from({ length: usageBuckets.value.length }, () => 0),
        total: 0
      }
      seriesMap.set(item.user_id, series)
    }

    const metricValue = activeUsageMetric.value === 'requests'
      ? item.requests || 0
      : activeUsageMetric.value === 'tokens'
        ? item.tokens || 0
        : item.actual_cost || 0

    series.values[index] += metricValue
    series.total += metricValue
  })

  return Array.from(seriesMap.values())
    .sort((a, b) => b.total - a.total || a.order - b.order)
    .slice(0, rankingLimit)
})

const userDisplayNameMap = computed(() => {
  const displayNames = new Map<number, string>()
  userTrend.value.forEach((item) => {
    const username = item.username?.trim()
    if (username) {
      displayNames.set(item.user_id, username)
    }
  })
  return displayNames
})

const usageTotals = computed(() => {
  return userTrend.value.reduce(
    (acc, item) => {
      acc.cost += item.actual_cost || 0
      acc.requests += item.requests || 0
      acc.tokens += item.tokens || 0
      return acc
    },
    { cost: 0, requests: 0, tokens: 0 }
  )
})

const usageMetricOptions = computed(() => [
  {
    key: 'requests' as const,
    label: t('admin.dashboard.metricApiCalls'),
    value: formatNumber(usageTotals.value.requests)
  },
  {
    key: 'cost' as const,
    label: t('admin.dashboard.metricAmount'),
    value: `$${formatCost(usageTotals.value.cost)}`
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

const formatUsageMetricValue = (value: number): string => {
  if (activeUsageMetric.value === 'cost') return `$${formatCost(value)}`
  if (activeUsageMetric.value === 'tokens') return formatTokens(value)
  return formatNumber(value)
}

const usageLegendItems = computed<DashboardUsageLegendItem[]>(() => {
  return usageSeries.value.map((series) => ({
    id: `usage-series-${series.id}`,
    label: series.label,
    value: formatUsageMetricValue(series.total),
    color: series.color,
    disabled: hiddenUsageSeries.value.has(`usage-series-${series.id}`)
  }))
})

const usageChartData = computed<ChartData<'line', number[], string> | null>(() => {
  if (!usageBuckets.value.length || !usageSeries.value.length) return null
  const visibleSeries = usageSeries.value.filter((series) => !hiddenUsageSeries.value.has(`usage-series-${series.id}`))
  if (!visibleSeries.length) return null

  return {
    labels: usageBuckets.value.map((bucket) => bucket.label),
    datasets: visibleSeries.map((series) => {
      const lineColor = resolveCssColor(series.color)

      return {
        label: series.label,
        data: series.values,
        borderColor: lineColor,
        backgroundColor: withColorAlpha(lineColor, 0.14),
        fill: true,
        tension: 0.42,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHitRadius: 16,
        pointBackgroundColor: lineColor,
        pointBorderColor: isDarkMode.value ? '#0b0f1a' : '#ffffff',
        pointBorderWidth: 1.5,
        borderWidth: 2.5
      }
    })
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
            return `${context.dataset.label}: $${formatCost(value)}`
          }
          if (activeUsageMetric.value === 'tokens') {
            return `${context.dataset.label}: ${formatTokens(value)}`
          }
          return `${context.dataset.label}: ${formatNumber(value)}`
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
          return formatNumber(numeric)
        }
      }
    }
  }
}))

const resolveRankingUserName = (item: UserSpendingRankingItem): string => {
  return userDisplayNameMap.value.get(item.user_id)
    || item.username?.trim()
    || item.email?.trim()
    || t('admin.dashboard.unknownUser', { id: item.user_id })
}

const userRankingEntries = computed<DashboardRankingEntry[]>(() => {
  return rankingItems.value.slice(0, 3).map((item) => ({
    id: `user-${item.user_id}`,
    rawId: item.user_id,
    name: resolveRankingUserName(item),
    requests: item.requests || 0,
    tokens: item.tokens || 0,
    cost: item.actual_cost || 0
  }))
})

const accountRankingEntries = computed<DashboardRankingEntry[]>(() => {
  return accountRankingItems.value.slice(0, 3).map((item) => ({
    id: `account-${item.account_id}`,
    rawId: item.account_id,
    name: item.account_name?.trim() || `${item.platform || t('usage.unknown')} #${item.account_id}`,
    requests: item.requests || 0,
    tokens: item.tokens || 0,
    cost: item.actual_cost || 0
  }))
})

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

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}

const applyTimeRange = (range: DashboardTimeRange) => {
  if (selectedTimeRange.value === range && userTrend.value.length > 0) return
  const next = resolveTimeRange(range)
  selectedTimeRange.value = range
  startDate.value = next.start
  endDate.value = next.end
  granularity.value = next.granularity
  void loadChartData()
}

const toggleUsageSeries = (seriesId: string) => {
  const next = new Set(hiddenUsageSeries.value)
  if (next.has(seriesId)) {
    next.delete(seriesId)
  } else {
    next.add(seriesId)
  }
  hiddenUsageSeries.value = next
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
  const currentSeq = ++snapshotLoadSeq
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
    if (currentSeq !== snapshotLoadSeq) return
    if (includeStats && response.stats) {
      stats.value = response.stats
    }
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== snapshotLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === snapshotLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadUserTrend = async () => {
  const currentSeq = ++userTrendLoadSeq
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getUserUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      limit: rankingLimit
    })
    if (currentSeq !== userTrendLoadSeq) return
    userTrend.value = response.trend || []
  } catch (error) {
    if (currentSeq !== userTrendLoadSeq) return
    console.error('Error loading user trend:', error)
    userTrend.value = []
  } finally {
    if (currentSeq === userTrendLoadSeq) {
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

const loadAccountRanking = async () => {
  const currentSeq = ++accountRankingLoadSeq
  accountRankingLoading.value = true
  try {
    const response = await adminAPI.dashboard.getAccountSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit
    })
    if (currentSeq !== accountRankingLoadSeq) return
    accountRankingItems.value = response.ranking || []
  } catch (error) {
    if (currentSeq !== accountRankingLoadSeq) return
    console.error('Error loading account ranking:', error)
    accountRankingItems.value = []
  } finally {
    if (currentSeq === accountRankingLoadSeq) {
      accountRankingLoading.value = false
    }
  }
}

const loadDashboardStats = async () => {
  await Promise.all([
    loadDashboardSnapshot(true),
    loadUserTrend(),
    loadUserSpendingRanking(),
    loadAccountRanking()
  ])
}

const loadChartData = async () => {
  await Promise.all([
    loadDashboardSnapshot(false),
    loadUserTrend(),
    loadUserSpendingRanking(),
    loadAccountRanking()
  ])
}

onMounted(() => {
  void loadDashboardStats()
})
</script>
