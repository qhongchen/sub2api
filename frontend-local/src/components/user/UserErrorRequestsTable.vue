<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="card flex min-h-0 flex-1 flex-col overflow-hidden">
      <DataTable
        :columns="columns"
        :data="rows"
        :loading="loading"
        clickable-rows
        server-side-sort
        default-sort-key="created_at"
        default-sort-order="desc"
        @sort="onSort"
        @rowClick="handleRowClick"
      >
        <template #cell-model="{ row }"><span class="text-sm font-medium text-gray-900 dark:text-white">{{ row.model || '-' }}</span></template>
        <template #cell-key_name="{ row }">
          <div class="text-sm">
            <span class="text-gray-900 dark:text-white">{{ row.key_name || '-' }}</span>
            <span v-if="row.key_deleted" class="ml-1 inline-flex items-center rounded bg-rose-100 px-1 py-px text-[10px] font-medium text-rose-600 ring-1 ring-inset ring-rose-200 dark:bg-rose-500/20 dark:text-rose-400 dark:ring-rose-500/30">{{ t('usage.errors.keyDeleted') }}</span>
          </div>
        </template>
        <template #cell-endpoint="{ row }">
          <div class="max-w-[320px] break-all text-xs text-gray-700 dark:text-gray-300">
            <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('usage.inbound') }}:</span>
            <span class="ml-1">{{ row.inbound_endpoint?.trim() || '-' }}</span>
          </div>
        </template>
        <template #cell-status="{ row }"><span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="statusClass(row.status_code)">{{ row.status_code || '-' }}</span></template>
        <template #cell-category="{ row }"><span class="text-sm text-gray-900 dark:text-white">{{ t('usage.errors.categories.' + row.category) }}</span></template>
        <template #cell-message="{ row }"><span class="block max-w-[280px] truncate text-sm text-gray-600 dark:text-gray-400" :title="row.message">{{ row.message || '-' }}</span></template>
        <template #cell-group="{ row }"><span v-if="row.group_name" class="inline-flex items-center rounded bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200">{{ row.group_name }}</span><span v-else>-</span></template>
        <template #cell-type="{ row }"><span v-if="requestType(row)" class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="requestType(row)!.className">{{ requestType(row)!.label }}</span><span v-else>-</span></template>
        <template #cell-platform="{ row }"><span class="text-sm text-gray-900 dark:text-white">{{ row.platform || '-' }}</span></template>
        <template #cell-client_ip="{ row }"><span class="font-mono text-sm text-gray-600 dark:text-gray-400">{{ row.client_ip || '-' }}</span></template>
        <template #cell-created_at="{ row }"><span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDateTime(row.created_at) }}</span></template>
        <template #cell-user_agent="{ row }"><span class="block max-w-[320px] truncate text-sm text-gray-600 dark:text-gray-400" :title="row.user_agent">{{ row.user_agent || '-' }}</span></template>
        <template #empty><EmptyState :message="t('usage.errors.empty')" /></template>
      </DataTable>
    </div>

    <Pagination v-if="total > 0" :page="page" :page-size="pageSize" :total="total" @update:page="$emit('update:page', $event)" @update:pageSize="$emit('update:pageSize', $event)" />
    <UserErrorDetailModal v-model:show="showDetail" :error-id="selectedId" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Pagination from '@/components/common/Pagination.vue'
import UserErrorDetailModal from '@/components/user/UserErrorDetailModal.vue'
import { formatDateTime } from '@/utils/format'
import type { UserErrorRequest } from '@/types'
import type { Column } from '@/components/common/types'

type UserErrorRow = UserErrorRequest & {
  client_ip?: string
  group_name?: string
  request_type?: number
  stream?: boolean
  user_agent?: string
}

const props = defineProps<{ rows: UserErrorRow[]; total: number; loading: boolean; page: number; pageSize: number; visibleColumnKeys?: string[] }>()
const emit = defineEmits<{ (e: 'update:page', value: number): void; (e: 'update:pageSize', value: number): void; (e: 'sort', sortBy: string, sortOrder: 'asc' | 'desc'): void }>()
const { t } = useI18n()

const allColumns = computed<Column[]>(() => [
  { key: 'key_name', label: t('usage.errors.keyName') },
  { key: 'model', label: t('usage.errors.model'), sortable: true },
  { key: 'endpoint', label: t('usage.errors.endpoint') },
  { key: 'client_ip', label: 'IP' },
  { key: 'group', label: t('admin.usage.group') },
  { key: 'type', label: t('usage.type') },
  { key: 'platform', label: t('usage.errors.platform') },
  { key: 'category', label: t('usage.errors.category') },
  { key: 'status', label: t('usage.errors.status'), sortable: true },
  { key: 'message', label: t('usage.errors.message') },
  { key: 'created_at', label: t('usage.errors.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent') }
])
const columns = computed(() => props.visibleColumnKeys ? allColumns.value.filter((column) => props.visibleColumnKeys!.includes(column.key)) : allColumns.value)

function onSort(key: string, order: 'asc' | 'desc') { emit('sort', key === 'status' ? 'status_code' : key, order) }
function statusClass(code: number) {
  if (code >= 500) return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  if (code === 429) return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200'
  if (code >= 400) return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
  return 'bg-gray-100 text-gray-800 dark:bg-dark-700 dark:text-gray-200'
}
function requestType(row: UserErrorRow) {
  const value = row.request_type ?? (row.stream == null ? 0 : row.stream ? 2 : 1)
  if (value === 3) return { label: t('usage.ws'), className: 'bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-200' }
  if (value === 2) return { label: t('usage.stream'), className: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' }
  if (value === 1) return { label: t('usage.sync'), className: 'bg-gray-100 text-gray-800 dark:bg-dark-700 dark:text-gray-200' }
  return null
}

const showDetail = ref(false)
const selectedId = ref<number | null>(null)
function openDetail(id: number) { selectedId.value = id; showDetail.value = true }
function handleRowClick(row: UserErrorRow) { openDetail(row.id) }
</script>
