<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { opsAPI } from '@/api/admin/ops'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import type { OpsAlertRuntimeSettings, EmailNotificationConfig, AlertSeverity, OpsAdvancedSettings, OpsMetricThresholds } from '../types'

const { t } = useI18n()
const appStore = useAppStore()

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const loading = ref(false)
const saving = ref(false)
const sendingChannelStatusTest = ref(false)

// 运行时设置
const runtimeSettings = ref<OpsAlertRuntimeSettings | null>(null)
// 邮件通知配置
const emailConfig = ref<EmailNotificationConfig | null>(null)
// 高级设置
const advancedSettings = ref<OpsAdvancedSettings | null>(null)
// 指标阈值配置
const metricThresholds = ref<OpsMetricThresholds>({
  sla_percent_min: 99.5,
  ttft_p99_ms_max: 500,
  request_error_rate_percent_max: 5,
  upstream_error_rate_percent_max: 5
})

// 加载所有配置
async function loadAllSettings() {
  resetRecipientInputs()
  loading.value = true
  try {
    const [runtime, email, advanced, thresholds] = await Promise.all([
      opsAPI.getAlertRuntimeSettings(),
      opsAPI.getEmailNotificationConfig(),
      opsAPI.getAdvancedSettings(),
      opsAPI.getMetricThresholds()
    ])
    runtimeSettings.value = runtime
    emailConfig.value = email
    advancedSettings.value = advanced
    // 如果后端返回了阈值，使用后端的值；否则保持默认值
    if (thresholds && Object.keys(thresholds).length > 0) {
        metricThresholds.value = {
          sla_percent_min: thresholds.sla_percent_min ?? 99.5,
          ttft_p99_ms_max: thresholds.ttft_p99_ms_max ?? 500,
          request_error_rate_percent_max: thresholds.request_error_rate_percent_max ?? 5,
          upstream_error_rate_percent_max: thresholds.upstream_error_rate_percent_max ?? 5
        }
    }
  } catch (err: any) {
    console.error('[OpsSettingsDialog] Failed to load settings', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.settings.loadFailed'))
  } finally {
    loading.value = false
  }
}

// 监听弹窗打开
watch(() => props.show, (show) => {
  if (show) {
    loadAllSettings()
  }
})

// 邮件输入
const alertRecipientInput = ref('')
const reportRecipientInput = ref('')
const channelStatusRecipientInput = ref('')

// 严重级别选项
const severityOptions: Array<{ value: AlertSeverity | ''; label: string }> = [
  { value: '', label: t('admin.ops.email.minSeverityAll') },
  { value: 'critical', label: t('common.critical') },
  { value: 'warning', label: t('common.warning') },
  { value: 'info', label: t('common.info') }
]

// 验证邮箱
function isValidEmailAddress(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

function resetRecipientInputs() {
  alertRecipientInput.value = ''
  reportRecipientInput.value = ''
  channelStatusRecipientInput.value = ''
}

// 将输入框中的邮箱提交到对应收件人列表。
function addRecipient(target: 'alert' | 'report' | 'channel_status'): boolean {
  if (!emailConfig.value) return false
  const raw = (target === 'alert' ? alertRecipientInput.value : target === 'report' ? reportRecipientInput.value : channelStatusRecipientInput.value).trim()
  if (!raw) return true

  if (!isValidEmailAddress(raw)) {
    appStore.showError(t('common.invalidEmail'))
    return false
  }

  const normalized = raw.toLowerCase()
  const list = target === 'alert' ? emailConfig.value.alert.recipients : target === 'report' ? emailConfig.value.report.recipients : emailConfig.value.channel_status.recipients
  if (!list.includes(normalized)) {
    list.push(normalized)
  }
  if (target === 'alert') alertRecipientInput.value = ''
  else if (target === 'report') reportRecipientInput.value = ''
  else channelStatusRecipientInput.value = ''
  return true
}

function addPendingRecipients(): boolean {
  return addRecipient('alert') && addRecipient('report') && addRecipient('channel_status')
}

// 移除收件人
function removeRecipient(target: 'alert' | 'report' | 'channel_status', email: string) {
  if (!emailConfig.value) return
  const list = target === 'alert' ? emailConfig.value.alert.recipients : target === 'report' ? emailConfig.value.report.recipients : emailConfig.value.channel_status.recipients
  const idx = list.indexOf(email)
  if (idx >= 0) list.splice(idx, 1)
}

// 验证
const validation = computed(() => {
  const errors: string[] = []

  // 验证运行时设置
  if (runtimeSettings.value) {
    const evalSeconds = runtimeSettings.value.evaluation_interval_seconds
    if (!Number.isFinite(evalSeconds) || evalSeconds < 1 || evalSeconds > 86400) {
      errors.push(t('admin.ops.runtime.validation.evalIntervalRange'))
    }
  }

  // 邮件配置允许先开启后补充收件人；运行时会在收件人为空时跳过发送。

  // 验证高级设置
  if (advancedSettings.value) {
    const { error_log_retention_days, minute_metrics_retention_days, hourly_metrics_retention_days } = advancedSettings.value.data_retention
    if (error_log_retention_days < 0 || error_log_retention_days > 365) {
      errors.push(t('admin.ops.settings.validation.retentionDaysRange'))
    }
    if (minute_metrics_retention_days < 0 || minute_metrics_retention_days > 365) {
      errors.push(t('admin.ops.settings.validation.retentionDaysRange'))
    }
    if (hourly_metrics_retention_days < 0 || hourly_metrics_retention_days > 365) {
      errors.push(t('admin.ops.settings.validation.retentionDaysRange'))
    }
  }

  // 验证指标阈值
  if (metricThresholds.value.sla_percent_min != null && (metricThresholds.value.sla_percent_min < 0 || metricThresholds.value.sla_percent_min > 100)) {
    errors.push(t('admin.ops.settings.validation.slaMinPercentRange'))
  }
  if (metricThresholds.value.ttft_p99_ms_max != null && metricThresholds.value.ttft_p99_ms_max < 0) {
    errors.push(t('admin.ops.settings.validation.ttftP99MaxRange'))
  }
  if (metricThresholds.value.request_error_rate_percent_max != null && (metricThresholds.value.request_error_rate_percent_max < 0 || metricThresholds.value.request_error_rate_percent_max > 100)) {
    errors.push(t('admin.ops.settings.validation.requestErrorRateMaxRange'))
  }
  if (metricThresholds.value.upstream_error_rate_percent_max != null && (metricThresholds.value.upstream_error_rate_percent_max < 0 || metricThresholds.value.upstream_error_rate_percent_max > 100)) {
    errors.push(t('admin.ops.settings.validation.upstreamErrorRateMaxRange'))
  }

  return { valid: errors.length === 0, errors }
})

// 保存所有配置
async function saveAllSettings() {
  if (!validation.value.valid) {
    appStore.showError(validation.value.errors[0])
    return
  }
  if (!addPendingRecipients()) return

  saving.value = true
  try {
    await Promise.all([
      runtimeSettings.value ? opsAPI.updateAlertRuntimeSettings(runtimeSettings.value) : Promise.resolve(),
      emailConfig.value ? opsAPI.updateEmailNotificationConfig(emailConfig.value) : Promise.resolve(),
      advancedSettings.value ? opsAPI.updateAdvancedSettings(advancedSettings.value) : Promise.resolve(),
      opsAPI.updateMetricThresholds(metricThresholds.value)
    ])
    appStore.showSuccess(t('admin.ops.settings.saveSuccess'))
    emit('saved')
    emit('close')
  } catch (err: any) {
    console.error('[OpsSettingsDialog] Failed to save settings', err)
    appStore.showError(err?.response?.data?.message || err?.response?.data?.detail || t('admin.ops.settings.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function sendChannelStatusTest() {
  if (!emailConfig.value) return
  if (!addRecipient('channel_status')) return
  if (emailConfig.value.channel_status.recipients.length === 0) {
    appStore.showError(t('admin.ops.settings.channelStatusTestRecipientsRequired'))
    return
  }

  sendingChannelStatusTest.value = true
  try {
    // 测试需要读取已保存的收件人，先持久化当前邮件配置。
    emailConfig.value = await opsAPI.updateEmailNotificationConfig(emailConfig.value)
    await opsAPI.sendChannelStatusTestEmail()
    appStore.showSuccess(t('admin.ops.settings.channelStatusTestSent'))
  } catch (err: any) {
    console.error('[OpsSettingsDialog] Failed to send channel status test email', err)
    appStore.showError(err?.message || err?.response?.data?.message || err?.response?.data?.detail || t('admin.ops.settings.channelStatusTestFailed'))
  } finally {
    sendingChannelStatusTest.value = false
  }
}

const hasChannelStatusRecipient = computed(() =>
  (emailConfig.value?.channel_status.recipients.length ?? 0) > 0 || channelStatusRecipientInput.value.trim() !== ''
)
</script>

<template>
  <BaseDialog :show="show" :title="t('admin.ops.settings.title')" width="extra-wide" @close="emit('close')">
    <div v-if="loading" class="py-10 text-center text-sm text-gray-500">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="runtimeSettings && emailConfig && advancedSettings" class="space-y-6">
      <!-- 验证错误 -->
      <div v-if="!validation.valid" class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-800 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-200">
        <div class="font-bold">{{ t('admin.ops.settings.validation.title') }}</div>
        <ul class="mt-1 list-disc space-y-1 pl-4">
          <li v-for="msg in validation.errors" :key="msg">{{ msg }}</li>
        </ul>
      </div>

      <!-- 数据采集频率 -->
      <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-700/50">
        <h4 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.ops.settings.dataCollection') }}</h4>
        <div>
          <label class="input-label">{{ t('admin.ops.settings.evaluationInterval') }}</label>
          <input
            v-model.number="runtimeSettings.evaluation_interval_seconds"
            type="number"
            min="1"
            max="86400"
            class="input"
          />
          <p class="mt-1 text-xs text-gray-500">{{ t('admin.ops.settings.evaluationIntervalHint') }}</p>
        </div>
      </div>

      <!-- 预警配置 -->
      <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-700/50">
        <h4 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.ops.settings.alertConfig') }}</h4>

        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <label class="font-medium text-gray-900 dark:text-white">{{ t('admin.ops.settings.enableAlert') }}</label>
            </div>
            <Toggle v-model="emailConfig.alert.enabled" />
          </div>

          <div v-if="emailConfig.alert.enabled">
            <label class="input-label">{{ t('admin.ops.settings.alertRecipients') }}</label>
            <div class="flex gap-2">
              <input
                v-model="alertRecipientInput"
                type="email"
                class="input"
                :placeholder="t('admin.ops.settings.emailPlaceholder')"
                @keydown.enter.prevent="addRecipient('alert')"
              />
              <button class="btn btn-secondary whitespace-nowrap" type="button" @click="addRecipient('alert')">
                {{ t('common.add') }}
              </button>
            </div>
            <div class="mt-2 flex flex-wrap gap-2">
              <span
                v-for="email in emailConfig.alert.recipients"
                :key="email"
                class="inline-flex items-center gap-2 rounded-full bg-blue-100 px-3 py-1 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
              >
                {{ email }}
                <button type="button" class="text-blue-700/80 hover:text-blue-900" @click="removeRecipient('alert', email)">×</button>
              </span>
            </div>
            <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.ops.settings.recipientsHint') }}
            </p>
          </div>

          <div v-if="emailConfig.alert.enabled">
            <label class="input-label">{{ t('admin.ops.settings.minSeverity') }}</label>
            <Select v-model="emailConfig.alert.min_severity" :options="severityOptions" />
          </div>
        </div>
      </div>

      <!-- 评估报告配置 -->
      <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-700/50">
        <h4 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.ops.settings.reportConfig') }}</h4>

        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <label class="font-medium text-gray-900 dark:text-white">{{ t('admin.ops.settings.enableReport') }}</label>
            </div>
            <Toggle v-model="emailConfig.report.enabled" />
          </div>

          <div v-if="emailConfig.report.enabled">
            <label class="input-label">{{ t('admin.ops.settings.reportRecipients') }}</label>
            <div class="flex gap-2">
              <input
                v-model="reportRecipientInput"
                type="email"
                class="input"
                :placeholder="t('admin.ops.settings.emailPlaceholder')"
                @keydown.enter.prevent="addRecipient('report')"
              />
              <button class="btn btn-secondary whitespace-nowrap" type="button" @click="addRecipient('report')">
                {{ t('common.add') }}
              </button>
            </div>
            <div class="mt-2 flex flex-wrap gap-2">
              <span
                v-for="email in emailConfig.report.recipients"
                :key="email"
                class="inline-flex items-center gap-2 rounded-full bg-blue-100 px-3 py-1 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
              >
                {{ email }}
                <button type="button" class="text-blue-700/80 hover:text-blue-900" @click="removeRecipient('report', email)">×</button>
              </span>
            </div>
            <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.ops.settings.recipientsHint') }}
            </p>
          </div>

          <div v-if="emailConfig.report.enabled" class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div class="flex items-center justify-between">
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.dailySummary') }}</label>
              <Toggle v-model="emailConfig.report.daily_summary_enabled" />
            </div>
            <div v-if="emailConfig.report.daily_summary_enabled">
              <input v-model="emailConfig.report.daily_summary_schedule" type="text" class="input" placeholder="0 9 * * *" />
            </div>
            <div class="flex items-center justify-between">
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.weeklySummary') }}</label>
              <Toggle v-model="emailConfig.report.weekly_summary_enabled" />
            </div>
            <div v-if="emailConfig.report.weekly_summary_enabled">
              <input v-model="emailConfig.report.weekly_summary_schedule" type="text" class="input" placeholder="0 9 * * 1" />
            </div>
          </div>
        </div>
      </div>

      <!-- 渠道状态通知配置 -->
      <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-700/50">
        <h4 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.ops.settings.channelStatusConfig') }}</h4>
        <p class="mb-4 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.ops.settings.channelStatusHint') }}</p>

        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <label class="font-medium text-gray-900 dark:text-white">{{ t('admin.ops.settings.enableChannelStatus') }}</label>
            </div>
            <Toggle v-model="emailConfig.channel_status.enabled" />
          </div>

          <div v-if="emailConfig.channel_status.enabled">
            <label class="input-label">{{ t('admin.ops.settings.channelStatusRecipients') }}</label>
            <div class="flex gap-2">
              <input
                v-model="channelStatusRecipientInput"
                type="email"
                class="input"
                :placeholder="t('admin.ops.settings.emailPlaceholder')"
                @keydown.enter.prevent="addRecipient('channel_status')"
              />
              <button class="btn btn-secondary whitespace-nowrap" type="button" @click="addRecipient('channel_status')">
                {{ t('common.add') }}
              </button>
            </div>
            <div class="mt-2 flex flex-wrap gap-2">
              <span
                v-for="email in emailConfig.channel_status.recipients"
                :key="email"
                class="inline-flex items-center gap-2 rounded-full bg-blue-100 px-3 py-1 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
              >
                {{ email }}
                <button type="button" class="text-blue-700/80 hover:text-blue-900" @click="removeRecipient('channel_status', email)">×</button>
              </span>
            </div>
            <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.ops.settings.recipientsHint') }}
            </p>
          </div>

          <div v-if="emailConfig.channel_status.enabled">
            <label class="input-label">{{ t('admin.ops.settings.consecutiveThreshold') }}</label>
            <input
              v-model.number="emailConfig.channel_status.consecutive_threshold"
              type="number"
              min="1"
              max="100"
              class="input"
            />
            <p class="mt-1 text-xs text-gray-500">{{ t('admin.ops.settings.consecutiveThresholdHint') }}</p>
          </div>

          <div v-if="emailConfig.channel_status.enabled" class="flex items-center justify-between gap-4">
            <p class="min-w-0 flex-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.ops.settings.channelStatusTestHint') }}</p>
            <button
              data-testid="channel-status-test-email"
              type="button"
              class="btn btn-secondary btn-sm shrink-0"
              :disabled="sendingChannelStatusTest || !hasChannelStatusRecipient"
              @click="sendChannelStatusTest"
            >
              {{ sendingChannelStatusTest ? t('admin.ops.settings.channelStatusTestSending') : t('admin.ops.settings.channelStatusTest') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 指标阈值配置 -->
      <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-700/50">
        <h4 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.ops.settings.metricThresholds') }}</h4>
        <p class="mb-4 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.ops.settings.metricThresholdsHint') }}</p>

        <div class="space-y-4">
          <div>
            <label class="input-label">{{ t('admin.ops.settings.slaMinPercent') }}</label>
            <input
              v-model.number="metricThresholds.sla_percent_min"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
            />
            <p class="mt-1 text-xs text-gray-500">{{ t('admin.ops.settings.slaMinPercentHint') }}</p>
          </div>


          <div>
            <label class="input-label">{{ t('admin.ops.settings.ttftP99MaxMs') }}</label>
            <input
              v-model.number="metricThresholds.ttft_p99_ms_max"
              type="number"
              min="0"
              step="50"
              class="input"
            />
            <p class="mt-1 text-xs text-gray-500">{{ t('admin.ops.settings.ttftP99MaxMsHint') }}</p>
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.settings.requestErrorRateMaxPercent') }}</label>
            <input
              v-model.number="metricThresholds.request_error_rate_percent_max"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
            />
            <p class="mt-1 text-xs text-gray-500">{{ t('admin.ops.settings.requestErrorRateMaxPercentHint') }}</p>
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.settings.upstreamErrorRateMaxPercent') }}</label>
            <input
              v-model.number="metricThresholds.upstream_error_rate_percent_max"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
            />
            <p class="mt-1 text-xs text-gray-500">{{ t('admin.ops.settings.upstreamErrorRateMaxPercentHint') }}</p>
          </div>
        </div>
      </div>

      <!-- 高级设置 -->
      <details class="rounded-2xl bg-gray-50 dark:bg-dark-700/50">
        <summary class="cursor-pointer p-4 text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('admin.ops.settings.advancedSettings') }}
        </summary>
        <div class="space-y-4 px-4 pb-4">
          <!-- 数据保留策略 -->
          <div class="space-y-3">
            <h5 class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.dataRetention') }}</h5>

            <div class="flex items-center justify-between">
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.enableCleanup') }}</label>
              <Toggle v-model="advancedSettings.data_retention.cleanup_enabled" />
            </div>

            <div v-if="advancedSettings.data_retention.cleanup_enabled">
              <label class="input-label">{{ t('admin.ops.settings.cleanupSchedule') }}</label>
              <input
                v-model="advancedSettings.data_retention.cleanup_schedule"
                type="text"
                class="input"
                placeholder="0 2 * * *"
              />
              <p class="mt-1 text-xs text-gray-500">{{ t('admin.ops.settings.cleanupScheduleHint') }}</p>
            </div>

            <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
              <div>
                <label class="input-label">{{ t('admin.ops.settings.errorLogRetentionDays') }}</label>
                <input
                  v-model.number="advancedSettings.data_retention.error_log_retention_days"
                  type="number"
                  min="0"
                  max="365"
                  class="input"
                />
              </div>
              <div>
                <label class="input-label">{{ t('admin.ops.settings.minuteMetricsRetentionDays') }}</label>
                <input
                  v-model.number="advancedSettings.data_retention.minute_metrics_retention_days"
                  type="number"
                  min="0"
                  max="365"
                  class="input"
                />
              </div>
              <div>
                <label class="input-label">{{ t('admin.ops.settings.hourlyMetricsRetentionDays') }}</label>
                <input
                  v-model.number="advancedSettings.data_retention.hourly_metrics_retention_days"
                  type="number"
                  min="0"
                  max="365"
                  class="input"
                />
              </div>
            </div>
            <p class="text-xs text-gray-500">{{ t('admin.ops.settings.retentionDaysHint') }}</p>
          </div>

          <!-- 预聚合任务 -->
          <div class="space-y-3">
            <h5 class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.aggregation') }}</h5>

            <div class="flex items-center justify-between">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.enableAggregation') }}</label>
                <p class="mt-1 text-xs text-gray-500">{{ t('admin.ops.settings.aggregationHint') }}</p>
              </div>
              <Toggle v-model="advancedSettings.aggregation.aggregation_enabled" />
            </div>
          </div>

          <!-- Error Filtering -->
          <div class="space-y-3">
            <h5 class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.errorFiltering') }}</h5>

            <div class="flex items-center justify-between">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.ignoreCountTokensErrors') }}</label>
                <p class="mt-1 text-xs text-gray-500">
                  {{ t('admin.ops.settings.ignoreCountTokensErrorsHint') }}
                </p>
              </div>
              <Toggle v-model="advancedSettings.ignore_count_tokens_errors" />
            </div>

            <div class="flex items-center justify-between">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.ignoreContextCanceled') }}</label>
                <p class="mt-1 text-xs text-gray-500">
                  {{ t('admin.ops.settings.ignoreContextCanceledHint') }}
                </p>
              </div>
              <Toggle v-model="advancedSettings.ignore_context_canceled" />
            </div>

            <div class="flex items-center justify-between">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.ignoreNoAvailableAccounts') }}</label>
                <p class="mt-1 text-xs text-gray-500">
                  {{ t('admin.ops.settings.ignoreNoAvailableAccountsHint') }}
                </p>
              </div>
              <Toggle v-model="advancedSettings.ignore_no_available_accounts" />
            </div>

            <div class="flex items-center justify-between">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.ignoreInsufficientBalanceErrors') }}</label>
                <p class="mt-1 text-xs text-gray-500">
                  {{ t('admin.ops.settings.ignoreInsufficientBalanceErrorsHint') }}
                </p>
              </div>
              <Toggle v-model="advancedSettings.ignore_insufficient_balance_errors" />
            </div>
          </div>

          <!-- Auto Refresh -->
          <div class="space-y-3">
            <h5 class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.autoRefresh') }}</h5>

            <div class="flex items-center justify-between">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.enableAutoRefresh') }}</label>
                <p class="mt-1 text-xs text-gray-500">
                  {{ t('admin.ops.settings.enableAutoRefreshHint') }}
                </p>
              </div>
              <Toggle v-model="advancedSettings.auto_refresh_enabled" />
            </div>

            <div v-if="advancedSettings.auto_refresh_enabled">
              <label class="input-label">{{ t('admin.ops.settings.refreshInterval') }}</label>
              <Select
                v-model="advancedSettings.auto_refresh_interval_seconds"
                :options="[
                  { value: 15, label: t('admin.ops.settings.refreshInterval15s') },
                  { value: 30, label: t('admin.ops.settings.refreshInterval30s') },
                  { value: 60, label: t('admin.ops.settings.refreshInterval60s') }
                ]"
              />
            </div>
          </div>

          <!-- Dashboard Cards -->
          <div class="space-y-3">
            <h5 class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.dashboardCards') }}</h5>

            <div class="flex items-center justify-between">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.displayAlertEvents') }}</label>
                <p class="mt-1 text-xs text-gray-500">
                  {{ t('admin.ops.settings.displayAlertEventsHint') }}
                </p>
              </div>
              <Toggle v-model="advancedSettings.display_alert_events" />
            </div>

            <div class="flex items-center justify-between">
              <div>
                <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.ops.settings.displayOpenAITokenStats') }}</label>
                <p class="mt-1 text-xs text-gray-500">
                  {{ t('admin.ops.settings.displayOpenAITokenStatsHint') }}
                </p>
              </div>
              <Toggle v-model="advancedSettings.display_openai_token_stats" />
            </div>
          </div>
        </div>
      </details>
    </div>

    <template #footer>
      <div class="flex justify-end gap-2">
        <button class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="saving || !validation.valid" @click="saveAllSettings">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>
