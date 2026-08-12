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
      <svg
        class="h-full w-full"
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
          :class="segment.strokeClass"
          stroke-width="2"
          stroke-linecap="round"
          vector-effect="non-scaling-stroke"
        />
      </svg>
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
import type { MonitorStatus, MonitorTimelinePoint } from '@/api/channelMonitor'
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

interface ChartPoint {
  index: number
  x: number
  y: number | null
  status: MonitorStatus
}

interface LineSegment {
  key: string
  path: string
  strokeClass: string
}

const GRID_LINES = [16, 44, 72]
const CHART_TOP = 10
const CHART_BOTTOM = 72
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
    const isIncident = point.status === STATUS_FAILED || point.status === STATUS_ERROR
    const y = isIncident
      ? INCIDENT_Y
      : latency === null
        ? null
        : CHART_BOTTOM - (latency / maxLatency) * (CHART_BOTTOM - CHART_TOP)

    return {
      index,
      x,
      y,
      status: point.status,
    }
  })
})

function strokeClassForStatus(status: MonitorStatus): string {
  if (status === STATUS_DEGRADED) return 'stroke-amber-500 dark:stroke-amber-400'
  if (status === STATUS_FAILED || status === STATUS_ERROR) {
    return 'stroke-red-500 dark:stroke-red-400'
  }
  return 'stroke-emerald-500 dark:stroke-emerald-400'
}

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
        strokeClass: strokeClassForStatus(previous.status),
      },
      {
        key: `${previous.index}-${current.index}-current`,
        path: `M ${midpointX.toFixed(3)} ${midpointY.toFixed(3)} L ${current.x.toFixed(3)} ${current.y.toFixed(3)}`,
        strokeClass: strokeClassForStatus(current.status),
      },
    )
  }

  return segments
})
</script>
