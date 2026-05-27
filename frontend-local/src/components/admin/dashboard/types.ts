export type DashboardMetricIcon =
  | 'key'
  | 'server'
  | 'chart'
  | 'userPlus'
  | 'cube'
  | 'database'
  | 'bolt'
  | 'clock'
  | 'terminal'

export type UsageMetricKey = 'cost' | 'requests' | 'tokens'

export type DashboardTimeRange =
  | 'today'
  | 'yesterday'
  | 'this_week'
  | 'last_7_days'
  | 'last_30_days'
  | 'this_month'

export interface DashboardMetricComparison {
  label: string
  value: string
  valueClass: string
}

export interface DashboardMetricCard {
  key: string
  title: string
  value: string
  subtitle?: string
  badge?: string
  icon: DashboardMetricIcon
  accentClass: string
  iconClass: string
  iconTextClass: string
  valueClass?: string
  comparisons: DashboardMetricComparison[]
}

export interface DashboardTimeRangeOption {
  key: DashboardTimeRange
  label: string
}

export interface UsageMetricOption {
  key: UsageMetricKey
  label: string
  value: string
}

export interface ActiveUsageMetricConfig extends UsageMetricOption {
  color: string
}

export interface DashboardUsageLegendItem {
  id: string
  label: string
  value: string
  color: string
  disabled?: boolean
}

export interface DashboardRankingEntry {
  id: string
  rawId: number
  name: string
  requests: number
  tokens: number
  cost: number
}

export interface DashboardActiveSessionEntry {
  id: string
  shortId: string
  name: string
  activeSessions: number
}
