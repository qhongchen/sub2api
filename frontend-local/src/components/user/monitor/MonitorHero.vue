<template>
  <div :class="['monitor-hero', { 'cch-toolbar-card': !flat }]">
    <div
      role="tablist"
      class="cch-segmented text-xs"
    >
      <button
        v-for="opt in windowOptions"
        :key="opt.value"
        type="button"
        role="tab"
        :aria-selected="window === opt.value"
        class="cch-segmented-button"
        :class="window === opt.value
          ? 'cch-segmented-button-active'
          : 'cch-segmented-button-muted'"
        @click="emit('update:window', opt.value)"
      >
        {{ opt.label }}
      </button>
    </div>

    <span
      class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold"
      :class="overallChipClass"
    >
      <span
        class="w-1.5 h-1.5 rounded-full mr-1.5"
        :class="overallDotClass"
      ></span>
      {{ overallLabel }}
    </span>

    <button
      type="button"
      class="consumer-icon-btn"
      :disabled="loading"
      :title="t('common.refresh')"
      @click="emit('refresh')"
    >
      <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
    </button>

    <button
      v-if="autoRefresh"
      type="button"
      class="inline-flex h-9 items-center gap-2 rounded-full px-2 text-sm text-gray-500 transition-colors hover:text-gray-950 dark:text-dark-300 dark:hover:text-white"
      :title="autoRefreshTitle"
      :aria-label="autoRefreshTitle"
      :aria-pressed="autoRefresh.enabled.value"
      @click="autoRefresh.setEnabled(!autoRefresh.enabled.value)"
    >
      <span
        class="h-1.5 w-1.5 rounded-full"
        :class="autoRefresh.enabled.value ? 'bg-emerald-500' : 'bg-gray-300 dark:bg-dark-600'"
      />
      <span
        class="relative inline-flex h-6 w-10 items-center rounded-full transition-colors"
        :class="autoRefresh.enabled.value ? 'bg-orange-500' : 'bg-gray-200 dark:bg-dark-700'"
      >
        <span
          class="inline-block h-5 w-5 rounded-full bg-white shadow transition-transform"
          :class="autoRefresh.enabled.value ? 'translate-x-4' : 'translate-x-0.5'"
        />
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
export type MonitorWindow = '7d' | '15d' | '30d'
export type OverallStatus = 'operational' | 'degraded'

const props = defineProps<{
  overallStatus: OverallStatus
  window: MonitorWindow
  loading: boolean
  flat?: boolean
  autoRefresh?: {
    enabled: { value: boolean }
    countdown: { value: number }
    setEnabled: (v: boolean) => void
  }
}>()

const emit = defineEmits<{
  (e: 'update:window', value: MonitorWindow): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const windowOptions = computed<{ value: MonitorWindow; label: string }[]>(() => [
  { value: '7d', label: t('channelStatus.windowTab.7d') },
  { value: '15d', label: t('channelStatus.windowTab.15d') },
  { value: '30d', label: t('channelStatus.windowTab.30d') },
])

const overallLabel = computed(() => t(`channelStatus.overall.${props.overallStatus}`))
const autoRefreshTitle = computed(() => {
  if (!props.autoRefresh?.enabled.value) return t('common.autoRefresh.title')
  return t('common.autoRefresh.countdown', { seconds: props.autoRefresh.countdown.value })
})

const overallChipClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'degraded':
    default:
      return 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
  }
})

const overallDotClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'bg-emerald-500 animate-pulse'
    case 'degraded':
    default:
      return 'bg-amber-500 animate-pulse'
  }
})

</script>

<style scoped>
.monitor-hero {
  @apply flex flex-wrap items-center justify-start gap-2 sm:gap-3 lg:justify-end;
}
</style>
