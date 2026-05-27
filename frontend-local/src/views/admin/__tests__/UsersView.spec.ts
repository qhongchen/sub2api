import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import type { AdminUser } from '@/types'
import UsersView from '../UsersView.vue'

const {
  listUsers,
  getAllGroups,
  getBatchUsersUsage,
  listEnabledDefinitions,
  getBatchUserAttributes
} = vi.hoisted(() => ({
  listUsers: vi.fn(),
  getAllGroups: vi.fn(),
  getBatchUsersUsage: vi.fn(),
  listEnabledDefinitions: vi.fn(),
  getBatchUserAttributes: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      list: listUsers,
      toggleStatus: vi.fn(),
      delete: vi.fn()
    },
    groups: {
      getAll: getAllGroups
    },
    dashboard: {
      getBatchUsersUsage
    },
    userAttributes: {
      listEnabledDefinitions,
      getBatchUserAttributes
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const createAdminUser = (): AdminUser => ({
  id: 42,
  username: 'scoped-user',
  email: 'scoped@example.com',
  role: 'user',
  balance: 0,
  concurrency: 1,
  status: 'active',
  allowed_groups: [],
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-04-17T00:00:00Z',
  updated_at: '2026-04-17T00:00:00Z',
  notes: '',
  last_active_at: '2026-04-16T02:00:00Z',
  last_used_at: '2026-04-17T02:00:00Z',
  current_concurrency: 0
})

const DataTableStub = {
  props: ['columns', 'data', 'virtualized'],
  emits: ['sort'],
  template: `
    <div>
      <div data-test="columns">{{ columns.map(col => col.key).join(',') }}</div>
      <div data-test="virtualized">{{ String(virtualized) }}</div>
      <button data-test="sort-status" @click="$emit('sort', 'status', 'desc')">sort</button>
      <div v-for="row in data" :key="row.id">
        <slot name="cell-email" :value="row.email" :row="row" />
        <slot name="cell-status" :value="row.status" :row="row" />
      </div>
    </div>
  `
}

describe('admin UsersView', () => {
  beforeEach(() => {
    localStorage.clear()

    listUsers.mockReset()
    getAllGroups.mockReset()
    getBatchUsersUsage.mockReset()
    listEnabledDefinitions.mockReset()
    getBatchUserAttributes.mockReset()

    listUsers.mockResolvedValue({
      items: [createAdminUser()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    getAllGroups.mockResolvedValue([])
    getBatchUsersUsage.mockResolvedValue({ stats: {} })
    listEnabledDefinitions.mockResolvedValue([])
    getBatchUserAttributes.mockResolvedValue({ values: {} })
  })

  it('shows compact default columns and requests visible-column sort', async () => {
    const wrapper = mount(UsersView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="table" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: true,
          EmptyState: true,
          GroupBadge: true,
          Select: true,
          UserAttributesConfigModal: true,
          UserConcurrencyCell: true,
          UserCreateModal: true,
          UserEditModal: true,
          UserApiKeysModal: true,
          UserAllowedGroupsModal: true,
          UserBalanceModal: true,
          UserBalanceHistoryModal: true,
          GroupReplaceModal: true,
          Icon: true,
          Teleport: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.find('.users-management-card').exists()).toBe(true)
    const columns = wrapper.get('[data-test="columns"]').text()
    const visibleColumns = columns.split(',')
    expect(visibleColumns).toEqual([
      'email',
      'role',
      'usage',
      'concurrency',
      'status',
      'actions'
    ])
    expect(visibleColumns).not.toContain('username')
    expect(visibleColumns).not.toContain('last_login_at')
    expect(visibleColumns).not.toContain('id')
    expect(visibleColumns).not.toContain('created_at')
    expect(visibleColumns).not.toContain('last_used_at')
    expect(wrapper.text()).toContain('scoped@example.com')
    expect(wrapper.text()).toContain('scoped-user')
    expect(wrapper.get('[data-test="virtualized"]').text()).toBe('false')

    await wrapper.get('[data-test="sort-status"]').trigger('click')
    await flushPromises()

    expect(listUsers).toHaveBeenLastCalledWith(
      1,
      20,
      expect.objectContaining({
        sort_by: 'status',
        sort_order: 'desc'
      }),
      expect.any(Object)
    )
  })
})
