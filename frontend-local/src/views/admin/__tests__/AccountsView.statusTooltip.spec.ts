import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import AccountsView from '../AccountsView.vue'
import type { Account } from '@/types'

const {
  listAccounts,
  listWithEtag,
  getBatchTodayStats,
  getAllProxies,
  getAllGroups
} = vi.hoisted(() => ({
  listAccounts: vi.fn(),
  listWithEtag: vi.fn(),
  getBatchTodayStats: vi.fn(),
  getAllProxies: vi.fn(),
  getAllGroups: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      list: listAccounts,
      listWithEtag,
      getBatchTodayStats,
      delete: vi.fn(),
      getAvailableModels: vi.fn(),
      refreshCredentials: vi.fn(),
      recoverState: vi.fn(),
      clearError: vi.fn(),
      resetAccountQuota: vi.fn(),
      setPrivacy: vi.fn(),
      setSchedulable: vi.fn(),
      exportData: vi.fn()
    },
    proxies: {
      getAll: getAllProxies
    },
    groups: {
      getAll: getAllGroups
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn()
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    token: 'test-token',
    isSimpleMode: false
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (params && typeof params.count !== 'undefined') return `${key}:${params.count}`
        return key
      }
    })
  }
})

function createErrorAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'OpenAI Account',
    notes: null,
    platform: 'openai',
    type: 'oauth',
    credentials: { email: 'user@example.com' },
    extra: {},
    proxy_id: null,
    concurrency: 5,
    priority: 1,
    rate_multiplier: 1,
    status: 'error',
    error_message: 'Authentication failed (401): Your authentication token has been invalidated. Please try logging in again.',
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: true,
    created_at: '2026-05-28T00:00:00Z',
    updated_at: '2026-05-28T00:00:00Z',
    group_ids: [],
    groups: [],
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  }
}

function mountAccountsView() {
  return mount(AccountsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
        },
        ConfirmDialog: true,
        AccountTableFilters: { template: '<div></div>' },
        AccountActionMenu: true,
        ImportDataModal: true,
        ReAuthAccountModal: true,
        AccountTestModal: true,
        AccountStatsModal: true,
        ScheduledTestsPanel: true,
        SyncFromCrsModal: true,
        TempUnschedStatusModal: true,
        ErrorPassthroughRulesModal: true,
        TLSFingerprintProfilesModal: true,
        CreateAccountModal: true,
        EditAccountModal: true,
        PlatformTypeBadge: true,
        AccountGroupsCell: true,
        AccountUsageCell: true,
        Icon: true
      }
    }
  })
}

describe('admin AccountsView status tooltip', () => {
  beforeEach(() => {
    localStorage.clear()
    document.body.innerHTML = ''
    vi.stubGlobal('innerWidth', 320)
    vi.stubGlobal('innerHeight', 640)

    listAccounts.mockReset()
    listWithEtag.mockReset()
    getBatchTodayStats.mockReset()
    getAllProxies.mockReset()
    getAllGroups.mockReset()

    listAccounts.mockResolvedValue({
      items: [createErrorAccount()],
      total: 1,
      page: 1,
      page_size: 1000,
      pages: 1
    })
    listWithEtag.mockResolvedValue({
      notModified: true,
      etag: null,
      data: null
    })
    getBatchTodayStats.mockResolvedValue({ stats: {} })
    getAllProxies.mockResolvedValue([])
    getAllGroups.mockResolvedValue([])
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('renders long error detail in a body-level tooltip that stays inside the left viewport edge', async () => {
    const wrapper = mountAccountsView()
    await flushPromises()

    const chip = wrapper.get('.account-status-chip')
    vi.spyOn(chip.element, 'getBoundingClientRect').mockReturnValue({
      x: 2,
      y: 160,
      left: 2,
      right: 54,
      top: 160,
      bottom: 184,
      width: 52,
      height: 24,
      toJSON: () => ({})
    } as DOMRect)

    await chip.trigger('mouseenter')
    await nextTick()
    await nextTick()

    const tooltip = document.body.querySelector('[role="tooltip"]') as HTMLElement | null

    expect(tooltip).not.toBeNull()
    expect(tooltip?.parentElement).toBe(document.body)
    expect(tooltip?.textContent).toContain('authentication token has been invalidated')
    expect(Number.parseFloat(tooltip?.style.left || '0')).toBeGreaterThanOrEqual(12)
  })
})
