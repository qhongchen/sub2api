import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import type {
  AdminUsageLog,
  DashboardStats,
  UserUsageTrendPoint,
  UserSpendingRankingResponse
} from '@/types'
import DashboardRankingGrid from '@/components/admin/dashboard/DashboardRankingGrid.vue'
import DashboardUsageChart from '@/components/admin/dashboard/DashboardUsageChart.vue'
import DashboardView from '../DashboardView.vue'

const push = vi.fn()

vi.mock('vue-chartjs', () => ({
  Line: {
    name: 'Line',
    template: '<div class="line-stub" />'
  }
}))

const {
  getSnapshotV2,
  getUserUsageTrend,
  getUserSpendingRanking,
  listAccounts,
  listUsage
} = vi.hoisted(() => ({
  getSnapshotV2: vi.fn(),
  getUserUsageTrend: vi.fn(),
  getUserSpendingRanking: vi.fn(),
  listAccounts: vi.fn(),
  listUsage: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    dashboard: {
      getSnapshotV2,
      getUserUsageTrend,
      getUserSpendingRanking
    },
    accounts: {
      list: listAccounts
    },
    usage: {
      list: listUsage
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn()
  })
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'admin.dashboard.unknownUser') {
          return `用户 #${params?.id ?? ''}`
        }
        return key
      }
    })
  }
})

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const createDashboardStats = (): DashboardStats => ({
  total_users: 0,
  today_new_users: 0,
  active_users: 0,
  hourly_active_users: 0,
  stats_updated_at: '',
  stats_stale: false,
  total_api_keys: 0,
  active_api_keys: 0,
  total_accounts: 0,
  normal_accounts: 0,
  error_accounts: 0,
  ratelimit_accounts: 0,
  overload_accounts: 0,
  total_requests: 0,
  total_input_tokens: 0,
  total_output_tokens: 0,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 0,
  total_tokens: 0,
  total_cost: 0,
  total_actual_cost: 0,
  today_requests: 0,
  today_input_tokens: 0,
  today_output_tokens: 0,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 0,
  today_tokens: 0,
  today_cost: 0,
  today_actual_cost: 0,
  today_account_cost: 0,
  average_duration_ms: 0,
  uptime: 0,
  rpm: 0,
  tpm: 0
})

const createTrendPoint = (overrides: Partial<UserUsageTrendPoint>): UserUsageTrendPoint => ({
  date: '2026-05-26 09:00',
  user_id: 1,
  email: 'alice@example.com',
  username: 'Alice',
  requests: 1,
  tokens: 100,
  cost: 0.1234,
  actual_cost: 0.1234,
  ...overrides
})

const createRankingResponse = (): UserSpendingRankingResponse => ({
  ranking: [],
  total_actual_cost: 0,
  total_requests: 0,
  total_tokens: 0,
  start_date: '',
  end_date: ''
})

const mountDashboard = async () => {
  const wrapper = mount(DashboardView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        LoadingSpinner: true,
        Icon: true
      }
    }
  })

  await flushPromises()
  return wrapper
}

describe('admin DashboardView', () => {
  beforeEach(() => {
    push.mockReset()
    getSnapshotV2.mockReset()
    getUserUsageTrend.mockReset()
    getUserSpendingRanking.mockReset()
    listAccounts.mockReset()
    listUsage.mockReset()

    getSnapshotV2.mockResolvedValue({
      stats: createDashboardStats(),
      trend: [],
      models: []
    })
    getUserUsageTrend.mockResolvedValue({
      trend: [],
      start_date: '',
      end_date: '',
      granularity: 'hour'
    })
    getUserSpendingRanking.mockResolvedValue(createRankingResponse())
    listAccounts.mockResolvedValue({ items: [] })
    listUsage.mockResolvedValue({ items: [] as AdminUsageLog[] })
  })

  it('默认使用 today 范围并请求用户趋势', async () => {
    await mountDashboard()

    const now = new Date()

    expect(getSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(now),
      end_date: formatLocalDate(now),
      granularity: 'hour'
    }))
    expect(getUserUsageTrend).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(now),
      end_date: formatLocalDate(now),
      granularity: 'hour',
      limit: 12
    }))
  })

  it('today 范围补齐 24 个点，并按用户拆成多条线', async () => {
    getUserUsageTrend.mockResolvedValueOnce({
      trend: [
        createTrendPoint({
          user_id: 1,
          username: 'Alice',
          email: 'alice@example.com',
          date: '2026-05-26 01:15',
          actual_cost: 1.2,
          requests: 3,
          tokens: 120
        }),
        createTrendPoint({
          user_id: 2,
          username: 'Bob',
          email: 'bob@example.com',
          date: '2026-05-26 03:45',
          actual_cost: 2.5,
          requests: 5,
          tokens: 240
        })
      ],
      start_date: '2026-05-26',
      end_date: '2026-05-26',
      granularity: 'hour'
    })

    const wrapper = await mountDashboard()
    const chart = wrapper.findComponent(DashboardUsageChart)
    const chartData = chart.props('chartData') as {
      labels: string[]
      datasets: Array<{ label: string; data: number[]; borderColor: string }>
    }
    const legendItems = chart.props('legendItems') as Array<{ label: string; value: string; color: string; disabled?: boolean }>
    const metricOptions = chart.props('metricOptions') as Array<{ key: string; value: string }>

    expect(chartData.labels).toHaveLength(24)
    expect(chartData.labels[0]).toBe('00:00')
    expect(chartData.labels[23]).toBe('23:00')
    expect(chartData.datasets).toHaveLength(2)
    expect(chartData.datasets[0].label).toBe('Bob')
    expect(chartData.datasets[0].data).toHaveLength(24)
    expect(chartData.datasets[1].data).toHaveLength(24)
    expect(chartData.datasets[1].data[1]).toBe(120)
    expect(chartData.datasets[1].data[0]).toBe(0)
    expect(chartData.datasets[0].data[3]).toBe(240)
    expect(chartData.datasets[0].data[2]).toBe(0)
    expect(chartData.datasets[0].borderColor).toBe('var(--chart-2)')
    expect(chartData.datasets[1].borderColor).toBe('var(--chart-1)')
    expect(chartData.datasets[0]).toEqual(expect.objectContaining({
      fill: true,
      pointRadius: 0,
      pointHoverRadius: 4
    }))
    expect(legendItems[0]).toEqual(expect.objectContaining({ label: 'Bob', value: '240', color: 'var(--chart-2)', disabled: false }))
    expect(legendItems[1]).toEqual(expect.objectContaining({ label: 'Alice', value: '120', color: 'var(--chart-1)', disabled: false }))
    expect(metricOptions.map((item) => item.key)).toEqual(['requests', 'cost', 'tokens'])
    expect(metricOptions.find((item) => item.key === 'cost')?.value).toBe('$3.70')
    expect(chart.props('activeMetric')).toBe('tokens')
  })

  it('7天范围返回 7 个点，并为缺失日期补 0', async () => {
    const now = new Date()
    const rangeStart = new Date(now)
    rangeStart.setDate(now.getDate() - 6)
    const middle = new Date(rangeStart)
    middle.setDate(rangeStart.getDate() + 2)

    getUserUsageTrend
      .mockResolvedValueOnce({
        trend: [],
        start_date: formatLocalDate(now),
        end_date: formatLocalDate(now),
        granularity: 'hour'
      })
      .mockResolvedValueOnce({
        trend: [
          createTrendPoint({
            date: `${formatLocalDate(middle)} 12:00`,
            actual_cost: 7,
            requests: 7,
            tokens: 7
          })
        ],
        start_date: formatLocalDate(rangeStart),
        end_date: formatLocalDate(now),
        granularity: 'day'
      })

    const wrapper = await mountDashboard()
    const chart = wrapper.findComponent(DashboardUsageChart)
    const buttons = chart.findAll('button')
    const last7DaysButton = buttons.find((button) => button.text() === 'admin.dashboard.rangeLast7Days')

    expect(last7DaysButton).toBeDefined()
    await last7DaysButton!.trigger('click')
    await flushPromises()

    const chartData = wrapper.findComponent(DashboardUsageChart).props('chartData') as {
      labels: string[]
      datasets: Array<{ data: number[] }>
    }

    expect(chartData.labels).toHaveLength(7)
    expect(chartData.datasets[0].data).toHaveLength(7)
    expect(chartData.datasets[0].data.filter((value) => value === 0)).toHaveLength(6)
    expect(chartData.datasets[0].data.some((value) => value === 7)).toBe(true)
    expect(getUserUsageTrend).toHaveBeenLastCalledWith(expect.objectContaining({
      start_date: formatLocalDate(rangeStart),
      end_date: formatLocalDate(now),
      granularity: 'day',
      limit: 12
    }))
  })

  it('金额显示保留两位小数，0 显示为 0.00', async () => {
    getUserUsageTrend.mockResolvedValueOnce({
      trend: [
        createTrendPoint({
          user_id: 3,
          username: '',
          email: '',
          actual_cost: 0,
          requests: 2,
          tokens: 0
        })
      ],
      start_date: '2026-05-26',
      end_date: '2026-05-26',
      granularity: 'hour'
    })

    const wrapper = await mountDashboard()
    let chart = wrapper.findComponent(DashboardUsageChart)
    const legendItems = chart.props('legendItems') as Array<{ label: string; value: string }>
    const metricOptions = chart.props('metricOptions') as Array<{ key: string; value: string }>

    expect(metricOptions.find((item) => item.key === 'cost')?.value).toBe('$0.00')
    expect(legendItems[0]?.value).toBe('0')

    await chart.vm.$emit('update:activeMetric', 'cost')
    await flushPromises()

    chart = wrapper.findComponent(DashboardUsageChart)
    const costLegendItems = chart.props('legendItems') as Array<{ label: string; value: string }>

    expect(legendItems[0]?.label).toBe('用户 #3')
    expect(costLegendItems[0]?.value).toBe('$0.00')
  })

  it('点击图例可以隐藏或恢复对应曲线', async () => {
    getUserUsageTrend.mockResolvedValueOnce({
      trend: [
        createTrendPoint({
          user_id: 1,
          username: 'Alice',
          email: 'alice@example.com',
          date: '2026-05-26 01:00',
          tokens: 100
        }),
        createTrendPoint({
          user_id: 2,
          username: 'Bob',
          email: 'bob@example.com',
          date: '2026-05-26 02:00',
          tokens: 300
        })
      ],
      start_date: '2026-05-26',
      end_date: '2026-05-26',
      granularity: 'hour'
    })

    const wrapper = await mountDashboard()
    let chart = wrapper.findComponent(DashboardUsageChart)

    expect((chart.props('chartData') as { datasets: Array<{ label: string }> }).datasets.map((item) => item.label)).toEqual(['Bob', 'Alice'])

    await chart.vm.$emit('toggleSeries', 'usage-series-2')
    await flushPromises()

    chart = wrapper.findComponent(DashboardUsageChart)
    expect((chart.props('chartData') as { datasets: Array<{ label: string }> }).datasets.map((item) => item.label)).toEqual(['Alice'])
    expect((chart.props('legendItems') as Array<{ label: string; disabled?: boolean }>)[0]).toEqual(expect.objectContaining({
      label: 'Bob',
      disabled: true
    }))

    await chart.vm.$emit('toggleSeries', 'usage-series-2')
    await flushPromises()

    chart = wrapper.findComponent(DashboardUsageChart)
    expect((chart.props('chartData') as { datasets: Array<{ label: string }> }).datasets.map((item) => item.label)).toEqual(['Bob', 'Alice'])
  })

  it('用户排行优先显示用户名，缺少用户名时回退邮箱', async () => {
    getUserUsageTrend.mockResolvedValueOnce({
      trend: [
        createTrendPoint({
          user_id: 1,
          username: 'shuiguai',
          email: 'shuiguai@sub2api.local',
          tokens: 100
        }),
        createTrendPoint({
          user_id: 2,
          username: '',
          email: 'xiaohezi@sub2api.local',
          tokens: 80
        })
      ],
      start_date: '2026-05-26',
      end_date: '2026-05-26',
      granularity: 'hour'
    })
    getUserSpendingRanking.mockResolvedValueOnce({
      ranking: [
        {
          user_id: 1,
          email: 'shuiguai@sub2api.local',
          username: 'shuiguai-from-ranking',
          actual_cost: 41.13,
          requests: 243,
          tokens: 28480000
        },
        {
          user_id: 2,
          email: 'xiaohezi@sub2api.local',
          username: 'xiaohezi-from-ranking',
          actual_cost: 4.77,
          requests: 68,
          tokens: 8620000
        }
      ],
      total_actual_cost: 45.9,
      total_requests: 311,
      total_tokens: 37100000,
      start_date: '2026-05-26',
      end_date: '2026-05-26'
    })

    const wrapper = await mountDashboard()
    const rankingGrid = wrapper.findComponent(DashboardRankingGrid)
    const userEntries = rankingGrid.props('userEntries') as Array<{ name: string }>

    expect(userEntries.map((item) => item.name)).toEqual([
      'shuiguai',
      'xiaohezi-from-ranking'
    ])
  })
})
