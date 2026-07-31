<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="extra-wide"
    @close="$emit('close')"
  >
    <div v-if="loading" class="py-8 text-center text-sm text-gray-500">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="!detail" class="py-8 text-center text-sm text-gray-500">
      {{ t('channelStatus.detailLoadError') }}
    </div>
    <div v-else class="space-y-5">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-gray-200 dark:border-dark-700">
            <tr class="text-xs uppercase tracking-wider text-gray-500 dark:text-gray-400">
              <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.model') }}</th>
              <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.latestStatus') }}</th>
              <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.latestLatency') }}</th>
              <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability7d') }}</th>
              <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability15d') }}</th>
              <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.availability30d') }}</th>
              <th class="py-2 pr-3">{{ t('channelStatus.detailColumns.avgLatency7d') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="m in detail.models"
              :key="m.model"
              class="border-b border-gray-100 dark:border-dark-800"
            >
              <td class="py-2 pr-3 font-medium text-gray-900 dark:text-gray-100">{{ m.model }}</td>
              <td class="py-2 pr-3">
                <span
                  class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px]"
                  :class="statusBadgeClass(m.latest_status)"
                >
                  {{ statusLabel(m.latest_status) }}
                </span>
              </td>
              <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatLatency(m.latest_latency_ms) }}</td>
              <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_7d) }}</td>
              <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_15d) }}</td>
              <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_30d) }}</td>
              <td class="py-2 pr-3 text-gray-700 dark:text-gray-300">{{ formatLatency(m.avg_latency_7d_ms) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <section
        v-if="authStore.isAdmin"
        class="border-t border-gray-200 pt-4 dark:border-dark-700"
      >
        <h4 class="mb-3 text-sm font-semibold text-gray-900 dark:text-gray-100">
          {{ t('channelStatus.historyTitle', { n: recentHistoryLimit }) }}
        </h4>

        <div v-if="historyLoading" class="py-8 text-center text-sm text-gray-500">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="historyItems.length === 0" class="py-8 text-center text-sm text-gray-500">
          {{ t('channelStatus.historyEmpty') }}
        </div>
        <div
          v-else
          class="max-h-80 overflow-auto rounded-md border border-gray-200 dark:border-dark-700"
        >
          <table class="w-full min-w-[980px] text-left text-xs">
            <thead
              class="sticky top-0 z-10 border-b border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-900"
            >
              <tr class="uppercase tracking-wider text-gray-500 dark:text-gray-400">
                <th class="w-[170px] px-3 py-2.5 font-semibold">
                  {{ t('channelStatus.historyColumns.checkedAt') }}
                </th>
                <th class="w-[150px] px-3 py-2.5 font-semibold">
                  {{ t('channelStatus.historyColumns.model') }}
                </th>
                <th class="w-[100px] px-3 py-2.5 font-semibold">
                  {{ t('channelStatus.historyColumns.status') }}
                </th>
                <th class="w-[130px] px-3 py-2.5 font-semibold">
                  {{ t('channelStatus.historyColumns.latency') }}
                </th>
                <th class="w-[130px] px-3 py-2.5 font-semibold">
                  {{ t('channelStatus.historyColumns.pingLatency') }}
                </th>
                <th class="min-w-[300px] px-3 py-2.5 font-semibold">
                  {{ t('channelStatus.historyColumns.message') }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in historyItems"
                :key="item.id"
                class="border-b border-gray-100 last:border-b-0 hover:bg-gray-50/70 dark:border-dark-800 dark:hover:bg-dark-700/40"
              >
                <td class="whitespace-nowrap px-3 py-2.5 text-gray-600 dark:text-gray-400">
                  {{ formatDateTime(item.checked_at) || '-' }}
                </td>
                <td class="px-3 py-2.5 font-medium text-gray-900 dark:text-gray-100">
                  {{ item.model }}
                </td>
                <td class="px-3 py-2.5">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px]"
                    :class="statusBadgeClass(item.status)"
                  >
                    {{ statusLabel(item.status) }}
                  </span>
                </td>
                <td class="whitespace-nowrap px-3 py-2.5 tabular-nums text-gray-700 dark:text-gray-300">
                  {{ formatLatency(item.latency_ms) }}
                </td>
                <td class="whitespace-nowrap px-3 py-2.5 tabular-nums text-gray-700 dark:text-gray-300">
                  {{ formatLatency(item.ping_latency_ms) }}
                </td>
                <td class="max-w-[360px] break-words px-3 py-2.5 leading-5 text-gray-600 dark:text-gray-400">
                  {{ item.message || '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button @click="$emit('close')" class="btn btn-secondary">
          {{ t('channelStatus.closeDetail') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import type { HistoryItem } from '@/api/admin/channelMonitor'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import {
  status as fetchChannelMonitorDetail,
  type UserMonitorDetail,
} from '@/api/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const recentHistoryLimit = 60

const props = defineProps<{
  show: boolean
  monitorId: number | null
  title: string
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { statusLabel, statusBadgeClass, formatLatency, formatPercent } = useChannelMonitorFormat()

const detail = ref<UserMonitorDetail | null>(null)
const historyItems = ref<HistoryItem[]>([])
const loading = ref(false)
const historyLoading = ref(false)
let loadSequence = 0

async function loadHistory(id: number, sequence: number) {
  historyLoading.value = true
  try {
    const { listHistory } = await import('@/api/admin/channelMonitor')
    const response = await listHistory(id, {
      limit: recentHistoryLimit,
    })
    if (sequence !== loadSequence) return
    historyItems.value = response.items || []
  } catch (err: unknown) {
    if (sequence !== loadSequence) return
    historyItems.value = []
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  } finally {
    if (sequence === loadSequence) historyLoading.value = false
  }
}

async function load(id: number) {
  const sequence = ++loadSequence
  detail.value = null
  historyItems.value = []
  loading.value = true
  const historyPromise = authStore.isAdmin
    ? loadHistory(id, sequence)
    : Promise.resolve()

  try {
    const response = await fetchChannelMonitorDetail(id)
    if (sequence !== loadSequence) return
    detail.value = response
  } catch (err: unknown) {
    if (sequence !== loadSequence) return
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  } finally {
    if (sequence === loadSequence) loading.value = false
  }

  await historyPromise
}

watch(
  () => [props.show, props.monitorId] as const,
  ([show, id]) => {
    if (!show) {
      loadSequence += 1
      detail.value = null
      historyItems.value = []
      loading.value = false
      historyLoading.value = false
      return
    }
    if (id != null) void load(id)
  },
  { immediate: true },
)
</script>
