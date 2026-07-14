<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.platformQuota.title')"
    width="wide"
    @close="$emit('close')"
  >
    <div v-if="user" class="space-y-4">
      <div
        v-if="hasActiveSubscription"
        class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-200"
      >
        {{ t('admin.users.platformQuota.subscriptionWarning') }}
      </div>
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ t('admin.users.platformQuota.subtitle', { email: user.email }) }}
      </p>

      <div v-if="loading" class="py-10 text-center text-gray-500">
        {{ t('common.loading') }}
      </div>
      <div v-else class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead>
            <tr class="border-b border-gray-200 text-gray-700 dark:border-dark-700 dark:text-gray-300">
              <th class="px-3 py-2 text-left font-medium">
                {{ t('admin.users.platformQuota.columns.platform') }}
              </th>
              <th class="px-3 py-2 text-left font-medium">
                {{ t('admin.users.platformQuota.columns.daily') }}
              </th>
              <th class="px-3 py-2 text-left font-medium">
                {{ t('admin.users.platformQuota.columns.weekly') }}
              </th>
              <th class="px-3 py-2 text-left font-medium">
                {{ t('admin.users.platformQuota.columns.monthly') }}
              </th>
              <th class="px-3 py-2 text-left font-medium">
                {{ t('admin.users.platformQuota.columns.usage') }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in quotas"
              :key="row.platform"
              class="border-b border-gray-100 dark:border-dark-800"
            >
              <td class="px-3 py-2 font-mono text-gray-900 dark:text-white">
                {{ row.platform }}
              </td>
              <td
                v-for="quotaWindow in QUOTA_WINDOWS"
                :key="quotaWindow"
                class="px-3 py-2"
              >
                <div class="flex items-center gap-1">
                  <input
                    v-model.number="row[`${quotaWindow}_limit_usd`]"
                    type="number"
                    min="0"
                    step="0.01"
                    class="input w-24"
                    :placeholder="t('admin.users.platformQuota.placeholder')"
                  />
                  <button
                    type="button"
                    class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-amber-500 disabled:opacity-50 dark:hover:bg-dark-700"
                    :disabled="!!resetting[`${row.platform}.${quotaWindow}`]"
                    :title="t('admin.users.platformQuota.reset.button')"
                    :aria-label="t('admin.users.platformQuota.reset.button')"
                    @click="onReset(row.platform, quotaWindow)"
                  >
                    <Icon
                      name="refresh"
                      size="xs"
                      :class="{ 'animate-spin': resetting[`${row.platform}.${quotaWindow}`] }"
                    />
                  </button>
                </div>
              </td>
              <td class="px-3 py-2 text-xs text-gray-500 dark:text-gray-400">
                {{ formatUsage(row.daily_usage_usd) }} /
                {{ formatUsage(row.weekly_usage_usd) }} /
                {{ formatUsage(row.monthly_usage_usd) }}
              </td>
            </tr>
          </tbody>
        </table>
        <p class="mt-3 text-xs text-gray-500">
          {{ t('admin.users.platformQuota.hint') }}
        </p>
        <div class="mt-3">
          <button type="button" class="btn btn-secondary text-sm" @click="onClearAll">
            {{ t('admin.users.platformQuota.clearAll') }}
          </button>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="$emit('close')">
          {{ t('admin.users.platformQuota.cancel') }}
        </button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="submitting || loading"
          @click="onSave"
        >
          {{ submitting ? t('admin.users.platformQuota.saving') : t('admin.users.platformQuota.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type {
  PlatformQuotaItem,
  PlatformQuotaPlatform,
  PlatformQuotaWindow
} from '@/api/admin/users'
import type { AdminUser } from '@/types'
import { useAppStore } from '@/stores/app'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

type QuotaInput = number | string | null

interface QuotaRow {
  platform: PlatformQuotaPlatform
  daily_limit_usd: QuotaInput
  weekly_limit_usd: QuotaInput
  monthly_limit_usd: QuotaInput
  daily_usage_usd: number
  weekly_usage_usd: number
  monthly_usage_usd: number
}

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close', 'success'])

const { t } = useI18n()
const appStore = useAppStore()
const PLATFORMS: PlatformQuotaPlatform[] = [
  'anthropic',
  'openai',
  'gemini',
  'antigravity',
  'grok'
]
const QUOTA_WINDOWS: PlatformQuotaWindow[] = ['daily', 'weekly', 'monthly']

const hasActiveSubscription = computed(() =>
  props.user?.subscriptions?.some((subscription) => subscription.status === 'active') ?? false
)
const loading = ref(false)
const submitting = ref(false)
const resetting = reactive<Record<string, boolean>>({})
const quotas = ref<QuotaRow[]>([])

function emptyRow(platform: PlatformQuotaPlatform): QuotaRow {
  return {
    platform,
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    daily_usage_usd: 0,
    weekly_usage_usd: 0,
    monthly_usage_usd: 0
  }
}

function normalize(items: PlatformQuotaItem[]): QuotaRow[] {
  const byPlatform = new Map<PlatformQuotaPlatform, PlatformQuotaItem>()
  for (const item of items) byPlatform.set(item.platform, item)
  return PLATFORMS.map((platform) => {
    const item = byPlatform.get(platform)
    if (!item) return emptyRow(platform)
    return {
      platform,
      daily_limit_usd: item.daily_limit_usd ?? null,
      weekly_limit_usd: item.weekly_limit_usd ?? null,
      monthly_limit_usd: item.monthly_limit_usd ?? null,
      daily_usage_usd: item.daily_usage_usd ?? 0,
      weekly_usage_usd: item.weekly_usage_usd ?? 0,
      monthly_usage_usd: item.monthly_usage_usd ?? 0
    }
  })
}

function formatUsage(value: number): string {
  return Number.isFinite(value) ? value.toFixed(2) : '-'
}

function errorMessage(error: unknown): string | undefined {
  const message = (error as { response?: { data?: { message?: unknown } } })?.response?.data
    ?.message
  return typeof message === 'string' ? message : undefined
}

async function load() {
  if (!props.user) return
  loading.value = true
  try {
    const data = await adminAPI.users.getPlatformQuotas(props.user.id)
    quotas.value = normalize(data.platform_quotas || [])
  } catch {
    appStore.showError(t('admin.users.platformQuota.loadFailed'))
    quotas.value = PLATFORMS.map(emptyRow)
  } finally {
    loading.value = false
  }
}

watch(
  () => props.show,
  (show) => {
    if (show && props.user) void load()
  }
)

function onClearAll() {
  if (!window.confirm(t('admin.users.platformQuota.clearAllConfirm'))) return
  for (const row of quotas.value) {
    row.daily_limit_usd = null
    row.weekly_limit_usd = null
    row.monthly_limit_usd = null
  }
}

function normalizeLimit(value: QuotaInput): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 ? value : null
}

async function onSave() {
  if (!props.user) return

  const invalidFields: string[] = []
  for (const row of quotas.value) {
    for (const quotaWindow of QUOTA_WINDOWS) {
      const value = row[`${quotaWindow}_limit_usd`]
      if (
        value !== '' &&
        value !== null &&
        (typeof value !== 'number' || !Number.isFinite(value) || value < 0)
      ) {
        invalidFields.push(`${row.platform}.${quotaWindow}`)
      }
    }
  }
  if (invalidFields.length > 0) {
    appStore.showError(
      t('admin.users.platformQuota.invalidNumber', { fields: invalidFields.join(', ') })
    )
    return
  }

  submitting.value = true
  try {
    await adminAPI.users.updatePlatformQuotas(
      props.user.id,
      quotas.value.map((row) => ({
        platform: row.platform,
        daily_limit_usd: normalizeLimit(row.daily_limit_usd),
        weekly_limit_usd: normalizeLimit(row.weekly_limit_usd),
        monthly_limit_usd: normalizeLimit(row.monthly_limit_usd)
      }))
    )
    appStore.showSuccess(t('admin.users.platformQuota.updateSuccess'))
    emit('success')
    emit('close')
  } catch (error) {
    appStore.showError(
      errorMessage(error) || t('admin.users.platformQuota.updateFailed')
    )
  } finally {
    submitting.value = false
  }
}

async function onReset(
  platform: PlatformQuotaPlatform,
  quotaWindow: PlatformQuotaWindow
) {
  if (!props.user) return
  const windowLabel = t(
    `admin.users.platformQuota.window${quotaWindow.charAt(0).toUpperCase()}${quotaWindow.slice(1)}`
  )
  if (
    !window.confirm(
      t('admin.users.platformQuota.reset.confirm', { platform, window: windowLabel })
    )
  ) {
    return
  }

  const key = `${platform}.${quotaWindow}`
  resetting[key] = true
  try {
    const data = await adminAPI.users.resetPlatformQuotaWindow(
      props.user.id,
      platform,
      quotaWindow
    )
    quotas.value = normalize(data.platform_quotas || [])
    appStore.showSuccess(
      t('admin.users.platformQuota.reset.success', { platform, window: windowLabel })
    )
  } catch (error) {
    appStore.showError(
      errorMessage(error) || t('admin.users.platformQuota.reset.failed')
    )
  } finally {
    resetting[key] = false
  }
}
</script>
