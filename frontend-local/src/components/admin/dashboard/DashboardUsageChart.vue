<template>
  <section class="cch-panel-card overflow-hidden p-0 md:p-1">
    <div class="flex flex-col border-b border-gray-200/70 dark:border-white/[0.06] lg:flex-row lg:items-center lg:justify-between">
      <div class="flex items-center gap-4 px-4 py-3">
        <h2 class="text-sm font-semibold text-gray-950 dark:text-white">
          {{ t('admin.dashboard.usageStatistics') }}
        </h2>
      </div>
      <div class="flex overflow-x-auto border-t border-gray-200/70 dark:border-white/[0.06] lg:border-l lg:border-t-0">
        <button
          v-for="option in timeRangeOptions"
          :key="option.key"
          type="button"
          class="whitespace-nowrap border-l border-gray-200/70 px-3 py-2 text-xs font-medium text-gray-600 transition-colors first:border-l-0 hover:bg-gray-100/70 hover:text-gray-950 data-[active=true]:bg-primary-50 data-[active=true]:text-primary-600 dark:border-white/[0.06] dark:text-dark-300 dark:hover:bg-white/[0.04] dark:hover:text-white dark:data-[active=true]:bg-primary-500/10 dark:data-[active=true]:text-primary-300"
          :data-active="selectedTimeRange === option.key"
          @click="emit('timeRangeChange', option.key)"
        >
          {{ option.label }}
        </button>
      </div>
    </div>

    <div class="grid border-b border-gray-200/70 dark:border-white/[0.06] md:grid-cols-3">
      <button
        v-for="metric in metricOptions"
        :key="metric.key"
        type="button"
        class="flex flex-col items-start gap-1 border-b border-gray-200/70 px-4 py-3 text-left transition-colors last:border-b-0 hover:bg-gray-100/70 data-[active=true]:bg-gray-100/80 dark:border-white/[0.06] dark:hover:bg-white/[0.03] dark:data-[active=true]:bg-white/[0.04] md:border-b-0 md:border-r md:last:border-r-0"
        :data-active="activeMetric === metric.key"
        @click="emit('update:activeMetric', metric.key)"
      >
        <span class="text-[10px] font-medium text-gray-500 dark:text-dark-400">
          {{ metric.label }}
        </span>
        <span class="text-lg font-bold tabular-nums text-gray-950 dark:text-white">
          {{ metric.value }}
        </span>
      </button>
    </div>

    <div class="px-4 py-4">
      <div v-if="loading" class="flex h-[320px] items-center justify-center">
        <LoadingSpinner />
      </div>
      <div v-else-if="chartData" class="h-[320px] w-full">
        <Line :data="chartData" :options="chartOptions" />
      </div>
      <div
        v-else
        class="flex h-[320px] items-center justify-center text-sm text-gray-500 dark:text-dark-400"
      >
        {{ t('admin.dashboard.noDataAvailable') }}
      </div>
    </div>

    <div class="relative min-h-[42px] px-4 pb-4">
      <div class="flex flex-wrap justify-center gap-2">
        <button
          v-for="item in legendItems"
          :key="item.id"
          type="button"
          class="inline-flex max-w-full items-center gap-2 rounded-md border border-gray-200 bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-700 shadow-sm transition hover:border-gray-300 hover:bg-white data-[disabled=true]:opacity-45 dark:border-white/[0.08] dark:bg-white/[0.04] dark:text-dark-200 dark:hover:border-white/[0.16] dark:hover:bg-white/[0.07]"
          :data-disabled="item.disabled"
          @click="emit('toggleSeries', item.id)"
        >
          <span
            class="h-2 w-2 flex-shrink-0 rounded-full"
            :class="{ 'ring-2 ring-gray-300 ring-offset-1 ring-offset-white dark:ring-dark-400 dark:ring-offset-[#0a0a0c]': item.disabled }"
            :style="{ backgroundColor: item.disabled ? 'transparent' : item.color, border: `1px solid ${item.color}` }"
          />
          <span class="max-w-[132px] truncate">{{ item.label }}</span>
          <span class="tabular-nums text-gray-500 dark:text-dark-400">{{ item.value }}</span>
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import type { ChartData, ChartOptions } from 'chart.js'
import { Line } from 'vue-chartjs'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type {
  ActiveUsageMetricConfig,
  DashboardTimeRange,
  DashboardTimeRangeOption,
  DashboardUsageLegendItem,
  UsageMetricKey,
  UsageMetricOption
} from './types'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
)

defineProps<{
  timeRangeOptions: DashboardTimeRangeOption[]
  selectedTimeRange: DashboardTimeRange
  metricOptions: UsageMetricOption[]
  activeMetric: UsageMetricKey
  activeMetricConfig: ActiveUsageMetricConfig
  legendItems: DashboardUsageLegendItem[]
  loading: boolean
  chartData: ChartData<'line', number[], string> | null
  chartOptions: ChartOptions<'line'>
}>()

const emit = defineEmits<{
  (event: 'timeRangeChange', value: DashboardTimeRange): void
  (event: 'update:activeMetric', value: UsageMetricKey): void
  (event: 'toggleSeries', value: string): void
}>()

const { t } = useI18n()
</script>
