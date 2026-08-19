<template>
  <div class="mt-4 min-w-0 overflow-hidden pt-3 border-t border-gray-100 dark:border-dark-700/60">
    <div
      class="flex justify-between text-[10px] font-semibold uppercase tracking-widest text-gray-400 mb-2"
    >
      <span>{{ t('monitorCommon.history60pts', { n: normalizedLength }) }}</span>
      <span class="tabular-nums">{{ t('monitorCommon.nextUpdateIn', { n: countdownSeconds }) }}</span>
    </div>

    <div
      v-if="maintenance"
      class="flex h-20 w-full items-center justify-center rounded border border-dashed border-gray-300 text-[10px] uppercase tracking-widest text-gray-400 dark:border-dark-600"
    >
      {{ t('monitorCommon.maintenancePaused') }}
    </div>
    <div
      v-else
      class="w-full min-w-0 px-1"
      role="img"
      :aria-label="t('monitorCommon.history60pts', { n: normalizedLength })"
    >
      <div class="grid grid-cols-[24px_minmax(0,1fr)] gap-x-1">
        <div class="relative h-14 text-[8px] font-medium tabular-nums text-gray-400">
          <span class="absolute left-0 top-0">30s</span>
          <span class="absolute bottom-0 left-0">0s</span>
        </div>

        <div class="relative h-14 min-w-0">
          <svg
            class="absolute inset-0 h-full w-full"
            viewBox="0 0 100 100"
            preserveAspectRatio="none"
            aria-hidden="true"
          >
            <line
              v-for="y in GRID_LINES"
              :key="y"
              x1="0"
              :y1="y"
              x2="100"
              :y2="y"
              class="stroke-gray-200/80 dark:stroke-dark-600/70"
              stroke-width="1"
              vector-effect="non-scaling-stroke"
            />
            <path
              v-for="segment in lineSegments"
              :key="segment.key"
              :d="segment.path"
              fill="none"
              :class="segment.strokeClass"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              vector-effect="non-scaling-stroke"
            />
          </svg>

          <span
            v-for="point in displayPoints"
            :key="point.key"
            class="absolute inset-y-0"
            :style="point.slotStyle"
            :title="point.title"
          ></span>
        </div>

        <span aria-hidden="true"></span>
        <div class="mt-2 flex h-5 min-w-0 items-end gap-[2px]" aria-hidden="true">
          <span
            v-for="slot in statusSlots"
            :key="slot.key"
            class="min-w-0 flex-1 rounded-sm"
            :class="slot.colorClass"
            :style="{ height: `${slot.heightPct}%` }"
            :title="slot.title"
          ></span>
        </div>
      </div>
    </div>

    <div
      class="ml-7 mt-1 flex justify-between text-[9px] uppercase tracking-widest text-gray-400"
    >
      <span>{{ t('monitorCommon.past') }}</span>
      <span>{{ t('monitorCommon.now') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorStatus, MonitorTimelinePoint } from '@/api/channelMonitor'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'
import {
  STATUS_DEGRADED,
  STATUS_ERROR,
  STATUS_FAILED,
  STATUS_OPERATIONAL,
} from '@/constants/channelMonitor'

const props = withDefaults(defineProps<{
  buckets?: MonitorTimelinePoint[]
  countdownSeconds: number
  length?: number
  maintenance?: boolean
}>(), {
  buckets: () => [],
  length: 60,
  maintenance: false,
})

const { t } = useI18n()
const { statusLabel, formatLatency, formatRelativeTime } = useChannelMonitorFormat()

interface ChartPoint {
  key: string
  index: number
  x: number
  y: number | null
  status: MonitorStatus
  isIncident: boolean
  slotStyle: Record<string, string>
  title: string
}

interface LineSegment {
  key: string
  path: string
  strokeClass: string
}

interface StatusSlot {
  key: string
  colorClass: string
  heightPct: number
  title?: string
}

const MAX_LATENCY_MS = 30_000
const CHART_TOP = 6
const CHART_BOTTOM = 94
const GRID_LINES = [CHART_TOP, (CHART_TOP + CHART_BOTTOM) / 2, CHART_BOTTOM]
const STATUS_BAR_HEIGHT = {
  operational: 100,
  degraded: 65,
  incident: 35,
  empty: 15,
} as const

const normalizedLength = computed(() => {
  const value = Math.floor(props.length)
  return Number.isFinite(value) && value > 0 ? value : 60
})

function normalizedLatency(value: number | null): number | null {
  if (value == null || !Number.isFinite(value) || value < 0) return null
  return value
}

const displayPoints = computed<ChartPoint[]>(() => {
  const real = [...(props.buckets ?? [])]
    .slice(0, normalizedLength.value)
    .reverse()

  const padCount = normalizedLength.value - real.length
  const slotWidth = 100 / normalizedLength.value

  return real.map((point, offset) => {
    const index = padCount + offset
    const x = (index + 0.5) * slotWidth
    const latency = normalizedLatency(point.latency_ms)
    const isIncident =
      point.status === STATUS_FAILED ||
      point.status === STATUS_ERROR ||
      (latency !== null && latency > MAX_LATENCY_MS)
    const y = isIncident
      ? CHART_TOP
      : latency === null
        ? null
        : CHART_BOTTOM - (latency / MAX_LATENCY_MS) * (CHART_BOTTOM - CHART_TOP)
    const latencyText = latency === null ? formatLatency(null) : `${formatLatency(latency)}ms`

    return {
      key: `${index}-${point.checked_at}`,
      index,
      x,
      y,
      status: point.status,
      isIncident,
      slotStyle: {
        left: `${index * slotWidth}%`,
        width: `${slotWidth}%`,
      },
      title: `${formatRelativeTime(point.checked_at)} · ${statusLabel(point.status)} · ${latencyText}`,
    }
  })
})

function strokeClassForPoint(point: ChartPoint): string {
  if (point.isIncident) {
    return 'stroke-red-500 dark:stroke-red-400'
  }
  if (point.status === STATUS_DEGRADED) return 'stroke-amber-500 dark:stroke-amber-400'
  return 'stroke-emerald-500 dark:stroke-emerald-400'
}

function barColorClassForPoint(point: ChartPoint): string {
  if (point.isIncident) return 'bg-red-500 dark:bg-red-400'
  if (point.status === STATUS_DEGRADED) return 'bg-amber-500 dark:bg-amber-400'
  if (point.status === STATUS_OPERATIONAL) return 'bg-emerald-500 dark:bg-emerald-400'
  return 'bg-gray-300 dark:bg-dark-600'
}

function barHeightForPoint(point: ChartPoint): number {
  if (point.isIncident) return STATUS_BAR_HEIGHT.incident
  if (point.status === STATUS_DEGRADED) return STATUS_BAR_HEIGHT.degraded
  if (point.status === STATUS_OPERATIONAL) return STATUS_BAR_HEIGHT.operational
  return STATUS_BAR_HEIGHT.empty
}

const statusSlots = computed<StatusSlot[]>(() => {
  const pointsByIndex = new Map(displayPoints.value.map((point) => [point.index, point]))

  return Array.from({ length: normalizedLength.value }, (_, index) => {
    const point = pointsByIndex.get(index)
    if (!point) {
      return {
        key: `empty-${index}`,
        colorClass: 'bg-gray-200 dark:bg-dark-600',
        heightPct: STATUS_BAR_HEIGHT.empty,
      }
    }

    return {
      key: point.key,
      colorClass: barColorClassForPoint(point),
      heightPct: barHeightForPoint(point),
      title: point.title,
    }
  })
})

const lineSegments = computed<LineSegment[]>(() => {
  const segments: LineSegment[] = []

  for (let index = 1; index < displayPoints.value.length; index += 1) {
    const previous = displayPoints.value[index - 1]
    const current = displayPoints.value[index]
    if (
      previous.index + 1 !== current.index ||
      previous.y === null ||
      current.y === null
    ) {
      continue
    }

    const midpointX = (previous.x + current.x) / 2
    const midpointY = (previous.y + current.y) / 2

    segments.push(
      {
        key: `${previous.index}-${current.index}-previous`,
        path: `M ${previous.x.toFixed(3)} ${previous.y.toFixed(3)} L ${midpointX.toFixed(3)} ${midpointY.toFixed(3)}`,
        strokeClass: strokeClassForPoint(previous),
      },
      {
        key: `${previous.index}-${current.index}-current`,
        path: `M ${midpointX.toFixed(3)} ${midpointY.toFixed(3)} L ${current.x.toFixed(3)} ${current.y.toFixed(3)}`,
        strokeClass: strokeClassForPoint(current),
      },
    )
  }

  return segments
})
</script>
