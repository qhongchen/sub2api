<template>
  <section class="space-y-3">
    <div class="cch-bento-grid">
      <div
        v-for="card in coreCards"
        :key="card.key"
        class="group cch-metric-card"
        :class="card.accentClass"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="flex min-w-0 items-center gap-3">
            <div
              class="flex-shrink-0 rounded-lg p-2 transition-transform duration-300 group-hover:scale-105"
              :class="card.iconClass"
            >
              <Icon :name="card.icon" size="sm" :class="card.iconTextClass" :stroke-width="2" />
            </div>
            <p class="truncate text-sm font-medium text-gray-500 transition-colors group-hover:text-gray-800 dark:text-dark-300 dark:group-hover:text-gray-100">
              {{ card.title }}
            </p>
          </div>
          <span
            v-if="card.badge"
            class="rounded-full border border-gray-200/70 bg-gray-50/80 px-2 py-0.5 text-[11px] font-medium text-gray-500 dark:border-white/10 dark:bg-white/[0.04] dark:text-dark-300"
          >
            {{ card.badge }}
          </span>
        </div>

        <div class="mt-3">
          <h3
            class="truncate text-2xl font-bold tracking-tight text-gray-950 transition-opacity duration-200 dark:text-white md:text-3xl"
            :class="card.valueClass"
          >
            {{ card.value }}
          </h3>
          <p v-if="card.subtitle" class="mt-1 truncate text-xs text-gray-500 dark:text-dark-400">
            {{ card.subtitle }}
          </p>
        </div>

        <div v-if="card.comparisons.length" class="relative z-10 mt-auto flex flex-wrap gap-x-4 gap-y-1 pt-3">
          <div
            v-for="comparison in card.comparisons"
            :key="comparison.label"
            class="flex items-center gap-1.5"
          >
            <span class="text-xs font-medium" :class="comparison.valueClass">
              {{ comparison.value }}
            </span>
            <span class="text-[10px] text-gray-500 dark:text-dark-400">
              {{ comparison.label }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="foldedCards.length" class="flex justify-end">
      <button
        type="button"
        class="inline-flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:border-gray-300 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300 dark:hover:border-dark-600 dark:hover:bg-dark-700"
        @click="emit('update:showMore', !showMore)"
      >
        <span>{{ showMore ? t('common.collapse') : t('admin.dashboard.moreMetrics') }}</span>
        <Icon
          :name="showMore ? 'chevronUp' : 'chevronDown'"
          size="xs"
          :stroke-width="2"
        />
      </button>
    </div>

    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="-translate-y-1 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="-translate-y-1 opacity-0"
    >
      <div v-if="showMore" class="cch-bento-grid">
        <div
          v-for="card in foldedCards"
          :key="card.key"
          class="group cch-metric-card"
          :class="card.accentClass"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="flex min-w-0 items-center gap-3">
              <div
                class="flex-shrink-0 rounded-lg p-2 transition-transform duration-300 group-hover:scale-105"
                :class="card.iconClass"
              >
                <Icon :name="card.icon" size="sm" :class="card.iconTextClass" :stroke-width="2" />
              </div>
              <p class="truncate text-sm font-medium text-gray-500 transition-colors group-hover:text-gray-800 dark:text-dark-300 dark:group-hover:text-gray-100">
                {{ card.title }}
              </p>
            </div>
            <span
              v-if="card.badge"
              class="rounded-full border border-gray-200/70 bg-gray-50/80 px-2 py-0.5 text-[11px] font-medium text-gray-500 dark:border-white/10 dark:bg-white/[0.04] dark:text-dark-300"
            >
              {{ card.badge }}
            </span>
          </div>

          <div class="mt-3">
            <h3 class="truncate text-2xl font-bold tracking-tight text-gray-950 dark:text-white md:text-3xl" :class="card.valueClass">
              {{ card.value }}
            </h3>
            <p v-if="card.subtitle" class="mt-1 truncate text-xs text-gray-500 dark:text-dark-400">
              {{ card.subtitle }}
            </p>
          </div>

          <div v-if="card.comparisons.length" class="relative z-10 mt-auto flex flex-wrap gap-x-4 gap-y-1 pt-3">
            <div
              v-for="comparison in card.comparisons"
              :key="comparison.label"
              class="flex items-center gap-1.5"
            >
              <span class="text-xs font-medium" :class="comparison.valueClass">
                {{ comparison.value }}
              </span>
              <span class="text-[10px] text-gray-500 dark:text-dark-400">
                {{ comparison.label }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { DashboardMetricCard } from './types'

defineProps<{
  coreCards: DashboardMetricCard[]
  foldedCards: DashboardMetricCard[]
  showMore: boolean
}>()

const emit = defineEmits<{
  (event: 'update:showMore', value: boolean): void
}>()

const { t } = useI18n()
</script>
