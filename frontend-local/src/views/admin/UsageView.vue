<template>
  <AppLayout>
    <div class="space-y-5">
      <section class="grid grid-cols-1 gap-4 md:gap-6 lg:grid-cols-2">
        <ModelDistributionChart
          v-model:source="modelDistributionSource"
          v-model:metric="modelDistributionMetric"
          :model-stats="requestedModelStats"
          :upstream-model-stats="upstreamModelStats"
          :mapping-model-stats="mappingModelStats"
          :loading="modelStatsLoading"
          :show-source-toggle="true"
          :show-metric-toggle="true"
          :start-date="startDate"
          :end-date="endDate"
          :filters="breakdownFilters"
        />
        <GroupDistributionChart
          v-model:metric="groupDistributionMetric"
          :group-stats="groupStats"
          :loading="chartsLoading"
          :show-metric-toggle="true"
          :start-date="startDate"
          :end-date="endDate"
          :filters="breakdownFilters"
        />
        <EndpointDistributionChart
          v-model:source="endpointDistributionSource"
          v-model:metric="endpointDistributionMetric"
          :endpoint-stats="inboundEndpointStats"
          :upstream-endpoint-stats="upstreamEndpointStats"
          :endpoint-path-stats="endpointPathStats"
          :loading="endpointStatsLoading"
          :show-source-toggle="true"
          :show-metric-toggle="true"
          :title="t('usage.endpointDistribution')"
          :start-date="startDate"
          :end-date="endDate"
          :filters="breakdownFilters"
        />
        <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
      </section>

      <section class="space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200/70 bg-white/70 px-3 py-2 text-sm font-medium text-gray-600 shadow-sm transition-colors hover:border-gray-300 hover:text-gray-950 dark:border-white/[0.08] dark:bg-dark-900/60 dark:text-dark-300 dark:hover:text-white"
            @click="filtersOpen = !filtersOpen"
          >
            <Icon name="filter" size="sm" :stroke-width="2" />
            <span>{{ t('usage.filterCriteria') }}</span>
            <span
              v-if="activeFilterCount > 0"
              class="rounded-full bg-primary-50 px-1.5 py-0.5 text-[10px] font-semibold text-primary-600 dark:bg-primary-500/10 dark:text-primary-300"
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

          <div class="ml-auto flex items-center gap-1">
            <div class="relative" ref="columnDropdownRef">
              <button
                type="button"
                class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-800 shadow-sm transition-colors hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-100 dark:hover:bg-white/[0.04]"
                :title="t('admin.users.columnSettings')"
                @click="showColumnDropdown = !showColumnDropdown"
              >
                <Icon name="grid" size="sm" />
                <span>{{ visibleColumnCount }}/{{ allColumns.length }}</span>
              </button>
              <div
                v-if="showColumnDropdown"
                class="absolute right-0 top-full z-50 mt-2 max-h-80 w-56 overflow-y-auto rounded-lg border border-gray-200 bg-white p-1 shadow-lg dark:border-dark-700 dark:bg-dark-800"
              >
                <button
                  v-for="col in toggleableColumns"
                  :key="col.key"
                  type="button"
                  class="flex w-full items-center justify-between rounded-md px-3 py-2 text-left text-sm text-gray-600 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
                  @click="toggleColumn(col.key)"
                >
                  <span>{{ col.label }}</span>
                  <Icon v-if="isColumnVisible(col.key)" name="check" size="sm" class="text-primary-500" :stroke-width="2" />
                </button>
              </div>
            </div>

            <button
              type="button"
              class="consumer-icon-btn"
              :title="t('common.refresh')"
              :disabled="loading"
              @click="refreshData"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
            </button>

            <button
              type="button"
              class="inline-flex h-9 items-center gap-2 rounded-full px-2 text-sm text-gray-500 transition-colors hover:text-gray-950 dark:text-dark-300 dark:hover:text-white"
              :title="t('usage.autoRefresh')"
              @click="toggleAutoRefresh"
            >
              <span
                class="h-1.5 w-1.5 rounded-full"
                :class="autoRefreshEnabled ? 'bg-emerald-500' : 'bg-gray-300 dark:bg-dark-600'"
              />
              <span
                class="relative inline-flex h-6 w-10 items-center rounded-full transition-colors"
                :class="autoRefreshEnabled ? 'bg-orange-500' : 'bg-gray-200 dark:bg-dark-700'"
              >
                <span
                  class="inline-block h-5 w-5 rounded-full bg-white shadow transition-transform"
                  :class="autoRefreshEnabled ? 'translate-x-4' : 'translate-x-0.5'"
                />
              </span>
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
          <div v-if="filtersOpen" class="cch-panel-card p-4 md:p-5">
            <div class="mb-5 flex flex-wrap items-center gap-2">
              <button
                v-for="preset in quickDatePresets"
                :key="preset.key"
                type="button"
                class="inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-sm font-semibold shadow-sm transition-colors"
                :class="activePreset === preset.key
                  ? 'border-orange-500 bg-orange-500 text-white'
                  : 'border-gray-200 bg-white text-gray-800 hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-100 dark:hover:bg-white/[0.04]'"
                @click="applyDatePreset(preset.key)"
              >
                <Icon :name="preset.icon" size="sm" />
                {{ preset.label }}
              </button>
            </div>

            <div class="grid gap-4 lg:grid-cols-2">
              <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]">
                <button
                  type="button"
                  class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                  @click="timeFiltersOpen = !timeFiltersOpen"
                >
                  <span class="flex items-center gap-3">
                    <span class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                      <Icon name="clock" size="sm" />
                    </span>
                    <span>
                      <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('usage.timeRange') }}</span>
                      <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('usage.timeRangeDescription') }}</span>
                    </span>
                  </span>
                  <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': timeFiltersOpen }" />
                </button>
                <div v-if="timeFiltersOpen" class="space-y-3 px-4 pb-4">
                  <div class="grid gap-3 sm:grid-cols-2">
                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.startDate') }}</span>
                      <input v-model="startDate" type="date" class="input" @change="syncDateRange" />
                    </label>
                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.endDate') }}</span>
                      <input v-model="endDate" type="date" class="input" @change="syncDateRange" />
                    </label>
                  </div>
                </div>
              </div>

              <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]">
                <button
                  type="button"
                  class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                  @click="identityFiltersOpen = !identityFiltersOpen"
                >
                  <span class="flex items-center gap-3">
                    <span class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                      <Icon name="user" size="sm" />
                    </span>
                    <span>
                      <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('usage.identityInfo') }}</span>
                      <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('usage.adminIdentityDescription') }}</span>
                    </span>
                  </span>
                  <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': identityFiltersOpen }" />
                </button>
                <div v-if="identityFiltersOpen" class="grid gap-3 px-4 pb-4 sm:grid-cols-2">
                  <label ref="userSearchRef" class="relative block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.usage.user') }}</span>
                    <input
                      v-model="userKeyword"
                      type="text"
                      class="input pr-8"
                      :placeholder="t('admin.usage.searchUserPlaceholder')"
                      @input="debounceUserSearch"
                      @focus="showUserDropdown = true"
                    />
                    <button
                      v-if="filters.user_id"
                      type="button"
                      class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                      @click="clearUser"
                    >
                      x
                    </button>
                    <div
                      v-if="showUserDropdown && (userResults.length > 0 || userKeyword)"
                      class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                    >
                      <button
                        v-for="user in userResults"
                        :key="user.id"
                        type="button"
                        class="flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                        @click="selectUser(user)"
                      >
                        <span class="truncate">{{ getSimpleUserDisplayName(user) }}</span>
                        <span class="text-xs text-gray-400">#{{ user.id }}</span>
                      </button>
                    </div>
                  </label>

                  <label ref="apiKeySearchRef" class="relative block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.apiKeyFilter') }}</span>
                    <input
                      v-model="apiKeyKeyword"
                      type="text"
                      class="input pr-8"
                      :placeholder="t('admin.usage.searchApiKeyPlaceholder')"
                      @input="debounceApiKeySearch"
                      @focus="onApiKeyFocus"
                    />
                    <button
                      v-if="filters.api_key_id"
                      type="button"
                      class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                      @click="clearApiKeyAndEmit"
                    >
                      x
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
                </div>
              </div>

              <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]">
                <button
                  type="button"
                  class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                  @click="requestFiltersOpen = !requestFiltersOpen"
                >
                  <span class="flex items-center gap-3">
                    <span class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                      <Icon name="server" size="sm" />
                    </span>
                    <span>
                      <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('usage.requestParams') }}</span>
                      <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('usage.adminRequestDescription') }}</span>
                    </span>
                  </span>
                  <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': requestFiltersOpen }" />
                </button>
                <div v-if="requestFiltersOpen" class="grid gap-3 px-4 pb-4 sm:grid-cols-2">
                  <label ref="accountSearchRef" class="relative block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.usage.account') }}</span>
                    <input
                      v-model="accountKeyword"
                      type="text"
                      class="input pr-8"
                      :placeholder="t('admin.usage.searchAccountPlaceholder')"
                      @input="debounceAccountSearch"
                      @focus="showAccountDropdown = true"
                    />
                    <button
                      v-if="filters.account_id"
                      type="button"
                      class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                      @click="clearAccount"
                    >
                      x
                    </button>
                    <div
                      v-if="showAccountDropdown && (accountResults.length > 0 || accountKeyword)"
                      class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                    >
                      <button
                        v-for="account in accountResults"
                        :key="account.id"
                        type="button"
                        class="flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                        @click="selectAccount(account)"
                      >
                        <span class="truncate">{{ account.name }}</span>
                        <span class="text-xs text-gray-400">#{{ account.id }}</span>
                      </button>
                    </div>
                  </label>

                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.model') }}</span>
                    <Select v-model="filters.model" :options="modelOptions" searchable />
                  </label>

                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.type') }}</span>
                    <Select v-model="filters.request_type" :options="requestTypeOptions" />
                  </label>

                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.usage.group') }}</span>
                    <Select v-model="filters.group_id" :options="groupOptions" searchable />
                  </label>

                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.usage.billingType') }}</span>
                    <Select v-model="filters.billing_type" :options="billingTypeOptions" />
                  </label>

                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.usage.billingMode') }}</span>
                    <Select v-model="filters.billing_mode" :options="billingModeOptions" />
                  </label>
                </div>
              </div>

            </div>

            <div class="mt-5 flex flex-wrap items-center gap-2">
              <button type="button" class="btn btn-primary" :disabled="loading" @click="applyFilters">
                {{ t('usage.applyFilters') }}
              </button>
              <button type="button" class="btn btn-secondary" @click="resetFilters">
                {{ t('common.reset') }}
              </button>
              <button type="button" class="btn btn-secondary" :disabled="exporting" @click="exportToExcel">
                <Icon name="download" size="sm" :class="{ 'animate-bounce': exporting }" />
                {{ exporting ? t('usage.exporting') : t('usage.exportExcel') }}
              </button>
              <button type="button" class="btn btn-danger" @click="openCleanupDialog">
                {{ t('admin.usage.cleanup.button') }}
              </button>
            </div>
          </div>
        </Transition>
      </section>

      <section class="cch-panel-card overflow-hidden">
        <div class="flex items-center justify-between gap-3 border-b border-gray-200/70 px-4 py-3 dark:border-white/[0.06]">
          <span class="text-sm text-gray-500 dark:text-dark-400">
            {{ t('usage.loadedRecords', { count: usageLogs.length }) }}
          </span>
          <span class="text-xs text-gray-400 dark:text-dark-500">
            {{ usageStatsSummary }}
          </span>
        </div>

        <UsageTable
          :data="usageLogs"
          :loading="loading"
          :columns="visibleColumns"
          shell-class="overflow-hidden"
          :server-side-sort="false"
          :default-sort-key="'created_at'"
          :default-sort-order="'desc'"
        />
        <div v-if="pagination.total > 0" class="border-t border-gray-200/70 p-4 dark:border-white/[0.06]">
          <Pagination
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </section>
    </div>
  </AppLayout>

  <UsageExportProgress
    :show="exportProgress.show"
    :progress="exportProgress.progress"
    :current="exportProgress.current"
    :total="exportProgress.total"
    :estimated-time="exportProgress.estimatedTime"
    @cancel="cancelExport"
  />
  <UsageCleanupDialog
    :show="cleanupDialogVisible"
    :filters="filters"
    :start-date="startDate"
    :end-date="endDate"
    @close="cleanupDialogVisible = false"
  />
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { saveAs } from 'file-saver'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { adminUsageAPI, type AdminUsageStatsResponse, type AdminUsageQueryParams, type SimpleApiKey, type SimpleUser } from '@/api/admin/usage'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { formatReasoningEffort } from '@/utils/format'
import { resolveUsageRequestType, requestTypeToLegacyStream } from '@/utils/usageRequestType'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import UsageExportProgress from '@/components/admin/usage/UsageExportProgress.vue'
import UsageCleanupDialog from '@/components/admin/usage/UsageCleanupDialog.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import GroupDistributionChart from '@/components/charts/GroupDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import EndpointDistributionChart from '@/components/charts/EndpointDistributionChart.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUsageLog, TrendDataPoint, ModelStat, GroupStat, EndpointStat } from '@/types'
import type { Column } from '@/components/common/types'

const { t } = useI18n()
const appStore = useAppStore()
const route = useRoute()

type DistributionMetric = 'tokens' | 'actual_cost'
type EndpointSource = 'inbound' | 'upstream' | 'path'
type ModelDistributionSource = 'requested' | 'upstream' | 'mapping'
type DatePresetKey = 'today' | 'this_week' | 'last_7_days' | 'last_30_days'

interface SimpleAccount {
  id: number
  name: string
}

const usageStats = ref<AdminUsageStatsResponse | null>(null)
const usageLogs = ref<AdminUsageLog[]>([])
const loading = ref(false)
const exporting = ref(false)

const trendData = ref<TrendDataPoint[]>([])
const requestedModelStats = ref<ModelStat[]>([])
const upstreamModelStats = ref<ModelStat[]>([])
const mappingModelStats = ref<ModelStat[]>([])
const groupStats = ref<GroupStat[]>([])
const chartsLoading = ref(false)
const modelStatsLoading = ref(false)
const endpointStatsLoading = ref(false)
const granularity = ref<'day' | 'hour'>('hour')
const modelDistributionMetric = ref<DistributionMetric>('tokens')
const modelDistributionSource = ref<ModelDistributionSource>('requested')
const groupDistributionMetric = ref<DistributionMetric>('tokens')
const endpointDistributionMetric = ref<DistributionMetric>('tokens')
const endpointDistributionSource = ref<EndpointSource>('inbound')
const inboundEndpointStats = ref<EndpointStat[]>([])
const upstreamEndpointStats = ref<EndpointStat[]>([])
const endpointPathStats = ref<EndpointStat[]>([])

const filtersOpen = ref(false)
const timeFiltersOpen = ref(true)
const identityFiltersOpen = ref(true)
const requestFiltersOpen = ref(false)
const showColumnDropdown = ref(false)
const autoRefreshEnabled = ref(false)
const activePreset = ref<DatePresetKey | null>(null)

const userSearchRef = ref<HTMLElement | null>(null)
const apiKeySearchRef = ref<HTMLElement | null>(null)
const accountSearchRef = ref<HTMLElement | null>(null)
const columnDropdownRef = ref<HTMLElement | null>(null)
const userKeyword = ref('')
const apiKeyKeyword = ref('')
const accountKeyword = ref('')
const userResults = ref<SimpleUser[]>([])
const apiKeyResults = ref<SimpleApiKey[]>([])
const accountResults = ref<SimpleAccount[]>([])
const showUserDropdown = ref(false)
const showApiKeyDropdown = ref(false)
const showAccountDropdown = ref(false)

let userSearchTimeout: ReturnType<typeof setTimeout> | null = null
let apiKeySearchTimeout: ReturnType<typeof setTimeout> | null = null
let accountSearchTimeout: ReturnType<typeof setTimeout> | null = null
let abortController: AbortController | null = null
let exportAbortController: AbortController | null = null
let autoRefreshTimer: ReturnType<typeof setInterval> | null = null
let chartReqSeq = 0
let statsReqSeq = 0
let modelStatsReqSeq = 0

const exportProgress = reactive({ show: false, progress: 0, current: 0, total: 0, estimatedTime: '' })
const cleanupDialogVisible = ref(false)
const loadedModelSources = reactive<Record<ModelDistributionSource, boolean>>({
  requested: false,
  upstream: false,
  mapping: false,
})

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const startOfWeek = (date: Date): Date => {
  const start = new Date(date)
  const day = start.getDay() || 7
  start.setDate(start.getDate() - day + 1)
  return start
}

const startDate = ref('')
const endDate = ref('')

const filters = ref<AdminUsageQueryParams>({
  user_id: undefined,
  api_key_id: undefined,
  account_id: undefined,
  group_id: undefined,
  model: undefined,
  request_type: undefined,
  billing_type: null,
  billing_mode: undefined,
  start_date: undefined,
  end_date: undefined,
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})

const sortState = reactive({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc',
})

const modelOptions = ref<SelectOption[]>([{ value: null, label: t('admin.usage.allModels') }])
const groupOptions = ref<SelectOption[]>([{ value: null, label: t('admin.usage.allGroups') }])

const quickDatePresets = computed<Array<{ key: DatePresetKey; label: string; icon: 'calendar' | 'clock' }>>(() => [
  { key: 'today', label: t('usage.quickToday'), icon: 'calendar' },
  { key: 'this_week', label: t('usage.quickThisWeek'), icon: 'calendar' },
  { key: 'last_7_days', label: t('usage.quickLast7Days'), icon: 'clock' },
  { key: 'last_30_days', label: t('usage.quickLast30Days'), icon: 'clock' },
])

const requestTypeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allTypes') },
  { value: 'sync', label: t('usage.sync') },
  { value: 'stream', label: t('usage.stream') },
  { value: 'ws_v2', label: t('usage.ws') },
])

const billingTypeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allBillingTypes') },
  { value: 0, label: t('admin.usage.billingTypeBalance') },
  { value: 1, label: t('admin.usage.billingTypeSubscription') },
])

const billingModeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allBillingModes') },
  { value: 'token', label: t('admin.usage.billingModeToken') },
  { value: 'per_request', label: t('admin.usage.billingModePerRequest') },
  { value: 'image', label: t('admin.usage.billingModeImage') },
])

const breakdownFilters = computed(() => {
  const f: Record<string, any> = {}
  if (filters.value.user_id) f.user_id = filters.value.user_id
  if (filters.value.api_key_id) f.api_key_id = filters.value.api_key_id
  if (filters.value.account_id) f.account_id = filters.value.account_id
  if (filters.value.group_id) f.group_id = filters.value.group_id
  if (filters.value.request_type != null) f.request_type = filters.value.request_type
  if (filters.value.billing_type != null) f.billing_type = filters.value.billing_type
  return f
})

const activeFilterCount = computed(() => {
  let count = 0
  if (filters.value.user_id) count += 1
  if (filters.value.api_key_id) count += 1
  if (filters.value.account_id) count += 1
  if (filters.value.group_id) count += 1
  if (filters.value.model) count += 1
  if (filters.value.request_type) count += 1
  if (filters.value.billing_type != null) count += 1
  if (filters.value.billing_mode) count += 1
  if (filters.value.start_date || filters.value.end_date) count += 1
  return count
})

const usageStatsSummary = computed(() => {
  const requests = usageStats.value?.total_requests || 0
  const tokens = usageStats.value?.total_tokens || 0
  const cost = usageStats.value?.total_actual_cost || 0
  return `${requests.toLocaleString()} ${t('usage.requestsShort')} · ${formatTokens(tokens)} ${t('usage.tokens')} · $${cost.toFixed(4)}`
})

const allColumns = computed<Column[]>(() => [
  { key: 'time', label: t('usage.time') },
  { key: 'user', label: t('admin.usage.user'), sortable: false },
  { key: 'api_key', label: t('usage.apiKeyFilter'), sortable: false },
  { key: 'account', label: t('admin.usage.account'), sortable: false },
  { key: 'request_id', label: t('usage.requestId'), sortable: false },
  { key: 'endpoint', label: t('usage.endpoint'), sortable: false },
  { key: 'model', label: t('usage.billingModel'), sortable: false },
  { key: 'tokens', label: t('usage.tokens'), sortable: false },
  { key: 'cache', label: t('usage.cache'), sortable: false },
  { key: 'cost', label: t('usage.cost'), sortable: false },
  { key: 'performance', label: t('usage.performance'), sortable: false },
  { key: 'user_agent', label: t('usage.userAgent'), sortable: false },
])

const ALWAYS_VISIBLE = ['time', 'user', 'account']
const DEFAULT_HIDDEN_COLUMNS = ['endpoint', 'cache']
const HIDDEN_COLUMNS_KEY = 'admin-usage-hidden-columns-v2'
const hiddenColumns = reactive<Set<string>>(new Set())
const toggleableColumns = computed(() => allColumns.value.filter(col => !ALWAYS_VISIBLE.includes(col.key)))
const visibleColumns = computed(() => allColumns.value.filter(col => ALWAYS_VISIBLE.includes(col.key) || !hiddenColumns.has(col.key)))
const visibleColumnCount = computed(() => visibleColumns.value.length)
const isColumnVisible = (key: string) => !hiddenColumns.has(key)

const getSingleQueryValue = (value: string | null | Array<string | null> | undefined): string | undefined => {
  if (Array.isArray(value)) return value.find((item): item is string => typeof item === 'string' && item.length > 0)
  return typeof value === 'string' && value.length > 0 ? value : undefined
}

const getNumericQueryValue = (value: string | null | Array<string | null> | undefined): number | undefined => {
  const raw = getSingleQueryValue(value)
  if (!raw) return undefined
  const parsed = Number(raw)
  return Number.isFinite(parsed) ? parsed : undefined
}

const getGranularityForRange = (start: string, end: string): 'day' | 'hour' => {
  if (!start || !end) return 'day'
  const startTime = new Date(`${start}T00:00:00`).getTime()
  const endTime = new Date(`${end}T00:00:00`).getTime()
  const daysDiff = Math.ceil((endTime - startTime) / (1000 * 60 * 60 * 24))
  return daysDiff <= 1 ? 'hour' : 'day'
}

const syncDateRange = () => {
  filters.value.start_date = startDate.value || undefined
  filters.value.end_date = endDate.value || undefined
  activePreset.value = null
  granularity.value = getGranularityForRange(startDate.value, endDate.value)
}

const applyDatePreset = (preset: DatePresetKey) => {
  const today = new Date()
  let start = new Date(today)

  if (preset === 'today') {
    start = new Date(today)
  } else if (preset === 'this_week') {
    start = startOfWeek(today)
  } else if (preset === 'last_7_days') {
    start.setDate(today.getDate() - 6)
  } else {
    start.setDate(today.getDate() - 29)
  }

  activePreset.value = preset
  startDate.value = formatLocalDate(start)
  endDate.value = formatLocalDate(today)
  filters.value.start_date = startDate.value
  filters.value.end_date = endDate.value
  granularity.value = getGranularityForRange(startDate.value, endDate.value)
  applyFilters()
}

const applyRouteQueryFilters = () => {
  const queryStartDate = getSingleQueryValue(route.query.start_date)
  const queryEndDate = getSingleQueryValue(route.query.end_date)
  const queryUserId = getNumericQueryValue(route.query.user_id)

  if (queryStartDate) startDate.value = queryStartDate
  if (queryEndDate) endDate.value = queryEndDate

  filters.value = {
    ...filters.value,
    user_id: queryUserId,
    start_date: startDate.value || undefined,
    end_date: endDate.value || undefined,
  }
  granularity.value = getGranularityForRange(startDate.value, endDate.value)
}

const buildUsageListParams = (
  page: number,
  pageSize: number,
  exactTotal: boolean,
): AdminUsageQueryParams => {
  const requestType = filters.value.request_type
  const legacyStream = requestType ? requestTypeToLegacyStream(requestType) : filters.value.stream
  return {
    page,
    page_size: pageSize,
    exact_total: exactTotal,
    ...filters.value,
    stream: legacyStream === null ? undefined : legacyStream,
    sort_by: sortState.sort_by,
    sort_order: sortState.sort_order,
  }
}

const loadLogs = async () => {
  abortController?.abort()
  const c = new AbortController()
  abortController = c
  loading.value = true
  try {
    const res = await adminAPI.usage.list(
      buildUsageListParams(pagination.page, pagination.page_size, false),
      { signal: c.signal },
    )
    if (!c.signal.aborted) {
      usageLogs.value = res.items
      pagination.total = res.total
    }
  } catch (error: any) {
    if (error?.name !== 'AbortError') {
      console.error('Failed to load usage logs:', error)
      appStore.showError(t('admin.usage.failedToLoad'))
    }
  } finally {
    if (abortController === c) loading.value = false
  }
}

const loadStats = async () => {
  const seq = ++statsReqSeq
  endpointStatsLoading.value = true
  try {
    const requestType = filters.value.request_type
    const legacyStream = requestType ? requestTypeToLegacyStream(requestType) : filters.value.stream
    const s = await adminAPI.usage.getStats({
      ...filters.value,
      stream: legacyStream === null ? undefined : legacyStream,
    })
    if (seq !== statsReqSeq) return
    usageStats.value = s
    inboundEndpointStats.value = s.endpoints || []
    upstreamEndpointStats.value = s.upstream_endpoints || []
    endpointPathStats.value = s.endpoint_paths || []
  } catch (error) {
    if (seq !== statsReqSeq) return
    console.error('Failed to load usage stats:', error)
    inboundEndpointStats.value = []
    upstreamEndpointStats.value = []
    endpointPathStats.value = []
  } finally {
    if (seq === statsReqSeq) endpointStatsLoading.value = false
  }
}

const resetModelStatsCache = () => {
  requestedModelStats.value = []
  upstreamModelStats.value = []
  mappingModelStats.value = []
  loadedModelSources.requested = false
  loadedModelSources.upstream = false
  loadedModelSources.mapping = false
}

const loadModelStats = async (source: ModelDistributionSource, force = false) => {
  if (!force && loadedModelSources[source]) return

  const seq = ++modelStatsReqSeq
  modelStatsLoading.value = true
  try {
    const requestType = filters.value.request_type
    const legacyStream = requestType ? requestTypeToLegacyStream(requestType) : filters.value.stream
    const baseParams = {
      start_date: filters.value.start_date || startDate.value || undefined,
      end_date: filters.value.end_date || endDate.value || undefined,
      user_id: filters.value.user_id,
      model: filters.value.model,
      api_key_id: filters.value.api_key_id,
      account_id: filters.value.account_id,
      group_id: filters.value.group_id,
      request_type: requestType,
      stream: legacyStream === null ? undefined : legacyStream,
      billing_type: filters.value.billing_type,
    }

    const response = await adminAPI.dashboard.getModelStats({ ...baseParams, model_source: source })
    if (seq !== modelStatsReqSeq) return

    const models = response.models || []
    if (source === 'requested') requestedModelStats.value = models
    else if (source === 'upstream') upstreamModelStats.value = models
    else mappingModelStats.value = models
    loadedModelSources[source] = true
  } catch (error) {
    if (seq !== modelStatsReqSeq) return
    console.error('Failed to load model stats:', error)
    if (source === 'requested') requestedModelStats.value = []
    else if (source === 'upstream') upstreamModelStats.value = []
    else mappingModelStats.value = []
    loadedModelSources[source] = false
  } finally {
    if (seq === modelStatsReqSeq) modelStatsLoading.value = false
  }
}

const loadChartData = async () => {
  const seq = ++chartReqSeq
  chartsLoading.value = true
  try {
    const requestType = filters.value.request_type
    const legacyStream = requestType ? requestTypeToLegacyStream(requestType) : filters.value.stream
    const snapshot = await adminAPI.dashboard.getSnapshotV2({
      start_date: filters.value.start_date || startDate.value || undefined,
      end_date: filters.value.end_date || endDate.value || undefined,
      granularity: granularity.value,
      user_id: filters.value.user_id,
      model: filters.value.model,
      api_key_id: filters.value.api_key_id,
      account_id: filters.value.account_id,
      group_id: filters.value.group_id,
      request_type: requestType,
      stream: legacyStream === null ? undefined : legacyStream,
      billing_type: filters.value.billing_type,
      include_stats: false,
      include_trend: true,
      include_model_stats: false,
      include_group_stats: true,
      include_users_trend: false,
    })
    if (seq !== chartReqSeq) return
    trendData.value = snapshot.trend || []
    groupStats.value = snapshot.groups || []
  } catch (error) {
    console.error('Failed to load chart data:', error)
  } finally {
    if (seq === chartReqSeq) chartsLoading.value = false
  }
}

const applyFilters = () => {
  pagination.page = 1
  resetModelStatsCache()
  loadLogs()
  loadStats()
  loadModelStats(modelDistributionSource.value, true)
  loadChartData()
}

const refreshData = () => {
  resetModelStatsCache()
  loadLogs()
  loadStats()
  loadModelStats(modelDistributionSource.value, true)
  loadChartData()
}

const resetFilters = () => {
  activePreset.value = null
  startDate.value = ''
  endDate.value = ''
  filters.value = {
    user_id: undefined,
    api_key_id: undefined,
    account_id: undefined,
    group_id: undefined,
    model: undefined,
    request_type: undefined,
    billing_type: null,
    billing_mode: undefined,
    start_date: undefined,
    end_date: undefined,
  }
  userKeyword.value = ''
  apiKeyKeyword.value = ''
  accountKeyword.value = ''
  userResults.value = []
  apiKeyResults.value = []
  accountResults.value = []
  granularity.value = getGranularityForRange(startDate.value, endDate.value)
  applyFilters()
}

const handlePageChange = (p: number) => {
  pagination.page = p
  loadLogs()
}

const handlePageSizeChange = (s: number) => {
  pagination.page_size = s
  pagination.page = 1
  loadLogs()
}

const toggleAutoRefresh = () => {
  autoRefreshEnabled.value = !autoRefreshEnabled.value
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
  if (autoRefreshEnabled.value) {
    autoRefreshTimer = setInterval(() => {
      refreshData()
    }, 5000)
  }
}

const openCleanupDialog = () => {
  cleanupDialogVisible.value = true
}

const cancelExport = () => exportAbortController?.abort()

const getRequestTypeLabel = (log: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const exportToExcel = async () => {
  if (exporting.value) return
  exporting.value = true
  exportProgress.show = true
  const c = new AbortController()
  exportAbortController = c
  try {
    let p = 1
    let total = pagination.total
    let exportedCount = 0
    const XLSX = await import('xlsx')
    const headers = [
      t('usage.time'), t('admin.usage.user'), t('usage.apiKeyFilter'),
      t('admin.usage.account'), t('usage.model'), t('usage.upstreamModel'), t('usage.reasoningEffort'), t('admin.usage.group'),
      t('usage.inboundEndpoint'), t('usage.upstreamEndpoint'),
      t('usage.type'),
      t('admin.usage.inputTokens'), t('admin.usage.outputTokens'),
      t('admin.usage.cacheReadTokens'), t('admin.usage.cacheCreationTokens'),
      t('admin.usage.inputCost'), t('admin.usage.outputCost'),
      t('admin.usage.cacheReadCost'), t('admin.usage.cacheCreationCost'),
      t('usage.rate'), t('usage.accountMultiplier'), t('usage.original'), t('usage.userBilled'), t('usage.accountBilled'),
      t('usage.firstToken'), t('usage.duration'),
      t('admin.usage.requestId'), t('usage.userAgent'), t('admin.usage.ipAddress'),
    ]
    const ws = XLSX.utils.aoa_to_sheet([headers])
    while (true) {
      const res = await adminUsageAPI.list(
        buildUsageListParams(p, 100, true),
        { signal: c.signal },
      )
      if (c.signal.aborted) break
      if (p === 1) {
        total = res.total
        exportProgress.total = total
      }
      const rows = (res.items || []).map((log: AdminUsageLog) => [
        log.created_at, log.user?.email || '', log.api_key?.name || '', log.account?.name || '', log.model,
        log.upstream_model || '', formatReasoningEffort(log.reasoning_effort), log.group?.name || '',
        log.inbound_endpoint || '', log.upstream_endpoint || '', getRequestTypeLabel(log),
        log.input_tokens, log.output_tokens, log.cache_read_tokens, log.cache_creation_tokens,
        log.input_cost?.toFixed(6) || '0.000000', log.output_cost?.toFixed(6) || '0.000000',
        log.cache_read_cost?.toFixed(6) || '0.000000', log.cache_creation_cost?.toFixed(6) || '0.000000',
        log.rate_multiplier?.toPrecision(4) || '1.00', (log.account_rate_multiplier ?? 1).toPrecision(4),
        log.total_cost?.toFixed(6) || '0.000000', log.actual_cost?.toFixed(6) || '0.000000',
        ((log.account_stats_cost ?? log.total_cost) * (log.account_rate_multiplier ?? 1)).toFixed(6), log.first_token_ms ?? '', log.duration_ms,
        log.request_id || '', log.user_agent || '', log.ip_address || '',
      ])
      if (rows.length) XLSX.utils.sheet_add_aoa(ws, rows, { origin: -1 })
      exportedCount += rows.length
      exportProgress.current = exportedCount
      exportProgress.progress = total > 0 ? Math.min(100, Math.round(exportedCount / total * 100)) : 0
      if (exportedCount >= total || res.items.length < 100) break
      p += 1
    }
    if (!c.signal.aborted) {
      const wb = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(wb, ws, 'Usage')
      saveAs(
        new Blob([XLSX.write(wb, { bookType: 'xlsx', type: 'array' })], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }),
        `usage_${filters.value.start_date}_to_${filters.value.end_date}.xlsx`,
      )
      appStore.showSuccess(t('usage.exportSuccess'))
    }
  } catch (error) {
    console.error('Failed to export:', error)
    appStore.showError(t('usage.exportFailed'))
  } finally {
    if (exportAbortController === c) {
      exportAbortController = null
      exporting.value = false
      exportProgress.show = false
    }
  }
}

const toggleColumn = (key: string) => {
  if (hiddenColumns.has(key)) hiddenColumns.delete(key)
  else hiddenColumns.add(key)
  try {
    localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
  } catch (e) {
    console.error('Failed to save columns:', e)
  }
}

const loadSavedColumns = () => {
  try {
    const saved = localStorage.getItem(HIDDEN_COLUMNS_KEY)
    if (saved) {
      (JSON.parse(saved) as string[]).forEach((key) => hiddenColumns.add(key))
    } else {
      DEFAULT_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key))
    }
  } catch {
    DEFAULT_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key))
  }
}

const debounceUserSearch = () => {
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  userSearchTimeout = setTimeout(async () => {
    if (!userKeyword.value) {
      userResults.value = []
      return
    }
    try {
      userResults.value = await adminAPI.usage.searchUsers(userKeyword.value)
    } catch {
      userResults.value = []
    }
  }, 300)
}

const debounceApiKeySearch = () => {
  if (apiKeySearchTimeout) clearTimeout(apiKeySearchTimeout)
  apiKeySearchTimeout = setTimeout(async () => {
    try {
      apiKeyResults.value = await adminAPI.usage.searchApiKeys(
        filters.value.user_id,
        apiKeyKeyword.value || '',
      )
    } catch {
      apiKeyResults.value = []
    }
  }, 300)
}

const debounceAccountSearch = () => {
  if (accountSearchTimeout) clearTimeout(accountSearchTimeout)
  accountSearchTimeout = setTimeout(async () => {
    if (!accountKeyword.value) {
      accountResults.value = []
      return
    }
    try {
      const res = await adminAPI.accounts.list(1, 20, { search: accountKeyword.value })
      accountResults.value = res.items.map((a: any) => ({ id: a.id, name: a.name }))
    } catch {
      accountResults.value = []
    }
  }, 300)
}

const getSimpleUserDisplayName = (user: SimpleUser): string => {
  return user.username?.trim() || user.email
}

const selectUser = async (user: SimpleUser) => {
  userKeyword.value = getSimpleUserDisplayName(user)
  showUserDropdown.value = false
  filters.value.user_id = user.id
  clearApiKey()
  try {
    apiKeyResults.value = await adminAPI.usage.searchApiKeys(user.id, '')
  } catch {
    apiKeyResults.value = []
  }
}

const clearUser = () => {
  userKeyword.value = ''
  userResults.value = []
  showUserDropdown.value = false
  filters.value.user_id = undefined
  clearApiKey()
}

const selectApiKey = (key: SimpleApiKey) => {
  apiKeyKeyword.value = key.name || String(key.id)
  showApiKeyDropdown.value = false
  filters.value.api_key_id = key.id
}

const clearApiKey = () => {
  apiKeyKeyword.value = ''
  apiKeyResults.value = []
  showApiKeyDropdown.value = false
  filters.value.api_key_id = undefined
}

const clearApiKeyAndEmit = () => {
  clearApiKey()
}

const onApiKeyFocus = () => {
  showApiKeyDropdown.value = true
  if (apiKeyResults.value.length === 0) debounceApiKeySearch()
}

const selectAccount = (account: SimpleAccount) => {
  accountKeyword.value = account.name
  showAccountDropdown.value = false
  filters.value.account_id = account.id
}

const clearAccount = () => {
  accountKeyword.value = ''
  accountResults.value = []
  showAccountDropdown.value = false
  filters.value.account_id = undefined
}

const onDocumentClick = (event: MouseEvent) => {
  const target = event.target as Node | null
  if (!target) return
  if (!(userSearchRef.value?.contains(target) ?? false)) showUserDropdown.value = false
  if (!(apiKeySearchRef.value?.contains(target) ?? false)) showApiKeyDropdown.value = false
  if (!(accountSearchRef.value?.contains(target) ?? false)) showAccountDropdown.value = false
  if (!(columnDropdownRef.value?.contains(target) ?? false)) showColumnDropdown.value = false
}

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(2)}K`
  return value.toLocaleString()
}

onMounted(async () => {
  applyRouteQueryFilters()
  loadSavedColumns()
  document.addEventListener('click', onDocumentClick)
  loadLogs()
  loadStats()
  loadModelStats(modelDistributionSource.value, true)
  window.setTimeout(() => {
    void loadChartData()
  }, 120)

  try {
    const [groups, models] = await Promise.all([
      adminAPI.groups.list(1, 1000),
      adminAPI.dashboard.getModelStats({ start_date: startDate.value || undefined, end_date: endDate.value || undefined }),
    ])
    groupOptions.value.push(...groups.items.map((g: any) => ({ value: g.id, label: g.name })))
    const uniqueModels = new Set<string>()
    models.models?.forEach((s: any) => {
      if (s.model) uniqueModels.add(s.model)
    })
    modelOptions.value.push(
      ...Array.from(uniqueModels)
        .sort()
        .map((model) => ({ value: model, label: model })),
    )
  } catch {
    // 筛选选项加载失败不影响主列表使用。
  }
})

onUnmounted(() => {
  abortController?.abort()
  exportAbortController?.abort()
  if (autoRefreshTimer) clearInterval(autoRefreshTimer)
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  if (apiKeySearchTimeout) clearTimeout(apiKeySearchTimeout)
  if (accountSearchTimeout) clearTimeout(accountSearchTimeout)
  document.removeEventListener('click', onDocumentClick)
})

watch(modelDistributionSource, (source) => {
  void loadModelStats(source)
})
</script>
