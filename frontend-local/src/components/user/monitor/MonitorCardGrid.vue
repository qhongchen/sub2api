<template>
  <div>
    <div
      v-if="loading && items.length === 0"
      class="grid gap-5 grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4"
    >
      <div
        v-for="i in 6"
        :key="i"
        class="cch-panel-card min-h-[280px] p-5 animate-pulse"
      >
        <div class="flex items-start gap-3">
          <div class="w-9 h-9 rounded-xl bg-gray-200 dark:bg-dark-700"></div>
          <div class="flex-1 space-y-2">
            <div class="h-4 w-2/3 rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="h-3 w-1/2 rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
          <div class="h-6 w-16 rounded-full bg-gray-200 dark:bg-dark-700"></div>
        </div>
        <div class="mt-5 grid grid-cols-2 gap-2">
          <div class="h-16 rounded-xl bg-gray-100 dark:bg-dark-900/40"></div>
          <div class="h-16 rounded-xl bg-gray-100 dark:bg-dark-900/40"></div>
        </div>
        <div class="mt-6 h-5 w-full rounded bg-gray-100 dark:bg-dark-900/40"></div>
      </div>
    </div>

    <EmptyState
      v-else-if="items.length === 0"
      :title="t('channelStatus.empty.title')"
      :description="t('channelStatus.empty.description')"
    />

    <div v-else class="space-y-8">
      <section
        v-for="group in providerGroups"
        :key="group.provider"
        class="space-y-3"
      >
        <div class="flex items-center justify-between gap-3">
          <div class="flex min-w-0 items-center gap-2">
            <span
              class="grid h-8 w-8 flex-none place-items-center rounded-lg ring-1 ring-black/5 dark:ring-white/10"
              :class="[providerGradient(group.provider), providerTintClass(group.provider)]"
            >
              <ProviderIcon :provider="group.provider" :size="18" />
            </span>
            <h2 class="truncate text-sm font-semibold text-gray-900 dark:text-gray-100">
              {{ providerLabel(group.provider) }}
            </h2>
          </div>
          <span class="rounded-md bg-gray-100 px-2 py-1 text-xs text-gray-500 dark:bg-dark-800 dark:text-gray-400">
            {{ t('channelStatus.providerGroupCount', { n: group.items.length }) }}
          </span>
        </div>

        <div class="grid gap-5 grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
          <MonitorCard
            v-for="item in group.items"
            :key="item.id"
            :item="item"
            :window="window"
            :availability-value="resolveAvailability(item)"
            :countdown-seconds="countdownSeconds"
            @click="emit('cardClick', item)"
          />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Provider, UserMonitorView, UserMonitorDetail } from '@/api/channelMonitor'
import EmptyState from '@/components/common/EmptyState.vue'
import {
  providerGradient,
  useChannelMonitorFormat,
} from '@/composables/useChannelMonitorFormat'
import { PROVIDERS } from '@/constants/channelMonitor'
import MonitorCard from './MonitorCard.vue'
import ProviderIcon from './ProviderIcon.vue'

const PROVIDER_TINT: Record<string, string> = {
  openai: 'text-emerald-600 dark:text-emerald-300',
  anthropic: 'text-orange-600 dark:text-orange-300',
  gemini: 'text-sky-600 dark:text-sky-300',
}

const props = defineProps<{
  items: UserMonitorView[]
  window: '7d' | '15d' | '30d'
  countdownSeconds: number
  loading: boolean
  detailCache: Record<number, UserMonitorDetail>
}>()

const emit = defineEmits<{
  (e: 'cardClick', item: UserMonitorView): void
}>()

const { t } = useI18n()
const { providerLabel } = useChannelMonitorFormat()

const providerGroups = computed(() => {
  const grouped = new Map<Provider, UserMonitorView[]>()
  for (const provider of PROVIDERS) {
    grouped.set(provider, [])
  }
  for (const item of props.items) {
    if (!grouped.has(item.provider)) {
      grouped.set(item.provider, [])
    }
    grouped.get(item.provider)?.push(item)
  }
  return Array.from(grouped.entries())
    .filter(([, groupItems]) => groupItems.length > 0)
    .map(([provider, groupItems]) => ({ provider, items: groupItems }))
})

function providerTintClass(provider: Provider): string {
  return PROVIDER_TINT[provider] ?? 'text-gray-500 dark:text-gray-300'
}

function resolveAvailability(item: UserMonitorView): number | null {
  if (props.window === '7d') {
    return item.availability_7d ?? null
  }
  const detail = props.detailCache[item.id]
  if (!detail) return null
  const primary = detail.models.find(m => m.model === item.primary_model)
  if (!primary) return null
  return props.window === '15d' ? primary.availability_15d ?? null : primary.availability_30d ?? null
}
</script>
