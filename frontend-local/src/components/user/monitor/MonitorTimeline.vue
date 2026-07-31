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
      class="flex h-14 w-full items-center justify-center rounded border border-dashed border-gray-300 text-[10px] uppercase tracking-widest text-gray-400 dark:border-dark-600"
    >
      {{ t('monitorCommon.maintenancePaused') }}
    </div>
    <div
      v-else
      class="h-14 w-full min-w-0 px-1"
      role="img"
      :aria-label="t('monitorCommon.history60pts', { n: normalizedLength })"
    >
      <div class="relative h-full w-full">
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
          <line
            x1="0"
            :y1="INCIDENT_Y"
            x2="100"
            :y2="INCIDENT_Y"
            class="stroke-gray-300/80 dark:stroke-dark-600"
            stroke-width="1"
            stroke-dasharray="2 3"
            vector-effect="non-scaling-stroke"
          />
          <path
            v-for="segment in lineSegments"
            :key="segment.key"
            :d="segment.path"
            fill="none"
            class="stroke-cyan-500 dark:stroke-cyan-400"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            vector-effect="non-scaling-stroke"
          />
        </svg>

        <span
          v-for="point in displayPoints"
          :key="point.key"
          class="group absolute inset-y-0"
          :style="point.slotStyle"
          :title="point.title"
        >
          <span
            class="absolute inset-y-1 left-1/2 w-px -translate-x-1/2 bg-gray-200 opacity-0 transition-opacity group-hover:opacity-100 dark:bg-dark-600"
          ></span>
          <span
            class="absolute left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full ring-1 ring-white transition-transform group-hover:scale-150 dark:ring-dark-800"
            :class="point.markerClass"
            :style="{ top: `${point.y}%` }"
          ></span>
        </span>
      </div>
    </div>

    <div
      class="mt-1 flex justify-between text-[9px] uppercase tracking-widest text-gray-400"
    >
      <span>{{ t('monitorCommon.past') }}</span>
      <span>{{ t('monitorCommon.now') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorTimelinePoint } from '@/api/channelMonitor'
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
  y: number
  connectsLine: boolean
  markerClass: string
  slotStyle: Record<string, string>
  title: string
}

interface LineSegment {
  key: string
  path: string
}

const GRID_LINES = [16, 44, 72]
const CHART_TOP = 10
const CHART_BOTTOM = 72
const NO_LATENCY_Y = 82
const INCIDENT_Y = 90

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

  const validLatencies = real
    .filter(
      (point) => point.status === STATUS_OPERATIONAL || point.status === STATUS_DEGRADED
    )
    .map((point) => normalizedLatency(point.latency_ms))
    .filter((latency): latency is number => latency !== null)
  const maxLatency = Math.max(1, ...validLatencies)
  const padCount = normalizedLength.value - real.length
  const slotWidth = 100 / normalizedLength.value

  return real.map((point, offset) => {
    const index = padCount + offset
    const x = (index + 0.5) * slotWidth
    const latency = normalizedLatency(point.latency_ms)
    const connectsLine =
      latency !== null &&
      (point.status === STATUS_OPERATIONAL || point.status === STATUS_DEGRADED)
    const y = connectsLine
      ? CHART_BOTTOM - (latency / maxLatency) * (CHART_BOTTOM - CHART_TOP)
      : point.status === STATUS_FAILED || point.status === STATUS_ERROR
        ? INCIDENT_Y
        : NO_LATENCY_Y
    const latencyText = latency === null ? formatLatency(null) : `${formatLatency(latency)}ms`
    const relative = formatRelativeTime(point.checked_at)
    const label = statusLabel(point.status)
    let markerClass = 'h-1.5 w-1.5 bg-gray-400 dark:bg-gray-500'

    if (point.status === STATUS_OPERATIONAL && connectsLine) {
      markerClass = 'h-1.5 w-1.5 bg-cyan-500 dark:bg-cyan-400'
    } else if (point.status === STATUS_DEGRADED) {
      markerClass = 'h-2 w-2 bg-amber-500 dark:bg-amber-400'
    } else if (point.status === STATUS_FAILED || point.status === STATUS_ERROR) {
      markerClass = 'h-2 w-2 bg-red-500 dark:bg-red-400'
    }

    return {
      key: `${index}-${point.checked_at}`,
      index,
      x,
      y,
      connectsLine,
      markerClass,
      slotStyle: {
        left: `${index * slotWidth}%`,
        width: `${slotWidth}%`,
      },
      title: `${relative} · ${label} · ${latencyText}`,
    }
  })
})

const lineSegments = computed<LineSegment[]>(() => {
  const segments: ChartPoint[][] = []
  let current: ChartPoint[] = []

  for (const point of displayPoints.value) {
    const previous = current[current.length - 1]
    if (!point.connectsLine || (previous && point.index !== previous.index + 1)) {
      if (current.length > 1) segments.push(current)
      current = []
    }
    if (point.connectsLine) current.push(point)
  }
  if (current.length > 1) segments.push(current)

  return segments.map((segment) => ({
    key: `${segment[0].index}-${segment[segment.length - 1].index}`,
    path: segment
      .map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x.toFixed(3)} ${point.y.toFixed(3)}`)
      .join(' '),
  }))
})
</script>
