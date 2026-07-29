<template>
  <div>
    <div
      class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200/70 px-4 py-3 dark:border-white/[0.06]"
    >
      <p class="text-xs text-gray-400 dark:text-dark-500">
        {{ t('admin.usage.tokenRanking.subtitle') }}
      </p>
      <div class="flex items-center gap-3">
        <span v-if="!loading && items.length > 0" class="text-xs text-gray-400 dark:text-dark-500">
          {{ t('admin.usage.tokenRanking.userCount', { count: items.length }) }}
        </span>
        <div class="w-28">
          <Select v-model="limit" :options="limitOptions" @change="load" />
        </div>
      </div>
    </div>

    <div class="overflow-x-auto">
      <table class="w-full min-w-max divide-y divide-gray-200 dark:divide-white/[0.06]">
        <thead class="bg-gray-50/80 dark:bg-white/[0.02]">
          <tr>
            <th
              class="w-16 px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-dark-400"
            >
              #
            </th>
            <th
              class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-dark-400"
            >
              {{ t('admin.usage.tokenRanking.columns.user') }}
            </th>
            <th
              v-for="column in sortableColumns"
              :key="column.key"
              class="cursor-pointer select-none whitespace-nowrap px-4 py-3 text-right text-xs font-medium uppercase transition-colors hover:bg-gray-100 dark:hover:bg-white/[0.04]"
              :class="sortBy === column.key
                ? 'text-primary-600 dark:text-primary-400'
                : 'text-gray-500 dark:text-dark-400'"
              @click="setSort(column.key)"
            >
              {{ t(column.label) }}
              <span v-if="sortBy === column.key" aria-hidden="true">↓</span>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white dark:divide-white/[0.06] dark:bg-dark-900">
          <tr v-if="loading">
            <td :colspan="sortableColumns.length + 2" class="py-12 text-center">
              <LoadingSpinner />
            </td>
          </tr>
          <tr v-else-if="items.length === 0">
            <td
              :colspan="sortableColumns.length + 2"
              class="py-12 text-center text-sm text-gray-400 dark:text-dark-500"
            >
              {{ t('admin.dashboard.noDataAvailable') }}
            </td>
          </tr>
          <tr
            v-for="(item, index) in items"
            v-else
            :key="item.user_id"
            class="cursor-pointer transition-colors hover:bg-gray-50 dark:hover:bg-white/[0.03]"
            :title="t('admin.usage.tokenRanking.rowHint')"
            @click="$emit('select-user', item.user_id, item.email)"
          >
            <td class="px-4 py-3">
              <span
                v-if="index < 3"
                class="inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-semibold"
                :class="RANK_BADGE_CLASSES[index]"
              >
                {{ index + 1 }}
              </span>
              <span v-else class="inline-block w-6 text-center text-sm tabular-nums text-gray-400">
                {{ index + 1 }}
              </span>
            </td>
            <td
              class="max-w-[260px] truncate px-4 py-3 text-sm font-medium text-gray-700 dark:text-dark-100"
              :title="item.email"
            >
              {{ item.email || `User #${item.user_id}` }}
              <span class="ml-1 font-normal text-gray-400 dark:text-dark-500">#{{ item.user_id }}</span>
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-gray-500 dark:text-dark-300">
              {{ item.requests.toLocaleString() }}
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-gray-500 dark:text-dark-300">
              {{ formatTokens(item.input_tokens) }}
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-gray-500 dark:text-dark-300">
              {{ formatTokens(item.output_tokens) }}
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-gray-500 dark:text-dark-300">
              {{ formatTokens(item.cache_tokens) }}
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm font-medium tabular-nums text-gray-950 dark:text-white">
              {{ formatTokens(item.total_tokens) }}
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm font-medium tabular-nums text-emerald-600 dark:text-emerald-400">
              ${{ formatCost(item.actual_cost) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getUserBreakdown, type UserBreakdownParams } from '@/api/admin/dashboard'
import Select from '@/components/common/Select.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { formatCompactNumber, formatCostFixed } from '@/utils/format'
import type { UserBreakdownItem } from '@/types'

const props = defineProps<{
  active: boolean
  startDate: string
  endDate: string
  filters: Record<string, unknown>
  model?: string
}>()

defineEmits<{ (event: 'select-user', userId: number, email: string): void }>()

const { t } = useI18n()

type SortKey = NonNullable<UserBreakdownParams['sort_by']>

const sortableColumns: Array<{ key: SortKey; label: string }> = [
  { key: 'requests', label: 'admin.usage.tokenRanking.columns.requests' },
  { key: 'input_tokens', label: 'admin.usage.tokenRanking.columns.inputTokens' },
  { key: 'output_tokens', label: 'admin.usage.tokenRanking.columns.outputTokens' },
  { key: 'cache_tokens', label: 'admin.usage.tokenRanking.columns.cacheTokens' },
  { key: 'total_tokens', label: 'admin.usage.tokenRanking.columns.totalTokens' },
  { key: 'actual_cost', label: 'admin.usage.tokenRanking.columns.cost' },
]

const limitOptions = [
  { value: 20, label: 'Top 20' },
  { value: 50, label: 'Top 50' },
  { value: 100, label: 'Top 100' },
  { value: 200, label: 'Top 200' },
]

const RANK_BADGE_CLASSES = [
  'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-400',
  'bg-gray-200 text-gray-600 dark:bg-gray-500/20 dark:text-gray-300',
  'bg-orange-100 text-orange-700 dark:bg-orange-500/20 dark:text-orange-400',
]

const items = ref<UserBreakdownItem[]>([])
const loading = ref(false)
const sortBy = ref<SortKey>('total_tokens')
const limit = ref(50)
let requestSequence = 0

const formatTokens = (value: number) => formatCompactNumber(value)
const formatCost = (value: number) => formatCostFixed(value, 4)

const setSort = (key: SortKey) => {
  if (sortBy.value === key) return
  sortBy.value = key
  void load()
}

const load = async () => {
  if (!props.active) return

  const sequence = ++requestSequence
  loading.value = true
  try {
    const params: UserBreakdownParams = {
      ...props.filters,
      start_date: props.startDate || undefined,
      end_date: props.endDate || undefined,
      sort_by: sortBy.value,
      limit: limit.value,
    }
    if (props.model) params.model = props.model

    const response = await getUserBreakdown(params)
    if (sequence !== requestSequence) return
    items.value = response.users || []
  } catch {
    if (sequence !== requestSequence) return
    items.value = []
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

watch(
  () => [
    props.active,
    props.startDate,
    props.endDate,
    props.model,
    props.filters,
  ],
  () => {
    if (props.active) {
      void load()
      return
    }
    requestSequence += 1
    loading.value = false
  },
  { immediate: true },
)

defineExpose({ reload: load })
</script>
