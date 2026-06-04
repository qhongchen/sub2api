<template>
  <section class="cch-panel-card overflow-hidden">
    <div
      v-if="error"
      class="border-b border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/20 dark:bg-red-500/10 dark:text-red-200"
    >
      <slot name="error" :error="error">
        {{ error }}
      </slot>
    </div>

    <div
      v-if="hasHeader"
      class="flex flex-col gap-2 border-b border-gray-200/70 px-4 py-3 dark:border-white/[0.06] sm:flex-row sm:items-center sm:justify-between"
    >
      <div class="min-w-0">
        <slot name="summary">
          <span v-if="summary" class="text-sm text-gray-500 dark:text-dark-400">
            {{ summary }}
          </span>
        </slot>
      </div>

      <div class="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1">
        <slot name="meta">
          <span
            v-for="item in normalizedSummaryItems"
            :key="item.key"
            class="text-xs text-gray-400 dark:text-dark-500"
          >
            {{ item.label }}
          </span>
        </slot>
      </div>
    </div>

    <slot />

    <div
      v-if="showFooter && $slots.footer"
      class="border-t border-gray-200/70 px-4 py-4 dark:border-white/[0.06]"
    >
      <slot name="footer" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue'

type SummaryItem = string | {
  key?: string | number
  label: string
}

const props = withDefaults(defineProps<{
  error?: string
  summary?: string
  summaryItems?: SummaryItem[]
  showFooter?: boolean
}>(), {
  error: '',
  summary: '',
  summaryItems: () => [],
  showFooter: true
})

const slots = useSlots()

const normalizedSummaryItems = computed(() =>
  props.summaryItems.map((item, index) => {
    if (typeof item === 'string') {
      return { key: item || index, label: item }
    }
    return {
      key: item.key ?? item.label ?? index,
      label: item.label
    }
  })
)

const hasHeader = computed(() =>
  Boolean(props.summary || normalizedSummaryItems.value.length || slots.summary || slots.meta)
)
</script>
