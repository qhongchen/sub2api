import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import UsageTable from '../UsageTable.vue'
import type { Column } from '@/components/common/types'

const messages: Record<string, string> = {
  'usage.costDetails': 'Cost Breakdown',
  'admin.usage.inputCost': 'Input Cost',
  'admin.usage.outputCost': 'Output Cost',
  'admin.usage.cacheCreationCost': 'Cache Creation Cost',
  'admin.usage.cacheReadCost': 'Cache Read Cost',
  'usage.inputTokenPrice': 'Input price',
  'usage.outputTokenPrice': 'Output price',
  'usage.perMillionTokens': '/ 1M tokens',
  'usage.serviceTier': 'Service tier',
  'usage.serviceTierPriority': 'Fast',
  'usage.serviceTierFlex': 'Flex',
  'usage.serviceTierStandard': 'Standard',
  'usage.rate': 'Rate',
  'usage.accountMultiplier': 'Account rate',
  'usage.original': 'Original',
  'usage.userBilled': 'User billed',
  'usage.accountBilled': 'Account billed',
  'usage.imageUnit': ' images',
  'usage.imageCount': 'Image count',
  'usage.imageBillingSize': 'Billing size',
  'usage.imageInputSize': 'Input size',
  'usage.imageOutputSize': 'Output size',
  'usage.imageSizeSource': 'Size source',
  'usage.imageSizeBreakdown': 'Size breakdown',
  'usage.imageSizeSourceOutput': 'Upstream output',
  'usage.imageSizeSourceInput': 'Request input',
  'usage.imageSizeSourceDefault': 'Default billing tier',
  'usage.imageSizeSourceLegacy': 'Legacy record',
  'usage.imageSizeSourceMissing': 'Not recorded',
  'usage.imageSizeNotRecorded': 'not recorded',
  'usage.imageSizeLegacyUnstandardized': 'legacy unstandardized',
  'usage.imageSizeUnknown': 'unknown',
  'usage.imageUnitPrice': 'Per-image price',
  'usage.imageTotalPrice': 'Image total price',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per request',
  'admin.usage.billingModeImage': 'Image',
  'usage.justNow': 'just now',
  'usage.tokens': 'Tokens',
  'usage.cost': 'Cost',
  'usage.requestId': 'Request ID',
  'usage.performance': 'Performance',
  'usage.unknownKey': 'Unknown key',
  'usage.upstreamAttempts': 'Upstream attempts',
  'usage.upstreamAttemptCount': '{count} upstream attempts',
  'usage.upstreamAttemptStatus': 'upstream {status}',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        let message = messages[key] ?? key
        for (const [name, value] of Object.entries(params || {})) {
          message = message.replaceAll(`{${name}}`, String(value))
        }
        return message
      },
    }),
  }
})

const tableColumns: Column[] = [
  { key: 'model', label: 'Model' },
  { key: 'tokens', label: 'Tokens' },
  { key: 'cost', label: 'Cost' },
]

const baseImageRow = {
  request_id: 'req-admin-image',
  model: 'gpt-image-2',
  actual_cost: 0.4,
  total_cost: 0.4,
  account_rate_multiplier: 1,
  rate_multiplier: 1,
  service_tier: null,
  input_cost: 0,
  output_cost: 0,
  cache_creation_cost: 0,
  cache_read_cost: 0,
  input_tokens: 0,
  output_tokens: 0,
  cache_creation_tokens: 0,
  cache_read_tokens: 0,
  cache_creation_5m_tokens: 0,
  cache_creation_1h_tokens: 0,
  cache_ttl_overridden: false,
  billing_mode: 'image',
  image_count: 2,
  image_size: '2K',
  image_input_size: null,
  image_output_size: null,
  image_size_source: null,
  image_size_breakdown: null,
}

describe('admin UsageTable tooltip', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
      x: 0,
      y: 0,
      top: 20,
      left: 20,
      right: 120,
      bottom: 40,
      width: 100,
      height: 20,
      toJSON: () => ({}),
    } as DOMRect)
  })

  it('shows service tier and billing breakdown in cost tooltip', async () => {
    const row = {
      request_id: 'req-admin-1',
      actual_cost: 0.092883,
      total_cost: 0.092883,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      service_tier: 'priority',
      input_cost: 0.020285,
      output_cost: 0.00303,
      cache_creation_cost: 0,
      cache_read_cost: 0.069568,
      input_tokens: 4057,
      output_tokens: 101,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: tableColumns,
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const costInfoButton = wrapper.find('button[aria-label="Cost Breakdown"]')
    expect(costInfoButton.exists()).toBe(true)
    await costInfoButton.trigger('mouseenter')
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Service tier')
    expect(text).toContain('Fast')
    expect(text).toContain('Rate')
    expect(text).toContain('1.00x')
    expect(text).toContain('Account rate')
    expect(text).toContain('User billed')
    expect(text).toContain('Account billed')
    expect(text).toContain('$0.092883')
    expect(text).toContain('$5.0000 / 1M tokens')
    expect(text).toContain('$30.0000 / 1M tokens')
    expect(text).toContain('$0.069568')
  })

  it('shows requested and upstream models separately for admin rows', () => {
    const row = {
      request_id: 'req-admin-model-1',
      model: 'claude-sonnet-4',
      upstream_model: 'claude-sonnet-4-20250514',
      actual_cost: 0,
      total_cost: 0,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: tableColumns,
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('claude-sonnet-4')
    expect(text).toContain('claude-sonnet-4-20250514')
  })

  it('renders pending status as a compact single-line badge', () => {
    const row = {
      request_id: 'req-admin-pending-1',
      kind: 'pending',
      outcome: 'pending',
      status_code: 0,
      duration_ms: null,
      actual_cost: null,
      total_cost: null,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
      user_agent: 'Codex Desktop/0.135.0-alpha.1 (macOS 15.5; arm64) renderer/123',
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [
          { key: 'status', label: 'Status' },
          { key: 'user_agent', label: 'User-Agent' },
        ],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const statusCell = wrapper.findAll('tbody td').at(0)
    expect(statusCell).toBeTruthy()
    expect(statusCell!.text()).toContain('Pending')
    expect(statusCell!.find('span').classes()).toEqual(expect.arrayContaining(['whitespace-nowrap']))

    const userAgentCell = wrapper.findAll('tbody td').at(1)
    expect(userAgentCell?.find('span').classes()).toEqual(expect.arrayContaining(['line-clamp-2']))
    expect(userAgentCell?.find('span').classes()).not.toContain('block')
  })

  it('renders non-billable cost as a dash even when cost fields are present', () => {
    const row = {
      request_id: 'req-admin-non-billable-1',
      kind: 'success',
      outcome: 'non_billable',
      billable: false,
      actual_cost: 0.041,
      total_cost: 0.037,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0.01,
      output_cost: 0.02,
      cache_creation_cost: 0.003,
      cache_read_cost: 0.004,
      input_tokens: 100,
      output_tokens: 50,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'cost', label: 'Cost' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.find('tbody').text()).toBe('-')
    expect(wrapper.find('tbody').text()).not.toContain('$0.041000')
    expect(wrapper.find('button[aria-label="Cost Breakdown"]').exists()).toBe(false)
  })

  it('middle-truncates session id and emits the full value when clicked', async () => {
    const row = {
      request_id: 'req-admin-session-1',
      session_id: '019e6eb8-b7e1-7483-a147-724f1bc99d31',
      session_source: 'header_session_id',
      prompt_cache_key: 'prompt-cache-key-should-not-render',
      actual_cost: null,
      total_cost: null,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'session', label: 'Session' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const sessionButton = wrapper.find('tbody button')
    expect(sessionButton.text()).toContain('019e6eb8-b7e...1bc99d31')
    expect(sessionButton.text()).not.toContain('prompt-cache-key-should-not-render')
    await sessionButton.trigger('click')

    expect(wrapper.emitted('session-click')).toEqual([['019e6eb8-b7e1-7483-a147-724f1bc99d31']])
  })

  it('keeps error details in the status tooltip instead of rendering an extra line', async () => {
    const row = {
      request_id: 'req-admin-error-1',
      kind: 'error',
      outcome: 'error',
      status_code: 500,
      duration_ms: 4600,
      message: 'upstream 500: model overloaded',
      phase: 'upstream_response',
      actual_cost: null,
      total_cost: null,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'status', label: 'Status' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const statusBadge = wrapper.find('tbody span')
    expect(statusBadge.text()).toBe('500')
    expect(statusBadge.attributes('title')).toBeUndefined()
    expect(wrapper.find('tbody').text()).not.toContain('upstream 500: model overloaded')

    await statusBadge.trigger('mouseenter')
    await nextTick()

    const tooltip = wrapper.find('[role="tooltip"]')
    expect(tooltip.exists()).toBe(true)
    expect(tooltip.text()).toContain('upstream 500: model overloaded')
  })

  it('uses request record error_message as the status tooltip', async () => {
    const row = {
      request_id: 'req-admin-error-message-1',
      kind: 'error',
      outcome: 'error',
      status_code: 503,
      duration_ms: 4600,
      error_message: 'Upstream service temporarily unavailable',
      phase: 'upstream_response',
      actual_cost: null,
      total_cost: null,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'status', label: 'Status' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const statusBadge = wrapper.find('tbody span')
    expect(statusBadge.text()).toBe('503')
    expect(statusBadge.attributes('title')).toBeUndefined()
    expect(wrapper.find('tbody').text()).not.toContain('Upstream service temporarily unavailable')

    await statusBadge.trigger('mouseenter')
    await nextTick()

    const tooltip = wrapper.find('[role="tooltip"]')
    expect(tooltip.exists()).toBe(true)
    expect(tooltip.text()).toContain('Upstream service temporarily unavailable')
  })

  it('does not show a status tooltip when the row has no status detail', async () => {
    const row = {
      request_id: 'req-admin-success-no-status-detail',
      kind: 'success',
      outcome: 'success',
      status_code: 200,
      duration_ms: 4600,
      actual_cost: 0,
      total_cost: 0,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'status', label: 'Status' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const statusBadge = wrapper.find('tbody span')
    expect(statusBadge.text()).toBe('200')

    await statusBadge.trigger('mouseenter')
    await nextTick()

    expect(wrapper.find('[role="tooltip"]').exists()).toBe(false)
  })

  it('keeps recovered upstream errors out of the successful status tooltip', async () => {
    const row = {
      request_id: 'req-admin-success-recovered-upstream',
      kind: 'success',
      outcome: 'success',
      status_code: 200,
      duration_ms: 4600,
      error_message: 'Recovered upstream error 401: Invalid API key',
      actual_cost: 0,
      total_cost: 0,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'status', label: 'Status' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const statusBadge = wrapper.find('tbody span')
    expect(statusBadge.text()).toBe('200')

    await statusBadge.trigger('mouseenter')
    await nextTick()

    expect(wrapper.find('[role="tooltip"]').exists()).toBe(false)
  })

  it('shows upstream attempts separately from the final status', async () => {
    const row = {
      request_id: 'req-admin-success-with-upstream-attempts',
      kind: 'success',
      outcome: 'success',
      status_code: 200,
      duration_ms: 4600,
      upstream_attempts: [
        {
          account_id: 20,
          account_name: 'L站福利1',
          upstream_status_code: 401,
          message: 'Invalid API key',
        },
        {
          account_id: 21,
          upstream_status_code: 429,
          message: 'rate limited',
        },
      ],
      actual_cost: 0,
      total_cost: 0,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'status', label: 'Status' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const statusCell = wrapper.find('tbody td')
    expect(statusCell.text()).toContain('200')
    expect(statusCell.text()).toContain('2 upstream attempts')
    expect(statusCell.text()).not.toContain('Invalid API key')

    const upstreamBadge = wrapper.find('[data-testid="upstream-attempts-badge"]')
    expect(upstreamBadge.exists()).toBe(true)
    await upstreamBadge.trigger('mouseenter')
    await nextTick()

    const tooltip = wrapper.find('[role="tooltip"]')
    expect(tooltip.exists()).toBe(true)
    expect(tooltip.text()).toContain('L站福利1')
    expect(tooltip.text()).toContain('upstream 401')
    expect(tooltip.text()).toContain('rate limited')
  })

  it('does not show a long-text tooltip when content is fully visible', async () => {
    vi.spyOn(HTMLElement.prototype, 'scrollWidth', 'get').mockReturnValue(80)
    vi.spyOn(HTMLElement.prototype, 'clientWidth', 'get').mockReturnValue(120)
    vi.spyOn(HTMLElement.prototype, 'scrollHeight', 'get').mockReturnValue(20)
    vi.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockReturnValue(24)

    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            request_id: 'req-admin-short-model',
            model: 'gpt-5',
            actual_cost: 0,
            total_cost: 0,
            account_rate_multiplier: 1,
            rate_multiplier: 1,
            input_cost: 0,
            output_cost: 0,
            cache_creation_cost: 0,
            cache_read_cost: 0,
            input_tokens: 0,
            output_tokens: 0,
          },
        ],
        loading: false,
        columns: [{ key: 'model', label: 'Model' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const modelText = wrapper.find('tbody td span.truncate')
    expect(modelText.text()).toBe('gpt-5')

    await modelText.trigger('mouseenter')
    await nextTick()

    expect(wrapper.find('[role="tooltip"]').exists()).toBe(false)
  })

  it('always shows full session details in the session tooltip on hover', async () => {
    vi.spyOn(HTMLElement.prototype, 'scrollWidth', 'get').mockReturnValue(80)
    vi.spyOn(HTMLElement.prototype, 'clientWidth', 'get').mockReturnValue(120)
    vi.spyOn(HTMLElement.prototype, 'scrollHeight', 'get').mockReturnValue(20)
    vi.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockReturnValue(24)

    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            request_id: 'req-admin-session-extra-1',
            session_id: 'visible-session',
            session_source: 'header_session_id',
            client_session_id: 'client-session-from-header',
            actual_cost: null,
            total_cost: null,
            account_rate_multiplier: 1,
            rate_multiplier: 1,
            input_cost: 0,
            output_cost: 0,
            cache_creation_cost: 0,
            cache_read_cost: 0,
            input_tokens: 0,
            output_tokens: 0,
          },
        ],
        loading: false,
        columns: [{ key: 'session', label: 'Session' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const sessionButton = wrapper.find('tbody button')
    await sessionButton.trigger('mouseenter')
    await nextTick()

    const tooltip = wrapper.find('[role="tooltip"]')
    expect(tooltip.exists()).toBe(true)
    expect(tooltip.text()).toContain('visible-session')
    expect(tooltip.text()).toContain('client-session-from-header')
  })

  it('renders the vendor model icon and keeps reasoning under the model text', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            request_id: 'req-admin-model-icon',
            model: '5.4-mini',
            upstream_model: 'gpt-5.5-20260601',
            model_mapping_chain: '5.4-mini→gpt-5.5',
            reasoning_effort: 'xhigh',
            actual_cost: 0,
            total_cost: 0,
            account_rate_multiplier: 1,
            rate_multiplier: 1,
            input_cost: 0,
            output_cost: 0,
            cache_creation_cost: 0,
            cache_read_cost: 0,
            input_tokens: 0,
            output_tokens: 0,
          },
        ],
        loading: false,
        columns: [{ key: 'model', label: 'Model' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
          ModelIcon: {
            props: ['model'],
            template: '<span data-testid="model-icon">{{ model }}</span>',
          },
        },
      },
    })

    const modelCell = wrapper.find('tbody td')
    const textColumn = modelCell.find('.min-w-0')
    expect(modelCell.find('[data-testid="model-icon"]').text()).toBe('gpt-5.5-20260601')
    expect(textColumn.text()).toContain('5.4-mini')
    expect(textColumn.text()).toContain('gpt-5.5-20260601')
    expect(textColumn.text()).toContain('XHigh')
  })

  it('uses the actual billing model as the primary model text for mapped rows', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            request_id: 'req-admin-mapped-billing-model',
            model: 'gpt-5.4',
            requested_model: 'gpt-5.4',
            upstream_model: 'gpt-5.5',
            billing_model: 'gpt-5.5',
            model_mapping_chain: 'gpt-5.4→gpt-5.5',
            actual_cost: 0.026144,
            total_cost: 0.026144,
            account_rate_multiplier: 1,
            rate_multiplier: 1,
            input_cost: 0.01,
            output_cost: 0.02,
            cache_creation_cost: 0,
            cache_read_cost: 0,
            input_tokens: 100,
            output_tokens: 50,
          },
        ],
        loading: false,
        columns: [{ key: 'model', label: 'Model' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
          ModelIcon: {
            props: ['model'],
            template: '<span data-testid="model-icon">{{ model }}</span>',
          },
        },
      },
    })

    const primaryModel = wrapper.find('tbody td .model-primary-text')
    expect(primaryModel.text()).toBe('gpt-5.5')
    expect(wrapper.find('tbody td').text()).toContain('gpt-5.4')
    expect(wrapper.find('[data-testid="model-icon"]').text()).toBe('gpt-5.5')
  })

  it('shows mapped models as a single requested-to-upstream relationship', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            request_id: 'req-admin-model-dedup',
            model: 'gpt-5.5',
            requested_model: 'gpt-5.4',
            upstream_model: 'gpt-5.5',
            billing_model: 'gpt-5.5',
            model_mapping_chain: 'gpt-5.4→gpt-5.5',
            reasoning_effort: 'xhigh',
            actual_cost: 0.026144,
            total_cost: 0.026144,
            account_rate_multiplier: 1,
            rate_multiplier: 1,
            input_cost: 0.01,
            output_cost: 0.02,
            cache_creation_cost: 0,
            cache_read_cost: 0,
            input_tokens: 100,
            output_tokens: 50,
          },
        ],
        loading: false,
        columns: [{ key: 'model', label: 'Model' }],
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
          ModelIcon: {
            props: ['model'],
            template: '<span data-testid="model-icon">{{ model }}</span>',
          },
        },
      },
    })

    const modelText = wrapper.find('tbody td .model-primary-text').element.parentElement?.textContent || ''
    expect(wrapper.find('tbody td .model-primary-text').text()).toBe('gpt-5.5')
    expect(wrapper.find('tbody td .model-secondary-text').text()).toBe('gpt-5.4 → gpt-5.5')
    expect(modelText).toContain('XHigh')
  })

  it.each([
    {
      name: 'defaulted row',
      row: {
        ...baseImageRow,
        request_id: 'req-admin-default-image',
        image_size: '2K',
        image_input_size: 'auto',
        image_output_size: null,
        image_size_source: 'default',
      },
      expected: ['2K', 'Default billing tier', 'auto', 'unknown'],
    },
    {
      name: 'output-sourced row',
      row: {
        ...baseImageRow,
        request_id: 'req-admin-output-image',
        image_size: '4K',
        image_input_size: '1024x1024',
        image_output_size: '3840x2160',
        image_size_source: 'output',
        image_size_breakdown: { '4K': 1 },
      },
      expected: ['4K', 'Upstream output', '1024x1024', '3840x2160', '4K x 1'],
    },
    {
      name: 'input-sourced row',
      row: {
        ...baseImageRow,
        request_id: 'req-admin-input-image',
        image_size: '1K',
        image_input_size: '1024x1024',
        image_output_size: null,
        image_size_source: 'input',
      },
      expected: ['1K', 'Request input', '1024x1024', 'unknown'],
    },
    {
      name: 'legacy unstandardized row',
      row: {
        ...baseImageRow,
        request_id: 'req-admin-legacy-unstandardized-image',
        image_size: '512x512',
        image_input_size: null,
        image_output_size: null,
        image_size_source: null,
      },
      expected: ['legacy unstandardized: 512x512', 'Legacy record', 'unknown'],
    },
  ])('shows image usage metadata for $name', async ({ row, expected }) => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: tableColumns,
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const costInfoButton = wrapper.find('button[aria-label="Cost Breakdown"]')
    expect(costInfoButton.exists()).toBe(true)
    await costInfoButton.trigger('mouseenter')
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Image count')
    expect(text).toContain('Billing size')
    expect(text).toContain('Size source')
    expect(text).toContain('Input size')
    expect(text).toContain('Output size')
    expect(text).toContain('Per-image price')
    for (const value of expected) {
      expect(text).toContain(value)
    }
  })

  it('displays historical image rows with missing billing_mode as image usage without a 2K fallback', async () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-admin-legacy-missing-image',
            billing_mode: null,
            image_size: null,
            image_input_size: null,
            image_output_size: null,
            image_size_source: null,
            image_size_breakdown: null,
          },
        ],
        loading: false,
        columns: tableColumns,
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true,
        },
      },
    })

    const costInfoButton = wrapper.find('button[aria-label="Cost Breakdown"]')
    expect(costInfoButton.exists()).toBe(true)
    await costInfoButton.trigger('mouseenter')
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Image')
    expect(text).toContain('Image count')
    expect(text).toContain('Per-image price')
    expect(text).toContain('not recorded')
    expect(text).not.toContain('(2K)')
  })
})
