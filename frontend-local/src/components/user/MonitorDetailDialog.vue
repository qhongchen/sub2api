<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="wide"
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

      <section>
        <h4 class="mb-2 text-sm font-semibold text-gray-900 dark:text-gray-100">
          {{ t('channelStatus.historyTitle', { n: recentHistoryLimit }) }}
        </h4>
        <div v-if="historyItems.length === 0" class="py-6 text-center text-sm text-gray-500">
          {{ t('channelStatus.historyEmpty') }}
        </div>
        <div
          v-else
          class="max-h-80 overflow-auto rounded-lg border border-gray-200 dark:border-dark-700"
        >
          <table class="w-full min-w-[1020px] text-left text-sm">
            <thead class="sticky top-0 border-b border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
              <tr class="text-xs uppercase tracking-wider text-gray-500 dark:text-gray-400">
                <th class="w-[180px] py-2 pl-3 pr-3">{{ t('channelStatus.historyColumns.checkedAt') }}</th>
                <th class="w-[130px] py-2 pr-3">{{ t('channelStatus.historyColumns.model') }}</th>
                <th class="w-[100px] py-2 pr-3">{{ t('channelStatus.historyColumns.status') }}</th>
                <th class="w-[140px] py-2 pr-3">{{ t('channelStatus.historyColumns.latency') }}</th>
                <th class="w-[150px] py-2 pr-3">{{ t('channelStatus.historyColumns.pingLatency') }}</th>
                <th class="min-w-[360px] py-2 pr-3">{{ t('channelStatus.historyColumns.message') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in historyItems"
                :key="item.id"
                class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
              >
                <td class="whitespace-nowrap py-2 pl-3 pr-3 text-xs text-gray-600 dark:text-gray-400">
                  {{ formatDateTime(item.checked_at) || '-' }}
                </td>
                <td class="py-2 pr-3 font-medium text-gray-900 dark:text-gray-100">{{ item.model }}</td>
                <td class="py-2 pr-3">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px]"
                    :class="statusBadgeClass(item.status)"
                  >
                    {{ statusLabel(item.status) }}
                  </span>
                </td>
                <td class="whitespace-nowrap py-2 pr-3 text-gray-700 dark:text-gray-300">
                  {{ formatLatency(item.latency_ms) }}
                </td>
                <td class="whitespace-nowrap py-2 pr-3 text-gray-700 dark:text-gray-300">
                  {{ formatLatency(item.ping_latency_ms) }}
                </td>
                <td class="py-2 pr-3 align-top text-gray-600 dark:text-gray-400">
                  <div
                    class="probe-message-preview"
                    @mouseenter="showMessageTooltip($event, item.message)"
                    @mouseleave="hideMessageTooltip"
                  >
                    {{ messageText(item.message) }}
                  </div>
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

  <Teleport to="body">
    <div
      v-if="messageTooltip.show"
      role="tooltip"
      class="monitor-message-tooltip"
      :class="{
        'monitor-message-tooltip-top': messageTooltip.placement === 'top',
        'monitor-message-tooltip-bottom': messageTooltip.placement === 'bottom'
      }"
      :style="{ top: `${messageTooltip.top}px`, left: `${messageTooltip.left}px` }"
    >
      {{ messageTooltip.content }}
      <span
        class="monitor-message-tooltip-arrow"
        :class="{
          'monitor-message-tooltip-arrow-top': messageTooltip.placement === 'top',
          'monitor-message-tooltip-arrow-bottom': messageTooltip.placement === 'bottom'
        }"
        :style="{ left: `${messageTooltip.arrowLeft}px` }"
      />
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import {
  history as fetchChannelMonitorHistory,
  status as fetchChannelMonitorDetail,
  type MonitorHistoryItem,
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
const { statusLabel, statusBadgeClass, formatLatency, formatPercent } = useChannelMonitorFormat()

const detail = ref<UserMonitorDetail | null>(null)
const historyItems = ref<MonitorHistoryItem[]>([])
const loading = ref(false)

const messageTooltip = reactive({
  show: false,
  content: '',
  top: 0,
  left: 0,
  arrowLeft: 0,
  placement: 'top' as 'top' | 'bottom',
})

async function load(id: number) {
  detail.value = null
  historyItems.value = []
  loading.value = true
  try {
    const [detailRes, historyRes] = await Promise.all([
      fetchChannelMonitorDetail(id),
      fetchChannelMonitorHistory(id, { limit: recentHistoryLimit }),
    ])
    detail.value = detailRes
    historyItems.value = historyRes.items || []
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  } finally {
    loading.value = false
  }
}

function messageText(message: string): string {
  return message || '-'
}

const isTooltipTargetOverflowing = (target: HTMLElement): boolean => {
  return target.scrollWidth > target.clientWidth + 1 || target.scrollHeight > target.clientHeight + 1
}

const estimateTooltipWidth = (content: string): number => {
  const padding = 24
  const estimatedTextWidth = Math.max(48, content.length * 7)
  const maxViewportWidth = Math.max(48, window.innerWidth - padding)
  return Math.min(480, maxViewportWidth, Math.max(48, estimatedTextWidth + padding))
}

const showMessageTooltip = (event: MouseEvent, message: string) => {
  const content = message?.trim() || ''
  if (!content || content === '-') {
    hideMessageTooltip()
    return
  }

  const target = event.currentTarget as HTMLElement | null
  if (!target) return

  if (!isTooltipTargetOverflowing(target)) {
    hideMessageTooltip()
    return
  }

  const rect = target.getBoundingClientRect()
  const padding = 12
  const gap = 8
  const tooltipWidth = estimateTooltipWidth(content)
  const tooltipHeightEstimate = Math.min(200, Math.max(56, Math.ceil(content.length / 40) * 18 + 28))
  const preferredLeft = rect.left + rect.width / 2 - tooltipWidth / 2
  const maxLeft = Math.max(padding, window.innerWidth - tooltipWidth - padding)
  const left = Math.max(padding, Math.min(preferredLeft, maxLeft))
  const hasTopSpace = rect.top >= tooltipHeightEstimate + gap + padding
  const placement = hasTopSpace ? 'top' : 'bottom'
  const top = placement === 'top'
    ? Math.max(padding, rect.top - gap)
    : Math.min(window.innerHeight - padding, rect.bottom + gap)
  const arrowLeft = Math.max(12, Math.min(rect.left + rect.width / 2 - left, tooltipWidth - 12))

  messageTooltip.show = true
  messageTooltip.content = content
  messageTooltip.top = top
  messageTooltip.left = left
  messageTooltip.arrowLeft = arrowLeft
  messageTooltip.placement = placement
}

const hideMessageTooltip = () => {
  messageTooltip.show = false
}

watch(
  () => [props.show, props.monitorId] as const,
  ([show, id]) => {
    if (!show) {
      detail.value = null
      historyItems.value = []
      return
    }
    if (id != null) void load(id)
  },
  { immediate: true },
)
</script>

<style scoped>
.probe-message-preview {
  display: -webkit-box;
  overflow: hidden;
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  cursor: default;
}

.monitor-message-tooltip {
  @apply pointer-events-none fixed z-[99999] max-w-[min(30rem,calc(100vw-1.5rem))] rounded-md bg-gray-900 px-3 py-2.5 text-left text-xs font-normal leading-relaxed text-white shadow-xl ring-1 ring-white/10 dark:bg-gray-700;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
  width: max-content;
  max-height: 12.5rem;
  overflow-y: auto;
}

.monitor-message-tooltip-top {
  transform: translateY(-100%);
}

.monitor-message-tooltip-bottom {
  transform: translateY(0);
}

.monitor-message-tooltip-arrow {
  @apply absolute h-2 w-2 -translate-x-1/2 rotate-45 bg-gray-900 dark:bg-gray-700;
}

.monitor-message-tooltip-arrow-top {
  @apply -bottom-1;
}

.monitor-message-tooltip-arrow-bottom {
  @apply -top-1;
}
</style>
