<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="cch-panel-card overflow-hidden">
        <div class="flex flex-col gap-4 p-5 lg:flex-row lg:items-start lg:justify-between">
          <div class="flex min-w-0 items-start gap-4">
            <div class="hidden h-12 w-12 shrink-0 items-center justify-center rounded-xl border border-primary-200 bg-primary-50 text-primary-600 dark:border-primary-500/20 dark:bg-primary-500/10 dark:text-primary-200 sm:flex">
              <Icon name="dollar" size="md" />
            </div>
            <div class="min-w-0 space-y-2">
              <div class="flex flex-wrap items-center gap-2">
                <h1 class="text-xl font-semibold tracking-tight text-gray-950 dark:text-white">
                  {{ t('payment.admin.dashboardTitle') }}
                </h1>
                <span class="rounded-full border border-gray-200 bg-white px-2.5 py-1 text-xs font-medium text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300">
                  {{ days }}{{ t('payment.admin.daySuffix') }}
                </span>
                <span
                  v-if="stats"
                  class="rounded-full border border-primary-200 bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:border-primary-500/20 dark:bg-primary-500/10 dark:text-primary-200"
                >
                  {{ stats.total_count }} {{ t('payment.admin.orders') }}
                </span>
              </div>
              <p class="max-w-2xl text-sm text-gray-500 dark:text-dark-300">
                {{ t('payment.admin.dashboardDesc') }}
              </p>
            </div>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <div class="cch-segmented">
              <button
                v-for="d in DAYS_OPTIONS"
                :key="d"
                type="button"
                class="cch-segmented-button"
                :class="days === d
                  ? 'cch-segmented-button-active'
                  : 'cch-segmented-button-muted'"
                @click="days = d"
              >
                {{ d }}{{ t('payment.admin.daySuffix') }}
              </button>
            </div>
            <button @click="loadDashboard" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </section>

      <div v-if="loading" class="cch-panel-card flex items-center justify-center py-14">
        <LoadingSpinner />
      </div>
      <template v-else-if="stats">
        <OrderStatsCards :stats="stats" />
        <DailyRevenueChart :data="stats.daily_series || []" :loading="loading" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <div class="cch-panel-card p-5">
            <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.paymentDistribution') }}</h3>
            <div v-if="!stats.payment_methods?.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
            <div v-else class="space-y-3">
              <div v-for="method in stats.payment_methods" :key="method.type" class="flex items-center justify-between rounded-xl border border-gray-100 bg-gray-50/70 px-3 py-2.5 dark:border-white/[0.06] dark:bg-white/[0.03]">
                <div class="flex items-center gap-2">
                  <span :class="['inline-block h-3 w-3 rounded-full', methodColor(method.type)]"></span>
                  <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.methods.' + method.type, method.type) }}</span>
                </div>
                <div class="text-right">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">&yen;{{ method.amount.toFixed(2) }}</span>
                  <span class="ml-2 text-xs text-gray-500 dark:text-gray-400">({{ method.count }})</span>
                </div>
              </div>
            </div>
          </div>
          <div class="cch-panel-card p-5">
            <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.topUsers') }}</h3>
            <div v-if="!stats.top_users?.length" class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('payment.admin.noData') }}</div>
            <div v-else class="space-y-2">
              <div v-for="(user, idx) in stats.top_users" :key="user.user_id" class="flex items-center justify-between rounded-xl border border-gray-100 bg-gray-50/70 px-3 py-2.5 transition-colors hover:border-primary-100 hover:bg-primary-50/40 dark:border-white/[0.06] dark:bg-white/[0.03] dark:hover:border-primary-500/20 dark:hover:bg-primary-500/10">
                <div class="flex items-center gap-3">
                  <span :class="['flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold', rankClass(idx)]">{{ idx + 1 }}</span>
                  <span class="text-sm text-gray-700 dark:text-gray-300">{{ user.email }}</span>
                </div>
                <span class="text-sm font-medium text-gray-900 dark:text-white">&yen;{{ user.amount.toFixed(2) }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { DashboardStats } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatsCards from '@/components/admin/payment/OrderStatsCards.vue'
import DailyRevenueChart from '@/components/admin/payment/DailyRevenueChart.vue'

const { t } = useI18n()
const appStore = useAppStore()

const DAYS_OPTIONS = [7, 30, 90] as const
const days = ref<number>(30)
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)

function methodColor(type: string): string {
  const c: Record<string, string> = {
    alipay: 'bg-blue-500', wxpay: 'bg-green-500',
    alipay_direct: 'bg-blue-400', wxpay_direct: 'bg-green-400',
    stripe: 'bg-purple-500',
  }
  return c[type] || 'bg-gray-400'
}

function rankClass(idx: number): string {
  if (idx === 0) return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
  if (idx === 1) return 'bg-gray-200 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
  if (idx === 2) return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
}

async function loadDashboard() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getDashboard(days.value)
    stats.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

watch(days, () => loadDashboard())
onMounted(() => loadDashboard())
</script>
