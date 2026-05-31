import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RequestLogsView from '../RequestLogsView.vue'

const {
  requestRecordsList,
  groupsList,
  accountsList,
  searchUsers,
  searchApiKeys,
  settingsFetch,
  setOpsMonitoringEnabledLocal,
  showError,
  opsState,
} = vi.hoisted(() => {
  vi.stubGlobal('localStorage', {
    getItem: vi.fn(() => null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
  })

  return {
    requestRecordsList: vi.fn(),
    groupsList: vi.fn(),
    accountsList: vi.fn(),
    searchUsers: vi.fn(),
    searchApiKeys: vi.fn(),
    settingsFetch: vi.fn(),
    setOpsMonitoringEnabledLocal: vi.fn(),
    showError: vi.fn(),
    opsState: { enabled: true },
  }
})

const messages: Record<string, string> = {
  'admin.ops.timeRange.5m': '5m',
  'admin.ops.timeRange.30m': '30m',
  'admin.ops.timeRange.1h': '1h',
  'admin.ops.timeRange.6h': '6h',
  'admin.ops.timeRange.24h': '24h',
  'admin.ops.timeRange.custom': 'Custom',
  'admin.requestLogs.title': 'Request Logs',
  'admin.requestLogs.description': 'Request-level diagnostics',
  'admin.requestLogs.refresh': 'Refresh',
  'admin.requestLogs.disabledTitle': 'Ops disabled',
  'admin.requestLogs.disabledDescription': 'Enable ops monitoring first',
  'admin.requestLogs.openSettings': 'Open Settings',
  'admin.requestLogs.filters': 'Filters',
  'admin.requestLogs.reset': 'Reset',
  'admin.requestLogs.apply': 'Apply',
  'admin.requestLogs.tableTitle': 'Request Details',
  'admin.requestLogs.columnSettings': 'Column Settings',
  'admin.requestLogs.resetColumns': 'Reset Columns',
  'admin.requestLogs.visibleColumnsSummary': 'Showing {visible}/{total} columns',
  'admin.requestLogs.billable': 'Billing',
  'admin.requestLogs.sessionId': 'Session',
  'admin.requestLogs.totalRequests': 'Total Requests',
  'admin.requestLogs.successRequests': 'Success Requests',
  'admin.requestLogs.errorRequests': 'Error Requests',
  'admin.requestLogs.errorRate': 'Error Rate',
  'admin.requestLogs.kindSuccess': 'Success',
  'admin.requestLogs.kindError': 'Error',
  'admin.requestLogs.noValue': '-',
  'admin.requestLogs.copyRequestId': 'Copy Request ID',
  'admin.requestLogs.requestIdCopied': 'Copied',
  'admin.requestLogs.viewError': 'View Error',
  'admin.requestLogs.loading': 'Loading',
  'admin.requestLogs.empty': 'Empty',
  'admin.requestLogs.emptyHint': 'Try a wider range',
  'admin.requestLogs.loadFailed': 'Load failed',
  'admin.requestLogs.requestTypes.sync': 'Sync',
  'admin.requestLogs.requestTypes.stream': 'Stream',
  'admin.requestLogs.requestTypes.ws_v2': 'WebSocket',
  'admin.requestLogs.requestTypes.unknown': 'Unknown',
  'admin.requestLogs.billableOptions.all': 'All',
  'admin.requestLogs.billableOptions.true': 'Billable',
  'admin.requestLogs.billableOptions.false': 'Non-billable',
  'admin.requestLogs.sessionIdPlaceholder': 'Enter Session ID',
  'admin.requestLogs.table.time': 'Time',
  'admin.requestLogs.table.user': 'User',
  'admin.requestLogs.table.apiKey': 'API Key',
  'admin.requestLogs.table.requestId': 'Request ID',
  'admin.requestLogs.table.sourceIp': 'Source IP',
  'admin.requestLogs.table.platform': 'Platform',
  'admin.requestLogs.table.model': 'Model',
  'admin.requestLogs.table.billingModel': 'Billing Model',
  'admin.requestLogs.table.tokens': 'Token',
  'admin.requestLogs.table.cache': 'Cache',
  'admin.requestLogs.table.cost': 'Cost',
  'admin.requestLogs.table.performance': 'Performance',
  'admin.requestLogs.table.status': 'Status',
  'admin.requestLogs.table.account': 'Account',
  'admin.requestLogs.table.session': 'Session',
  'admin.requestLogs.table.billable': 'Billing',
  'admin.requestLogs.table.group': 'Group',
  'admin.requestLogs.table.userAgent': 'User-Agent',
  'admin.requestLogs.table.inboundEndpoint': 'Inbound',
  'admin.requestLogs.table.errorSummary': 'Error Summary',
  'admin.requestLogs.table.actions': 'Actions',
  'admin.usage.inputTokens': 'Input Tokens',
  'admin.usage.outputTokens': 'Output Tokens',
  'admin.usage.cacheCreationTokens': 'Cache Creation Tokens',
  'admin.usage.cacheReadTokens': 'Cache Read Tokens',
  'admin.usage.inputCost': 'Input Cost',
  'admin.usage.outputCost': 'Output Cost',
  'admin.usage.cacheCreationCost': 'Cache Creation Cost',
  'admin.usage.cacheReadCost': 'Cache Read Cost',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per Request',
  'admin.usage.billingModeImage': 'Image',
  'usage.justNow': 'just now',
  'usage.secondsAgo': '{count}s ago',
  'usage.minutesAgo': '{count}m ago',
  'usage.hoursAgo': '{count}h ago',
  'usage.daysAgo': '{count}d ago',
  'usage.tokenDetails': 'Token Breakdown',
  'usage.costDetails': 'Cost Breakdown',
  'usage.totalTokens': 'Total Tokens',
  'usage.original': 'Original',
  'usage.userBilled': 'User Billed',
  'usage.accountBilled': 'Account Billed',
  'usage.accountMultiplier': 'Account rate',
  'usage.rate': 'Rate',
  'usage.serviceTier': 'Service tier',
  'usage.inputTokenPrice': 'Input price',
  'usage.outputTokenPrice': 'Output price',
  'usage.perMillionTokens': '/ 1M tokens',
  'usage.imageUnit': ' images',
  'usage.imageCount': 'Image Count',
  'usage.imageBillingSize': 'Billing Size',
  'usage.imageSizeSource': 'Size Source',
  'usage.imageInputSize': 'Input Size',
  'usage.imageOutputSize': 'Output Size',
  'usage.imageSizeBreakdown': 'Size Breakdown',
  'usage.imageSizeSourceOutput': 'Upstream Output',
  'usage.imageSizeSourceInput': 'Request Input',
  'usage.imageSizeSourceDefault': 'Default',
  'usage.imageSizeSourceLegacy': 'Legacy',
  'usage.imageSizeSourceMissing': 'Missing',
  'usage.imageSizeNotRecorded': 'Not Recorded',
  'usage.imageSizeLegacyUnstandardized': 'Legacy Size',
  'usage.imageSizeUnknown': 'Unknown',
  'usage.imageUnitPrice': 'Per Image',
  'usage.sync': 'Sync',
  'usage.stream': 'Stream',
  'usage.ws': 'WebSocket',
  'usage.unknown': 'Unknown',
  'usage.statusFailed': 'Failed',
  'usage.statusPending': 'Pending',
  'usage.statusSuccess': 'Success',
  'usage.billable': 'Billable',
  'usage.nonBillable': 'Non-billable',
}

vi.mock('@/api', () => ({
  adminAPI: {
    requestRecords: {
      list: requestRecordsList,
    },
    groups: {
      list: groupsList,
    },
    accounts: {
      list: accountsList,
    },
    usage: {
      searchUsers,
      searchApiKeys,
    },
  },
}))

vi.mock('@/stores', () => ({
  useAdminSettingsStore: () => ({
    fetch: settingsFetch,
    get opsMonitoringEnabled() {
      return opsState.enabled
    },
    setOpsMonitoringEnabledLocal,
  }),
  useAppStore: () => ({
    showError,
    showWarning: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => value,
  formatNumber: (value: number) => String(value),
  formatReasoningEffort: (value?: string | null) => value ? 'XHigh' : '-',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

const AppLayoutStub = { template: '<main><slot /></main>' }
const IconStub = { template: '<span class="icon" />' }
const PaginationStub = { template: '<nav data-test="pagination" />' }
const RouterLinkStub = { props: ['to'], template: '<a><slot /></a>' }

const errorRow = {
  kind: 'error',
  outcome: 'error',
  billable: false,
  created_at: '2026-05-29T01:30:00Z',
  request_id: 'req-error',
  session_id: 'sess-error',
  session_source: 'metadata_session_id',
  client_session_id: 'client-sess-error',
  status_code: 503,
  platform: 'openai',
  model: '5.4-mini',
  request_type: 'stream',
  stream: true,
  duration_ms: 1500,
  first_token_ms: 120,
  user_id: 42,
  user: { id: 42, email: 'user@example.com', username: 'tester' },
  api_key_id: 7,
  api_key: { id: 7, name: 'diagnostic-key' },
  account_id: 9,
  account: { id: 9, name: 'openai-account' },
  group_id: 11,
  group: { id: 11, name: 'default-group' },
  inbound_endpoint: '/v1/chat/completions',
  upstream_endpoint: 'https://upstream.example/v1/chat/completions',
  ip_address: '10.0.0.8',
  user_agent: 'codex-cli/1.0',
  error_id: 123,
  phase: 'upstream_response',
  severity: 'error',
  message: '503 service unavailable',
}

const successRow = {
  kind: 'success',
  outcome: 'success',
  billable: true,
  created_at: '2026-05-29T01:20:00Z',
  request_id: 'req-success',
  session_id: 'sess-success',
  session_source: 'header_session_id',
  client_session_id: 'client-sess-success',
  status_code: 200,
  platform: 'openai',
  model: '5.4-mini',
  request_type: 'sync',
  stream: false,
  duration_ms: 800,
  first_token_ms: null,
  user_id: 42,
  user: { id: 42, email: 'user@example.com', username: 'tester' },
  api_key_id: 7,
  api_key: { id: 7, name: 'diagnostic-key' },
  account_id: 9,
  account: { id: 9, name: 'openai-account' },
  group_id: 11,
  group: { id: 11, name: 'default-group' },
  inbound_endpoint: '/v1/chat/completions',
  upstream_endpoint: '/v1/chat/completions',
  ip_address: '10.0.0.8',
  user_agent: 'codex-cli/1.0',
  upstream_model: 'gpt-5.5',
  model_mapping_chain: '5.4-mini→gpt-5.5',
  billing_tier: 'XHigh',
  service_tier: 'priority',
  reasoning_effort: 'xhigh',
  input_tokens: 100000,
  output_tokens: 500,
  cache_creation_tokens: 60000,
  cache_read_tokens: 40000,
  cache_creation_5m_tokens: 30000,
  cache_creation_1h_tokens: 30000,
  input_cost: 0.01,
  output_cost: 0.02,
  cache_creation_cost: 0.003,
  cache_read_cost: 0.004,
  total_cost: 0.037,
  actual_cost: 0.041,
  rate_multiplier: 1.1,
  billing_type: 0,
  billing_mode: 'token',
  account_rate_multiplier: 1.2,
  account_stats_cost: 0.05,
  cache_ttl_overridden: true,
}

const pendingRow = {
  kind: 'pending',
  outcome: 'pending',
  billable: false,
  created_at: '2026-05-29T01:40:00Z',
  request_id: 'req-pending',
  session_id: 'sess-pending',
  session_source: 'header_session_id',
  client_session_id: 'client-sess-pending',
  status_code: 0,
  platform: 'openai',
  model: '5.4-mini',
  request_type: 'stream',
  stream: true,
  duration_ms: null,
  first_token_ms: null,
  user_id: 42,
  user: { id: 42, email: 'user@example.com', username: 'tester' },
  api_key_id: 7,
  api_key: { id: 7, name: 'diagnostic-key' },
  account_id: 9,
  account: { id: 9, name: 'openai-account' },
  group_id: 11,
  group: { id: 11, name: 'default-group' },
  inbound_endpoint: '/v1/responses',
  upstream_endpoint: '/v1/responses',
  ip_address: '10.0.0.8',
  user_agent: 'codex-cli/1.0',
}

const imageSuccessRow = {
  ...successRow,
  request_id: 'req-image',
  model: 'gpt-image-1',
  billing_mode: 'image',
  input_tokens: 0,
  output_tokens: 0,
  cache_creation_tokens: 0,
  cache_read_tokens: 0,
  image_output_tokens: 1200,
  image_output_cost: 0.06,
  image_count: 2,
  image_size: '2K',
  image_input_size: '1024x1024',
  image_output_size: '2048x2048',
  image_size_source: 'output',
  image_size_breakdown: { '2K': 2 },
  total_cost: 0.06,
  actual_cost: 0.072,
}

function mountRequestLogsView() {
  return mount(RequestLogsView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Icon: IconStub,
        Pagination: PaginationStub,
        RouterLink: RouterLinkStub,
      },
    },
  })
}

describe('RequestLogsView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    opsState.enabled = true
    requestRecordsList.mockReset()
    groupsList.mockReset()
    accountsList.mockReset()
    searchUsers.mockReset()
    searchApiKeys.mockReset()
    settingsFetch.mockReset()
    setOpsMonitoringEnabledLocal.mockReset()
    showError.mockReset()

    settingsFetch.mockResolvedValue(undefined)
    groupsList.mockResolvedValue({ items: [{ id: 11, name: 'default-group' }] })
    accountsList.mockResolvedValue({ items: [] })
    searchUsers.mockResolvedValue([])
    searchApiKeys.mockResolvedValue([])
    requestRecordsList.mockResolvedValue({
      items: [errorRow, successRow],
      total: 9,
      page: 1,
      page_size: 20,
      pages: 1,
      summary: {
        total_requests: 9,
        success_requests: 4,
        error_requests: 5,
        error_rate: 0.5,
      },
    })
  })

  afterEach(() => {
    document.body.innerHTML = ''
    vi.useRealTimers()
  })

  it('loads recent one-hour request logs by default and renders summary from API', async () => {
    const wrapper = mountRequestLogsView()
    await flushPromises()

    expect(requestRecordsList).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 20,
        time_range: '1h',
        sort: 'created_at_desc',
      }),
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )
    expect(wrapper.text()).toContain('Request Logs')
    expect(wrapper.text()).toContain('Total Requests')
    expect(wrapper.text()).toContain('9')
    expect(wrapper.text()).toContain('Success Requests')
    expect(wrapper.text()).toContain('4')
    expect(wrapper.text()).toContain('Error Requests')
    expect(wrapper.text()).toContain('5')
    expect(wrapper.text()).toContain('50.00%')
  })

  it('passes kind, 503 status, model, and exclude-200 filters to the API', async () => {
    const wrapper = mountRequestLogsView()
    await flushPromises()

    const vm = wrapper.vm as any
    vm.filters.kind = 'error'
    vm.filters.statusPreset = '503'
    vm.filters.model = '5.4-mini'
    vm.applyFilters()
    await flushPromises()

    expect(requestRecordsList).toHaveBeenLastCalledWith(
      expect.objectContaining({
        kind: 'error',
        status_codes: '503',
        model: '5.4-mini',
      }),
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )

    vm.filters.statusPreset = '!200'
    vm.applyFilters()
    await flushPromises()

    const latestParams = requestRecordsList.mock.calls.at(-1)?.[0]
    expect(latestParams).toEqual(expect.objectContaining({
      kind: 'error',
      status_codes_other: true,
      model: '5.4-mini',
    }))
    expect(latestParams).not.toHaveProperty('status_codes')
  })

  it('passes session and billable filters to the request records API', async () => {
    const wrapper = mountRequestLogsView()
    await flushPromises()

    const vm = wrapper.vm as any
    vm.filters.session_id = 'sess-success'
    vm.filters.billable = 'true'
    vm.applyFilters()
    await flushPromises()

    expect(requestRecordsList).toHaveBeenLastCalledWith(
      expect.objectContaining({
        session_id: 'sess-success',
        billable: true,
      }),
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )
  })

  it('filters by session immediately when a request record session is clicked', async () => {
    const wrapper = mountRequestLogsView()
    await flushPromises()

    const sessionButton = wrapper.findAll('tbody button').find((button) => button.text().includes('sess-success'))
    expect(sessionButton).toBeTruthy()
    await sessionButton!.trigger('click')
    await flushPromises()

    expect(requestRecordsList).toHaveBeenLastCalledWith(
      expect.objectContaining({
        session_id: 'sess-success',
        page: 1,
      }),
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )
  })

  it('renders failed rows through the shared usage table and lets request ids be toggled on', async () => {
    const wrapper = mountRequestLogsView()
    await flushPromises()

    const bodyText = wrapper.find('tbody').text()
    expect(bodyText).toContain('503')
    expect(bodyText).not.toContain('503 service unavailable')
    const statusBadge = wrapper.findAll('tbody span').find((span) => span.text() === '503')
    expect(statusBadge).toBeTruthy()
    expect(statusBadge?.attributes('title')).toBeUndefined()
    await statusBadge!.trigger('mouseenter')
    await flushPromises()
    expect(document.body.querySelector('[role="tooltip"]')?.textContent).toContain('503 service unavailable')

    const columnButton = wrapper.find('button[title="Column Settings"]')
    expect(columnButton.exists()).toBe(true)
    await columnButton.trigger('click')

    const requestIdColumnButton = wrapper.findAll('button').find((button) => button.text().includes('Request ID'))
    expect(requestIdColumnButton).toBeTruthy()
    await requestIdColumnButton!.trigger('click')

    expect(wrapper.find('thead').text()).toContain('Request ID')
    expect(wrapper.find('tbody').text()).toContain('req-error')
    expect(wrapper.find('tbody').text()).toContain('req-success')
  })

  it('renders session from request records without exposing billing state as its own column', async () => {
    const wrapper = mountRequestLogsView()
    await flushPromises()

    const headerText = wrapper.find('thead').text()
    expect(headerText).toContain('Session')
    const headerColumns = wrapper.findAll('thead th').map((header) => header.text())
    expect(headerColumns).not.toContain('Billing')

    const bodyText = wrapper.find('tbody').text()
    expect(bodyText).toContain('sess-error')
    expect(bodyText).toContain('metadata_session_id')
    expect(bodyText).not.toContain('Non-billable')
    expect(bodyText).not.toContain('Billable')
  })

  it('shows pending request records as in progress without stream badges', async () => {
    requestRecordsList.mockResolvedValueOnce({
      items: [pendingRow],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
      summary: {
        total_requests: 1,
        success_requests: 0,
        error_requests: 0,
        error_rate: 0,
      },
    })

    const wrapper = mountRequestLogsView()
    await flushPromises()

    const bodyText = wrapper.find('tbody').text()
    expect(bodyText).toContain('Pending')
    expect(bodyText).not.toContain('Error')
    expect(bodyText).not.toContain('Stream')
  })

  it('renders usage-aligned default columns and lets hidden fields be toggled on', async () => {
    const wrapper = mountRequestLogsView()
    await flushPromises()

    const headerText = wrapper.find('thead').text()
    expect(headerText).toContain('Time')
    expect(headerText).toContain('User')
    expect(headerText).toContain('Account')
    expect(headerText).toContain('Billing Model')
    expect(headerText).toContain('Token')
    expect(headerText).toContain('Cache')
    expect(headerText).toContain('Cost')
    expect(headerText).toContain('Performance')
    expect(headerText).toContain('User-Agent')
    expect(headerText).toContain('Status')
    expect(headerText).not.toContain('API Key')
    expect(headerText).not.toContain('Request ID')
    expect(headerText).not.toContain('Source IP')
    expect(headerText).not.toContain('Platform')
    expect(headerText).not.toContain('Actions')
    expect(headerText).not.toContain('Group')
    expect(headerText).not.toContain('Inbound')
    expect(headerText).not.toContain('Error Summary')

    const bodyText = wrapper.find('tbody').text()
    expect(bodyText).toContain('200.50K')
    expect(bodyText).toContain('100.00K')
    expect(bodyText).toContain('$0.041000')
    expect(bodyText).toContain('TTL')
    expect(bodyText).toContain('XHigh')
    expect(bodyText).not.toContain('#42 tester')
    expect(bodyText).toContain('tester')
    expect(bodyText).toContain('openai-account')
    expect(bodyText).not.toContain('#9 openai-account')

    const tokenInfoButton = wrapper.find('button[aria-label="Token Breakdown"]')
    expect(tokenInfoButton.exists()).toBe(true)
    await tokenInfoButton.trigger('mouseenter')
    await flushPromises()
    expect(document.body.textContent).toContain('Token Breakdown')
    expect(document.body.textContent).toContain('Input Tokens')
    expect(document.body.textContent).toContain('Total Tokens')

    await tokenInfoButton.trigger('mouseleave')
    const costInfoButton = wrapper.find('button[aria-label="Cost Breakdown"]')
    expect(costInfoButton.exists()).toBe(true)
    await costInfoButton.trigger('mouseenter')
    await flushPromises()
    expect(document.body.textContent).toContain('Cost Breakdown')
    expect(document.body.textContent).toContain('Input price')
    expect(document.body.textContent).toContain('User Billed')

    const columnButton = wrapper.find('button[title="Column Settings"]')
    expect(columnButton.exists()).toBe(true)
    await columnButton.trigger('click')

    const apiKeyButton = wrapper.findAll('button').find((button) => button.text().includes('API Key'))
    expect(apiKeyButton).toBeTruthy()
    await apiKeyButton!.trigger('click')

    expect(wrapper.find('thead').text()).toContain('API Key')
    expect(localStorage.setItem).toHaveBeenCalledWith(
      'admin-request-logs-hidden-columns-v2',
      expect.any(String)
    )
  })

  it('renders image usage facts with usage-record semantics', async () => {
    requestRecordsList.mockResolvedValueOnce({
      items: [imageSuccessRow],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
      summary: {
        total_requests: 1,
        success_requests: 1,
        error_requests: 0,
        error_rate: 0,
      },
    })

    const wrapper = mountRequestLogsView()
    await flushPromises()

    const bodyText = wrapper.find('tbody').text()
    expect(bodyText).toContain('2 images')
    expect(bodyText).toContain('Image · 2K')
    expect(bodyText).toContain('$0.072000')
  })

  it('shows monitoring disabled state without loading request logs', async () => {
    opsState.enabled = false
    const wrapper = mountRequestLogsView()
    await flushPromises()

    expect(requestRecordsList).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Ops disabled')
    expect(wrapper.text()).toContain('Open Settings')
  })
})
