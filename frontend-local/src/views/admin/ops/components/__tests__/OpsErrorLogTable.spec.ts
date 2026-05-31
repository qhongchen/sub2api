import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import OpsErrorLogTable from '../OpsErrorLogTable.vue'
import type { OpsErrorLog } from '@/api/admin/ops'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('../../utils/opsFormatters', () => ({
  formatDateTime: (value: string) => `2026-05-29 ${value.slice(11, 19)}`,
  getSeverityClass: () => 'severity-stub',
}))

const ElTooltipStub = defineComponent({
  name: 'ElTooltip',
  props: {
    content: { type: String, default: '' },
    placement: { type: String, default: '' },
    showAfter: { type: Number, default: 0 },
  },
  template: '<span class="el-tooltip-stub" :data-content="content" :data-show-after="showAfter"><slot /></span>',
})

const PaginationStub = defineComponent({
  name: 'Pagination',
  template: '<div class="pagination-stub" />',
})

function makeLog(overrides: Partial<OpsErrorLog> = {}): OpsErrorLog {
  return {
    id: 1,
    created_at: '2026-05-29T17:29:09.000Z',
    phase: 'request',
    type: 'api_error',
    error_owner: 'provider',
    error_source: 'upstream_http',
    severity: 'error',
    status_code: 502,
    upstream_status_code: 401,
    platform: 'openai',
    model: 'gpt-5.5',
    resolved: false,
    client_request_id: 'client-1',
    request_id: 'server-1',
    message: 'unauthorized upstream',
    user_email: 'user@example.test',
    account_name: 'account-a',
    group_name: 'group-a',
    inbound_endpoint: '/v1/responses',
    upstream_endpoint: '/v1/responses',
    requested_model: 'gpt-5.5',
    request_type: 1,
    ...overrides,
  }
}

function mountTable(rows: OpsErrorLog[]) {
  return mount(OpsErrorLogTable, {
    props: {
      rows,
      total: rows.length,
      loading: false,
      page: 1,
      pageSize: 10,
    },
    global: {
      stubs: {
        ElTooltip: ElTooltipStub,
        Pagination: PaginationStub,
      },
    },
  })
}

describe('OpsErrorLogTable', () => {
  it('主状态码显示客户端状态码，同时展示不同的上游状态码', () => {
    const wrapper = mountTable([makeLog()])

    expect(wrapper.text()).toContain('502')
    expect(wrapper.text()).toContain('upstream 401')
    expect(wrapper.find('[data-content="Client 502 / Upstream 401"]').exists()).toBe(true)
    expect(wrapper.find('[data-content="Client 502 / Upstream 401"]').attributes('data-show-after')).toBe('0')
  })

  it('上游状态码与客户端状态码一致时不展示辅助标记', () => {
    const wrapper = mountTable([makeLog({ upstream_status_code: 502 })])

    expect(wrapper.text()).toContain('502')
    expect(wrapper.text()).not.toContain('upstream 502')
  })

  it('长文本字段不使用原生 title 延迟提示', () => {
    const wrapper = mountTable([
      makeLog({
        upstream_status_code: 502,
        requested_model: '',
        upstream_model: '',
        model: 'very-long-model-name-for-tooltip',
        message: 'very long upstream error payload',
      }),
    ])

    expect(wrapper.find('[title]').exists()).toBe(false)
    expect(wrapper.find('[data-content="very-long-model-name-for-tooltip"]').exists()).toBe(true)
    expect(wrapper.find('[data-content="very long upstream error payload"]').exists()).toBe(true)
    for (const tooltip of wrapper.findAll('.el-tooltip-stub')) {
      expect(tooltip.attributes('data-show-after')).toBe('0')
    }
  })
})
