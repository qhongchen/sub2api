<template>
  <div class="cch-bento-grid">
    <!-- Today Revenue -->
    <div class="group cch-metric-card cch-accent-emerald">
      <div class="flex items-start gap-3">
        <div class="rounded-lg bg-green-100 p-2 dark:bg-green-900/30">
          <Icon name="dollar" size="md" class="text-green-600 dark:text-green-400" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="text-sm font-medium text-gray-500 group-hover:text-gray-800 dark:text-dark-300 dark:group-hover:text-gray-100">{{ t('payment.admin.todayRevenue') }}</p>
          <p class="mt-3 truncate text-2xl font-bold tracking-tight text-gray-950 dark:text-white">${{ formatMoney(stats.today_amount) }}</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ stats.today_count }} {{ t('payment.admin.orders') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Total Revenue -->
    <div class="group cch-metric-card cch-accent-blue">
      <div class="flex items-start gap-3">
        <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
          <Icon name="creditCard" size="md" class="text-blue-600 dark:text-blue-400" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="text-sm font-medium text-gray-500 group-hover:text-gray-800 dark:text-dark-300 dark:group-hover:text-gray-100">{{ t('payment.admin.totalRevenue') }}</p>
          <p class="mt-3 truncate text-2xl font-bold tracking-tight text-gray-950 dark:text-white">${{ formatMoney(stats.total_amount) }}</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ stats.total_count }} {{ t('payment.admin.orders') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Today Orders -->
    <div class="group cch-metric-card cch-accent-purple">
      <div class="flex items-start gap-3">
        <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
          <Icon name="chart" size="md" class="text-purple-600 dark:text-purple-400" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="text-sm font-medium text-gray-500 group-hover:text-gray-800 dark:text-dark-300 dark:group-hover:text-gray-100">{{ t('payment.admin.todayOrders') }}</p>
          <p class="mt-3 truncate text-2xl font-bold tracking-tight text-gray-950 dark:text-white">{{ stats.today_count }}</p>
        </div>
      </div>
    </div>

    <!-- Average Amount -->
    <div class="group cch-metric-card cch-accent-amber">
      <div class="flex items-start gap-3">
        <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
          <Icon name="chart" size="md" class="text-amber-600 dark:text-amber-400" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="text-sm font-medium text-gray-500 group-hover:text-gray-800 dark:text-dark-300 dark:group-hover:text-gray-100">{{ t('payment.admin.avgAmount') }}</p>
          <p class="mt-3 truncate text-2xl font-bold tracking-tight text-gray-950 dark:text-white">${{ formatMoney(stats.avg_amount) }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { DashboardStats } from '@/types/payment'

const { t } = useI18n()

defineProps<{
  stats: DashboardStats
}>()

function formatMoney(value?: number | null): string {
  return (value ?? 0).toFixed(2)
}
</script>
