<template>
  <AppLayout>
    <div class="space-y-5">
      <section class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-semibold text-gray-950 dark:text-white">
            {{ t('admin.requestLogs.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">
            {{ t('admin.requestLogs.description') }}
          </p>
        </div>

        <button
          type="button"
          class="consumer-icon-btn"
          :title="t('admin.requestLogs.refresh')"
          :disabled="loading || opsDisabled"
          @click="refreshData"
        >
          <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
        </button>
      </section>

      <section
        v-if="opsDisabled"
        class="rounded-lg border border-amber-200 bg-amber-50 p-6 text-amber-900 shadow-sm dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-100"
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="flex gap-3">
            <span class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-200">
              <Icon name="exclamationTriangle" size="sm" :stroke-width="2" />
            </span>
            <div>
              <h2 class="text-base font-semibold">{{ t('admin.requestLogs.disabledTitle') }}</h2>
              <p class="mt-1 max-w-2xl text-sm text-amber-800 dark:text-amber-100/80">
                {{ t('admin.requestLogs.disabledDescription') }}
              </p>
            </div>
          </div>
          <router-link
            to="/admin/settings"
            class="inline-flex items-center justify-center gap-2 rounded-lg bg-amber-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-amber-700"
          >
            <Icon name="cog" size="sm" />
            {{ t('admin.requestLogs.openSettings') }}
          </router-link>
        </div>
      </section>

      <template v-else>
        <section class="cch-panel-card p-4 md:p-5">
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200/70 bg-white/70 px-3 py-2 text-sm font-medium text-gray-600 shadow-sm transition-colors hover:border-gray-300 hover:text-gray-950 dark:border-white/[0.08] dark:bg-dark-900/60 dark:text-dark-300 dark:hover:text-white"
              @click="filtersOpen = !filtersOpen"
            >
              <Icon name="filter" size="sm" :stroke-width="2" />
              <span>{{ t('admin.requestLogs.filters') }}</span>
              <span
                v-if="activeFilterCount > 0"
                class="rounded-full bg-orange-50 px-1.5 py-0.5 text-[10px] font-semibold text-orange-600 dark:bg-orange-500/10 dark:text-orange-300"
              >
                {{ activeFilterCount }}
              </span>
              <Icon
                name="chevronDown"
                size="xs"
                class="transition-transform"
                :class="{ 'rotate-180': filtersOpen }"
              />
            </button>

            <div class="flex items-center gap-2">
              <div ref="columnDropdownRef" class="relative">
                <button
                  type="button"
                  class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-800 shadow-sm transition-colors hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-100 dark:hover:bg-white/[0.04]"
                  :title="t('admin.requestLogs.columnSettings')"
                  @click.stop="showColumnDropdown = !showColumnDropdown"
                >
                  <Icon name="grid" size="sm" />
                  <span>{{ visibleColumnCount }}/{{ allColumns.length }}</span>
                </button>

                <div
                  v-if="showColumnDropdown"
                  class="absolute right-0 top-full z-50 mt-2 max-h-96 w-60 overflow-y-auto rounded-lg border border-gray-200 bg-white p-1 shadow-lg dark:border-dark-700 dark:bg-dark-800"
                >
                  <div class="px-3 py-2 text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-dark-400">
                    {{ t('admin.requestLogs.columnSettings') }}
                  </div>
                  <button
                    v-for="col in toggleableColumns"
                    :key="col.key"
                    type="button"
                    class="flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-left text-sm text-gray-600 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
                    @click="toggleColumn(col.key)"
                  >
                    <span class="truncate">{{ col.label }}</span>
                    <Icon v-if="isColumnVisible(col.key)" name="check" size="sm" class="shrink-0 text-orange-500" :stroke-width="2" />
                  </button>
                  <div class="my-1 border-t border-gray-100 dark:border-white/[0.06]" />
                  <button
                    type="button"
                    class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-400 dark:hover:bg-dark-700 dark:hover:text-white"
                    @click="resetColumns"
                  >
                    <Icon name="refresh" size="xs" />
                    {{ t('admin.requestLogs.resetColumns') }}
                  </button>
                </div>
              </div>
              <button type="button" class="btn btn-ghost btn-sm" @click="resetFilters">
                {{ t('admin.requestLogs.reset') }}
              </button>
              <button type="button" class="btn btn-primary btn-sm" :disabled="loading" @click="applyFilters">
                {{ t('admin.requestLogs.apply') }}
              </button>
            </div>
          </div>

          <Transition
            enter-active-class="transition duration-200 ease-out"
            enter-from-class="-translate-y-1 opacity-0"
            enter-to-class="translate-y-0 opacity-100"
            leave-active-class="transition duration-150 ease-in"
            leave-from-class="translate-y-0 opacity-100"
            leave-to-class="-translate-y-1 opacity-0"
          >
            <div v-if="filtersOpen" class="space-y-4">
              <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.timeRange') }}</span>
                  <select v-model="filters.time_range" class="input" @change="onTimeRangeChange">
                    <option v-for="option in timeRangeOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </option>
                  </select>
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.status') }}</span>
                  <select v-model="filters.kind" class="input">
                    <option value="all">{{ t('admin.requestLogs.allStatuses') }}</option>
                    <option value="success">{{ t('admin.requestLogs.kindSuccess') }}</option>
                    <option value="error">{{ t('admin.requestLogs.kindError') }}</option>
                  </select>
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.billable') }}</span>
                  <select v-model="filters.billable" class="input">
                    <option value="">{{ t('admin.requestLogs.billableOptions.all') }}</option>
                    <option value="true">{{ t('admin.requestLogs.billableOptions.true') }}</option>
                    <option value="false">{{ t('admin.requestLogs.billableOptions.false') }}</option>
                  </select>
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.statusCode') }}</span>
                  <select v-model="filters.statusPreset" class="input">
                    <option value="">{{ t('admin.requestLogs.allStatusCodes') }}</option>
                    <option value="200">200</option>
                    <option value="!200">{{ t('admin.requestLogs.excludeStatus200') }}</option>
                    <option value="4xx">4xx</option>
                    <option value="5xx">5xx</option>
                    <option value="503">503</option>
                    <option value="429">429</option>
                    <option value="custom">{{ t('admin.requestLogs.customStatusCode') }}</option>
                  </select>
                </label>

                <label v-if="filters.statusPreset === 'custom'" class="block">
                  <span class="input-label">{{ t('admin.requestLogs.customStatusCode') }}</span>
                  <input
                    v-model.trim="filters.customStatusCode"
                    type="number"
                    min="100"
                    max="599"
                    class="input"
                    placeholder="503"
                    @keyup.enter="applyFilters"
                  />
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.platform') }}</span>
                  <select v-model="filters.platform" class="input">
                    <option value="">{{ t('admin.requestLogs.allPlatforms') }}</option>
                    <option v-for="option in platformOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </option>
                  </select>
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.model') }}</span>
                  <input
                    v-model.trim="filters.model"
                    type="search"
                    class="input"
                    :placeholder="t('admin.requestLogs.modelPlaceholder')"
                    @keyup.enter="applyFilters"
                  />
                </label>

                <label ref="userSearchRef" class="relative block">
                  <span class="input-label">{{ t('admin.requestLogs.user') }}</span>
                  <input
                    v-model="userKeyword"
                    type="search"
                    class="input pr-8"
                    :placeholder="t('admin.requestLogs.searchUserPlaceholder')"
                    @focus="showUserDropdown = true"
                    @input="debounceUserSearch"
                    @keyup.enter="applyFilters"
                  />
                  <button
                    v-if="filters.user_id"
                    type="button"
                    class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                    @click="clearUser"
                  >
                    <Icon name="x" size="xs" />
                  </button>
                  <div
                    v-if="showUserDropdown && userResults.length > 0"
                    class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                  >
                    <button
                      v-for="user in userResults"
                      :key="user.id"
                      type="button"
                      class="flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                      @click="selectUser(user)"
                    >
                      <span class="truncate">{{ getUserDisplayName(user) }}</span>
                      <span class="text-xs text-gray-400">#{{ user.id }}</span>
                    </button>
                  </div>
                </label>

                <label ref="apiKeySearchRef" class="relative block">
                  <span class="input-label">{{ t('admin.requestLogs.apiKey') }}</span>
                  <input
                    v-model="apiKeyKeyword"
                    type="search"
                    class="input pr-8"
                    :placeholder="t('admin.requestLogs.searchApiKeyPlaceholder')"
                    @focus="onApiKeyFocus"
                    @input="debounceApiKeySearch"
                    @keyup.enter="applyFilters"
                  />
                  <button
                    v-if="filters.api_key_id"
                    type="button"
                    class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                    @click="clearApiKey"
                  >
                    <Icon name="x" size="xs" />
                  </button>
                  <div
                    v-if="showApiKeyDropdown && apiKeyResults.length > 0"
                    class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                  >
                    <button
                      v-for="key in apiKeyResults"
                      :key="key.id"
                      type="button"
                      class="flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                      @click="selectApiKey(key)"
                    >
                      <span class="truncate">{{ key.name || `#${key.id}` }}</span>
                      <span class="text-xs text-gray-400">#{{ key.id }}</span>
                    </button>
                  </div>
                </label>

                <label ref="accountSearchRef" class="relative block">
                  <span class="input-label">{{ t('admin.requestLogs.account') }}</span>
                  <input
                    v-model="accountKeyword"
                    type="search"
                    class="input pr-8"
                    :placeholder="t('admin.requestLogs.searchAccountPlaceholder')"
                    @focus="showAccountDropdown = true"
                    @input="debounceAccountSearch"
                    @keyup.enter="applyFilters"
                  />
                  <button
                    v-if="filters.account_id"
                    type="button"
                    class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                    @click="clearAccount"
                  >
                    <Icon name="x" size="xs" />
                  </button>
                  <div
                    v-if="showAccountDropdown && accountResults.length > 0"
                    class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                  >
                    <button
                      v-for="account in accountResults"
                      :key="account.id"
                      type="button"
                      class="flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                      @click="selectAccount(account)"
                    >
                      <span class="truncate">{{ account.name || `#${account.id}` }}</span>
                      <span class="text-xs text-gray-400">#{{ account.id }}</span>
                    </button>
                  </div>
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.group') }}</span>
                  <select v-model="filters.group_id" class="input">
                    <option value="">{{ t('admin.requestLogs.allGroups') }}</option>
                    <option v-for="group in groupOptions" :key="group.id" :value="String(group.id)">
                      {{ group.name }}
                    </option>
                  </select>
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.requestId') }}</span>
                  <input
                    v-model.trim="filters.request_id"
                    type="search"
                    class="input font-mono"
                    :placeholder="t('admin.requestLogs.requestIdPlaceholder')"
                    @keyup.enter="applyFilters"
                  />
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.sessionId') }}</span>
                  <input
                    v-model.trim="filters.session_id"
                    type="search"
                    class="input font-mono"
                    :placeholder="t('admin.requestLogs.sessionIdPlaceholder')"
                    @keyup.enter="applyFilters"
                  />
                </label>

                <label class="block md:col-span-2">
                  <span class="input-label">{{ t('admin.requestLogs.keyword') }}</span>
                  <input
                    v-model.trim="filters.q"
                    type="search"
                    class="input"
                    :placeholder="t('admin.requestLogs.keywordPlaceholder')"
                    @keyup.enter="applyFilters"
                  />
                </label>

                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.sort') }}</span>
                  <select v-model="filters.sort" class="input">
                    <option value="created_at_desc">{{ t('admin.requestLogs.sortNewest') }}</option>
                    <option value="duration_desc">{{ t('admin.requestLogs.sortDuration') }}</option>
                  </select>
                </label>
              </div>

              <div v-if="filters.time_range === 'custom'" class="grid gap-3 md:grid-cols-2">
                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.startTime') }}</span>
                  <input v-model="customStartTime" type="datetime-local" class="input" @keyup.enter="applyFilters" />
                </label>
                <label class="block">
                  <span class="input-label">{{ t('admin.requestLogs.endTime') }}</span>
                  <input v-model="customEndTime" type="datetime-local" class="input" @keyup.enter="applyFilters" />
                </label>
              </div>
            </div>
          </Transition>
        </section>

        <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <div
            v-for="card in summaryCards"
            :key="card.key"
            class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-white/[0.08] dark:bg-dark-900"
          >
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm font-medium text-gray-500 dark:text-dark-300">{{ card.label }}</p>
              <span class="flex h-8 w-8 items-center justify-center rounded-lg" :class="card.iconClass">
                <Icon :name="card.icon" size="sm" :stroke-width="2" />
              </span>
            </div>
            <p class="mt-3 text-2xl font-semibold text-gray-950 dark:text-white">{{ card.value }}</p>
          </div>
        </section>

        <section class="cch-panel-card overflow-hidden">
          <div v-if="errorMessage" class="border-b border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/20 dark:bg-red-500/10 dark:text-red-200">
            {{ errorMessage }}
          </div>

          <div class="flex items-center justify-between gap-3 border-b border-gray-200/70 px-4 py-3 dark:border-white/[0.06]">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ t('usage.loadedRecords', { count: logs.length }) }}
            </span>
            <span class="text-xs text-gray-400 dark:text-dark-500">
              {{ requestLogStatsSummary }}
            </span>
          </div>

          <UsageTable
            :data="logs"
            :loading="loading"
            :columns="visibleColumns"
            shell-class="overflow-hidden"
            :server-side-sort="false"
            :default-sort-key="'created_at'"
            :default-sort-order="'desc'"
            :show-request-type-badge="false"
            @session-click="filterBySession"
          />

          <div v-if="pagination.total > 0" class="border-t border-gray-200/70 p-4 dark:border-white/[0.06]">
            <Pagination
              :total="pagination.total"
              :page="pagination.page"
              :page-size="pagination.page_size"
              @update:page="handlePageChange"
              @update:page-size="handlePageSizeChange"
            />
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import { adminAPI } from '@/api'
import type {
  AdminRequestLog,
  AdminRequestLogKindFilter,
  AdminRequestLogQueryParams,
  AdminRequestLogSort,
  AdminRequestLogSummary
} from '@/api/admin/requestRecords'
import type { SimpleApiKey, SimpleUser } from '@/api/admin/usage'
import { useAdminSettingsStore, useAppStore } from '@/stores'
import { formatNumber } from '@/utils/format'
import type { Column } from '@/components/common/types'

type TimeRange = '5m' | '30m' | '1h' | '6h' | '24h' | 'custom'
type BillableFilter = '' | 'true' | 'false'
type StatusPreset = '' | '200' | '!200' | '4xx' | '5xx' | '503' | '429' | 'custom'
type RequestLogColumn = Column & { key: RequestLogColumnKey }
type RequestLogColumnKey =
  | 'time'
  | 'user'
  | 'account'
  | 'session'
  | 'model'
  | 'tokens'
  | 'cache'
  | 'cost'
  | 'performance'
  | 'user_agent'
  | 'status'
  | 'api_key'
  | 'request_id'
  | 'ip_address'
  | 'platform'
  | 'group'
  | 'endpoint'

interface RequestLogFilters {
  time_range: TimeRange
  kind: AdminRequestLogKindFilter
  billable: BillableFilter
  statusPreset: StatusPreset
  customStatusCode: string
  platform: string
  model: string
  user_id?: number
  api_key_id?: number
  account_id?: number
  group_id: string
  request_id: string
  session_id: string
  q: string
  sort: AdminRequestLogSort
}

interface SimpleAccount {
  id: number
  name: string
}

interface SimpleGroup {
  id: number
  name: string
}

const { t } = useI18n()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()

const loading = ref(false)
const errorMessage = ref('')
const opsDisabled = ref(false)
const filtersOpen = ref(true)
const logs = ref<AdminRequestLog[]>([])
const summary = ref<AdminRequestLogSummary>({
  total_requests: 0,
  success_requests: 0,
  error_requests: 0,
  error_rate: 0
})
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive<RequestLogFilters>({
  time_range: '1h',
  kind: 'all',
  billable: '',
  statusPreset: '',
  customStatusCode: '',
  platform: '',
  model: '',
  group_id: '',
  request_id: '',
  session_id: '',
  q: '',
  sort: 'created_at_desc'
})

const customStartTime = ref('')
const customEndTime = ref('')
const userKeyword = ref('')
const apiKeyKeyword = ref('')
const accountKeyword = ref('')
const userResults = ref<SimpleUser[]>([])
const apiKeyResults = ref<SimpleApiKey[]>([])
const accountResults = ref<SimpleAccount[]>([])
const groupOptions = ref<SimpleGroup[]>([])
const showUserDropdown = ref(false)
const showApiKeyDropdown = ref(false)
const showAccountDropdown = ref(false)
const userSearchRef = ref<HTMLElement | null>(null)
const apiKeySearchRef = ref<HTMLElement | null>(null)
const accountSearchRef = ref<HTMLElement | null>(null)
const columnDropdownRef = ref<HTMLElement | null>(null)
const showColumnDropdown = ref(false)

let abortController: AbortController | null = null
let userSearchTimeout: ReturnType<typeof setTimeout> | null = null
let apiKeySearchTimeout: ReturnType<typeof setTimeout> | null = null
let accountSearchTimeout: ReturnType<typeof setTimeout> | null = null

const timeRangeOptions = computed(() => [
  { value: '5m', label: t('admin.ops.timeRange.5m') },
  { value: '30m', label: t('admin.ops.timeRange.30m') },
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '6h', label: t('admin.ops.timeRange.6h') },
  { value: '24h', label: t('admin.ops.timeRange.24h') },
  { value: 'custom', label: t('admin.ops.timeRange.custom') }
] satisfies Array<{ value: TimeRange; label: string }>)

const platformOptions = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'antigravity', label: 'Antigravity' },
  { value: 'unknown', label: 'unknown' }
]

const DEFAULT_VISIBLE_COLUMNS: RequestLogColumnKey[] = [
  'time',
  'user',
  'account',
  'session',
  'model',
  'tokens',
  'cache',
  'cost',
  'performance',
  'user_agent',
  'status'
]
const HIDDEN_COLUMNS_KEY = 'admin-request-logs-hidden-columns-v2'
const hiddenColumns = reactive<Set<RequestLogColumnKey>>(new Set())

const allColumns = computed<RequestLogColumn[]>(() => [
  { key: 'time', label: t('admin.requestLogs.table.time') },
  { key: 'user', label: t('admin.requestLogs.table.user') },
  { key: 'account', label: t('admin.requestLogs.table.account') },
  { key: 'session', label: t('admin.requestLogs.table.session') },
  { key: 'model', label: t('admin.requestLogs.table.billingModel') },
  { key: 'tokens', label: t('admin.requestLogs.table.tokens') },
  { key: 'cache', label: t('admin.requestLogs.table.cache') },
  { key: 'cost', label: t('admin.requestLogs.table.cost') },
  { key: 'performance', label: t('admin.requestLogs.table.performance') },
  { key: 'user_agent', label: t('admin.requestLogs.table.userAgent') },
  { key: 'status', label: t('admin.requestLogs.table.status') },
  { key: 'api_key', label: t('admin.requestLogs.table.apiKey') },
  { key: 'request_id', label: t('admin.requestLogs.table.requestId') },
  { key: 'ip_address', label: t('admin.requestLogs.table.sourceIp') },
  { key: 'platform', label: t('admin.requestLogs.table.platform') },
  { key: 'group', label: t('admin.requestLogs.table.group') },
  { key: 'endpoint', label: t('admin.requestLogs.table.inboundEndpoint') }
])

const toggleableColumns = computed<RequestLogColumn[]>(() => allColumns.value)
const visibleColumns = computed<RequestLogColumn[]>(() => allColumns.value.filter((column) => !hiddenColumns.has(column.key)))
const visibleColumnCount = computed(() => visibleColumns.value.length)
const allColumnKeys = computed(() => new Set(allColumns.value.map((column) => column.key)))

function isRequestLogColumnKey(value: string): value is RequestLogColumnKey {
  return allColumnKeys.value.has(value as RequestLogColumnKey)
}

function defaultHiddenColumns(): RequestLogColumnKey[] {
  return allColumns.value
    .map((column) => column.key)
    .filter((key): key is RequestLogColumnKey => !DEFAULT_VISIBLE_COLUMNS.includes(key))
}

function saveHiddenColumns() {
  try {
    localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
  } catch (error) {
    console.error('[RequestLogsView] Failed to save column visibility', error)
  }
}

function loadSavedColumns() {
  hiddenColumns.clear()
  const defaults = defaultHiddenColumns()
  try {
    const raw = localStorage.getItem(HIDDEN_COLUMNS_KEY)
    const parsed = raw ? JSON.parse(raw) : defaults
    const keys = Array.isArray(parsed) ? parsed : defaults
    keys.forEach((key) => {
      if (typeof key === 'string' && isRequestLogColumnKey(key)) {
        hiddenColumns.add(key)
      }
    })
  } catch {
    defaults.forEach((key) => hiddenColumns.add(key))
  }
}

function isColumnVisible(key: RequestLogColumnKey): boolean {
  return !hiddenColumns.has(key)
}

function toggleColumn(key: RequestLogColumnKey) {
  if (hiddenColumns.has(key)) {
    hiddenColumns.delete(key)
  } else {
    hiddenColumns.add(key)
  }
  saveHiddenColumns()
}

function resetColumns() {
  hiddenColumns.clear()
  defaultHiddenColumns().forEach((key) => hiddenColumns.add(key))
  saveHiddenColumns()
}

const summaryCards = computed(() => [
  {
    key: 'total',
    label: t('admin.requestLogs.totalRequests'),
    value: formatNumber(summary.value.total_requests),
    icon: 'chartBar' as const,
    iconClass: 'bg-gray-100 text-gray-600 dark:bg-white/[0.06] dark:text-dark-200'
  },
  {
    key: 'success',
    label: t('admin.requestLogs.successRequests'),
    value: formatNumber(summary.value.success_requests),
    icon: 'checkCircle' as const,
    iconClass: 'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-300'
  },
  {
    key: 'error',
    label: t('admin.requestLogs.errorRequests'),
    value: formatNumber(summary.value.error_requests),
    icon: 'xCircle' as const,
    iconClass: 'bg-red-50 text-red-600 dark:bg-red-500/10 dark:text-red-300'
  },
  {
    key: 'rate',
    label: t('admin.requestLogs.errorRate'),
    value: `${(summary.value.error_rate * 100).toFixed(2)}%`,
    icon: 'exclamationTriangle' as const,
    iconClass: 'bg-amber-50 text-amber-600 dark:bg-amber-500/10 dark:text-amber-300'
  }
])

const requestLogStatsSummary = computed(() => (
  `${formatNumber(summary.value.total_requests)} ${t('admin.requestLogs.totalRequests')} · ${formatNumber(summary.value.error_requests)} ${t('admin.requestLogs.errorRequests')} · ${(summary.value.error_rate * 100).toFixed(2)}%`
))

const activeFilterCount = computed(() => {
  let count = 0
  if (filters.time_range !== '1h' || customStartTime.value || customEndTime.value) count += 1
  if (filters.kind !== 'all') count += 1
  if (filters.billable) count += 1
  if (filters.statusPreset) count += 1
  if (filters.platform) count += 1
  if (filters.model) count += 1
  if (filters.user_id) count += 1
  if (filters.api_key_id) count += 1
  if (filters.account_id) count += 1
  if (filters.group_id) count += 1
  if (filters.request_id) count += 1
  if (filters.session_id) count += 1
  if (filters.q) count += 1
  return count
})

function toRFC3339(localValue: string): string | undefined {
  if (!localValue) return undefined
  const d = new Date(localValue)
  if (Number.isNaN(d.getTime())) return undefined
  return d.toISOString()
}

function buildParams(): AdminRequestLogQueryParams {
  const params: AdminRequestLogQueryParams = {
    page: pagination.page,
    page_size: pagination.page_size,
    sort: filters.sort
  }

  if (filters.time_range === 'custom') {
    const startTime = toRFC3339(customStartTime.value)
    const endTime = toRFC3339(customEndTime.value)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
  } else {
    params.time_range = filters.time_range
  }

  if (filters.kind !== 'all') params.kind = filters.kind
  if (filters.billable) params.billable = filters.billable === 'true'
  if (filters.statusPreset === '!200') {
    params.status_codes_other = true
  } else if (filters.statusPreset === 'custom') {
    const code = filters.customStatusCode.trim()
    if (code) params.status_codes = code
  } else if (filters.statusPreset) {
    params.status_codes = filters.statusPreset
  }
  if (filters.platform) params.platform = filters.platform
  if (filters.model.trim()) params.model = filters.model.trim()
  if (filters.user_id) params.user_id = filters.user_id
  if (filters.api_key_id) params.api_key_id = filters.api_key_id
  if (filters.account_id) params.account_id = filters.account_id
  if (filters.group_id) params.group_id = Number(filters.group_id)
  if (filters.request_id.trim()) params.request_id = filters.request_id.trim()
  if (filters.session_id.trim()) params.session_id = filters.session_id.trim()
  if (filters.q.trim()) params.q = filters.q.trim()

  return params
}

function isCanceled(error: unknown): boolean {
  return !!error && typeof error === 'object' && 'code' in error && (error as { code?: string }).code === 'ERR_CANCELED'
}

function extractErrorMessage(error: unknown): string {
  if (error && typeof error === 'object' && 'message' in error) {
    const message = String((error as { message?: unknown }).message || '')
    if (message) return message
  }
  return t('admin.requestLogs.loadFailed')
}

function isOpsDisabledError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false
  const record = error as { code?: unknown; message?: unknown; status?: unknown }
  return record.code === 'OPS_DISABLED' || record.message === 'Ops monitoring is disabled'
}

async function loadRequestLogs() {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  loading.value = true
  errorMessage.value = ''

  try {
    const result = await adminAPI.requestRecords.list(buildParams(), { signal: controller.signal })
    if (controller.signal.aborted) return
    logs.value = result.items || []
    pagination.total = result.total || 0
    pagination.page = result.page || pagination.page
    pagination.page_size = result.page_size || pagination.page_size
    summary.value = result.summary || {
      total_requests: 0,
      success_requests: 0,
      error_requests: 0,
      error_rate: 0
    }
  } catch (error) {
    if (isCanceled(error)) return
    if (isOpsDisabledError(error)) {
      opsDisabled.value = true
      adminSettingsStore.setOpsMonitoringEnabledLocal(false)
      return
    }
    console.error('[RequestLogsView] Failed to load request logs', error)
    errorMessage.value = extractErrorMessage(error)
    appStore.showError(errorMessage.value)
  } finally {
    if (abortController === controller) {
      abortController = null
      loading.value = false
    }
  }
}

async function loadGroupOptions() {
  try {
    const res = await adminAPI.groups.list(1, 1000)
    groupOptions.value = (res.items || []).map((group: any) => ({
      id: Number(group.id),
      name: group.name || `#${group.id}`
    }))
  } catch (error) {
    console.warn('[RequestLogsView] Failed to load group options', error)
  }
}

function applyFilters() {
  pagination.page = 1
  loadRequestLogs()
}

function refreshData() {
  loadRequestLogs()
}

function filterBySession(sessionId: string) {
  const normalized = sessionId.trim()
  if (!normalized) return
  filters.session_id = normalized
  pagination.page = 1
  loadRequestLogs()
}

function resetFilters() {
  filters.time_range = '1h'
  filters.kind = 'all'
  filters.billable = ''
  filters.statusPreset = ''
  filters.customStatusCode = ''
  filters.platform = ''
  filters.model = ''
  filters.user_id = undefined
  filters.api_key_id = undefined
  filters.account_id = undefined
  filters.group_id = ''
  filters.request_id = ''
  filters.session_id = ''
  filters.q = ''
  filters.sort = 'created_at_desc'
  customStartTime.value = ''
  customEndTime.value = ''
  userKeyword.value = ''
  apiKeyKeyword.value = ''
  accountKeyword.value = ''
  userResults.value = []
  apiKeyResults.value = []
  accountResults.value = []
  applyFilters()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadRequestLogs()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadRequestLogs()
}

function onTimeRangeChange() {
  if (filters.time_range !== 'custom') {
    customStartTime.value = ''
    customEndTime.value = ''
  }
}

function debounceUserSearch() {
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  userSearchTimeout = setTimeout(async () => {
    if (!userKeyword.value.trim()) {
      userResults.value = []
      return
    }
    try {
      userResults.value = await adminAPI.usage.searchUsers(userKeyword.value.trim())
    } catch {
      userResults.value = []
    }
  }, 300)
}

function debounceApiKeySearch() {
  if (apiKeySearchTimeout) clearTimeout(apiKeySearchTimeout)
  apiKeySearchTimeout = setTimeout(async () => {
    try {
      apiKeyResults.value = await adminAPI.usage.searchApiKeys(filters.user_id, apiKeyKeyword.value.trim())
    } catch {
      apiKeyResults.value = []
    }
  }, 300)
}

function debounceAccountSearch() {
  if (accountSearchTimeout) clearTimeout(accountSearchTimeout)
  accountSearchTimeout = setTimeout(async () => {
    if (!accountKeyword.value.trim()) {
      accountResults.value = []
      return
    }
    try {
      const res = await adminAPI.accounts.list(1, 20, { search: accountKeyword.value.trim() })
      accountResults.value = (res.items || []).map((account: any) => ({
        id: Number(account.id),
        name: account.name || `#${account.id}`
      }))
    } catch {
      accountResults.value = []
    }
  }, 300)
}

function getUserDisplayName(user: SimpleUser): string {
  return user.username?.trim() || user.email || `#${user.id}`
}

async function selectUser(user: SimpleUser) {
  filters.user_id = user.id
  userKeyword.value = getUserDisplayName(user)
  showUserDropdown.value = false
  clearApiKey()
  try {
    apiKeyResults.value = await adminAPI.usage.searchApiKeys(user.id, '')
  } catch {
    apiKeyResults.value = []
  }
}

function clearUser() {
  filters.user_id = undefined
  userKeyword.value = ''
  userResults.value = []
  showUserDropdown.value = false
  clearApiKey()
}

function selectApiKey(key: SimpleApiKey) {
  filters.api_key_id = key.id
  apiKeyKeyword.value = key.name || `#${key.id}`
  showApiKeyDropdown.value = false
}

function clearApiKey() {
  filters.api_key_id = undefined
  apiKeyKeyword.value = ''
  apiKeyResults.value = []
  showApiKeyDropdown.value = false
}

function onApiKeyFocus() {
  showApiKeyDropdown.value = true
  if (apiKeyResults.value.length === 0) debounceApiKeySearch()
}

function selectAccount(account: SimpleAccount) {
  filters.account_id = account.id
  accountKeyword.value = account.name || `#${account.id}`
  showAccountDropdown.value = false
}

function clearAccount() {
  filters.account_id = undefined
  accountKeyword.value = ''
  accountResults.value = []
  showAccountDropdown.value = false
}

function onDocumentClick(event: MouseEvent) {
  const target = event.target as Node | null
  if (!target) return
  if (!(userSearchRef.value?.contains(target) ?? false)) showUserDropdown.value = false
  if (!(apiKeySearchRef.value?.contains(target) ?? false)) showApiKeyDropdown.value = false
  if (!(accountSearchRef.value?.contains(target) ?? false)) showAccountDropdown.value = false
  if (!(columnDropdownRef.value?.contains(target) ?? false)) showColumnDropdown.value = false
}

onMounted(async () => {
  loadSavedColumns()
  document.addEventListener('click', onDocumentClick)
  await adminSettingsStore.fetch()
  opsDisabled.value = !adminSettingsStore.opsMonitoringEnabled
  if (opsDisabled.value) return
  void loadGroupOptions()
  await loadRequestLogs()
})

onBeforeUnmount(() => {
  abortController?.abort()
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  if (apiKeySearchTimeout) clearTimeout(apiKeySearchTimeout)
  if (accountSearchTimeout) clearTimeout(accountSearchTimeout)
  document.removeEventListener('click', onDocumentClick)
})
</script>
