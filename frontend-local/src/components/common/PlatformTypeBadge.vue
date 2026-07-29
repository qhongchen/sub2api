<template>
  <div class="inline-flex min-w-0 flex-col gap-0.5 text-xs font-medium">
    <div
      v-if="showPlatform || showType"
      :class="[
        'inline-flex items-center',
        isCombinedMode ? 'overflow-hidden rounded-md' : 'gap-1.5'
      ]"
    >
      <span
        v-if="showPlatform"
        :class="[
          'inline-flex items-center gap-1 px-2 py-1',
          !isCombinedMode && 'rounded-md',
          platformClass
        ]"
      >
        <PlatformIcon :platform="platform" size="xs" />
        <span>{{ platformLabel }}</span>
      </span>
      <span
        v-if="showType"
        :class="[
          'inline-flex items-center gap-1 px-1.5 py-1',
          !isCombinedMode && 'rounded-md',
          typeClass
        ]"
      >
        <svg
          v-if="type === 'oauth'"
          class="h-3 w-3"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
          />
        </svg>
        <Icon v-else-if="type === 'setup-token'" name="shield" size="xs" />
        <Icon v-else-if="type === 'service_account'" name="cloud" size="xs" />
        <Icon v-else name="key" size="xs" />
        <span>{{ typeLabel }}</span>
      </span>
    </div>

    <div
      v-if="(showPlan && planLabel) || (showPrivacy && privacyBadge)"
      :class="[
        'inline-flex items-center',
        isCombinedMode ? 'overflow-hidden rounded-md' : 'gap-1.5'
      ]"
    >
      <span
        v-if="showPlan && planLabel"
        :class="[
          'inline-flex items-center gap-1 px-1.5 py-1',
          !isCombinedMode && 'rounded-md',
          planBadgeClass
        ]"
      >
        <GrokFreeIcon
          v-if="isGrokFreePlan"
          data-testid="grok-free-plan-icon"
        />
        <Icon
          v-else-if="planIconName"
          :name="planIconName"
          size="xs"
          data-testid="grok-plan-icon"
          aria-hidden="true"
        />
        <span>{{ planLabel }}</span>
      </span>
      <span
        v-if="showPrivacy && privacyBadge"
        :class="[
          'inline-flex items-center gap-1 px-1.5 py-1',
          !isCombinedMode && 'rounded-md',
          privacyBadge.class
        ]"
        :title="privacyBadge.title"
      >
        <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" :d="privacyBadge.icon" />
        </svg>
        <span>{{ privacyBadge.label }}</span>
      </span>
    </div>

    <div
      v-if="showPlan && expiresLabel"
      class="pl-0.5 text-[10px] leading-tight text-gray-400 dark:text-gray-500"
      :title="subscriptionExpiresAt"
    >
      {{ expiresLabel }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AccountPlatform, AccountType } from '@/types'
import GrokFreeIcon from './GrokFreeIcon.vue'
import PlatformIcon from './PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

type DisplayMode = 'all' | 'platform' | 'type' | 'plan' | 'privacy'

interface Props {
  platform: AccountPlatform
  type: AccountType
  authMode?: string
  mode?: DisplayMode
  planType?: string
  privacyMode?: string
  subscriptionExpiresAt?: string
}

const props = defineProps<Props>()

const displayMode = computed<DisplayMode>(() => props.mode ?? 'all')
const isCombinedMode = computed(() => displayMode.value === 'all')
const showPlatform = computed(() => displayMode.value === 'all' || displayMode.value === 'platform')
const showType = computed(() => displayMode.value === 'all' || displayMode.value === 'type')
const showPlan = computed(() => displayMode.value === 'all' || displayMode.value === 'plan')
const showPrivacy = computed(() => displayMode.value === 'all' || displayMode.value === 'privacy')

const platformLabel = computed(() => {
  if (props.platform === 'anthropic') return 'Anthropic'
  if (props.platform === 'openai') return 'OpenAI'
  if (props.platform === 'antigravity') return 'Antigravity'
  if (props.platform === 'grok') return 'Grok'
  return 'Gemini'
})

const typeLabel = computed(() => {
  if (props.platform === 'openai' && props.type === 'oauth') {
    const normalizedAuthMode = (props.authMode || '').trim().toLowerCase().replace(/[\s_-]+/g, '')
    if (normalizedAuthMode === 'agentidentity') return 'Agent Identity'
    if (normalizedAuthMode === 'personalaccesstoken') return 'PAT'
  }

  switch (props.type) {
    case 'oauth':
      return 'OAuth'
    case 'setup-token':
      return 'Token'
    case 'apikey':
      return 'Key'
    case 'bedrock':
      return 'AWS'
    case 'service_account':
      return 'Vertex'
    default:
      return props.type
  }
})

const normalizedPlanType = computed(() =>
  (props.planType || '').trim().toLowerCase().replace(/[\s_-]+/g, '')
)

const planLabel = computed(() => {
  if (!normalizedPlanType.value) return ''
  switch (normalizedPlanType.value) {
    case 'plus':
      return 'Plus'
    case 'team':
      return 'Team'
    case 'chatgptpro':
    case 'pro':
      return 'Pro'
    case 'free':
    case 'basic':
      return props.platform === 'grok' ? 'Grok Free' : 'Free'
    case 'supergrok':
      return 'SuperGrok'
    case 'supergrokheavy':
      return 'SuperGrok Heavy'
    case 'abnormal':
      return t('admin.accounts.subscriptionAbnormal')
    default:
      return props.planType
  }
})

const isGrokFreePlan = computed(() =>
  props.platform === 'grok' &&
  (normalizedPlanType.value === 'free' || normalizedPlanType.value === 'basic')
)

const planIconName = computed<'bolt' | null>(() => {
  if (props.platform !== 'grok') return null
  if (
    normalizedPlanType.value === 'supergrok' ||
    normalizedPlanType.value === 'supergrokheavy'
  ) {
    return 'bolt'
  }
  return null
})

const platformClass = computed(() => {
  if (props.platform === 'anthropic') {
    return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
  }
  if (props.platform === 'openai') {
    return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
  }
  if (props.platform === 'antigravity') {
    return 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'
  }
  if (props.platform === 'grok') {
    return 'bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300'
  }
  return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
})

const typeClass = computed(() => {
  if (props.platform === 'anthropic') {
    return 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400'
  }
  if (props.platform === 'openai') {
    return 'bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400'
  }
  if (props.platform === 'antigravity') {
    return 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400'
  }
  if (props.platform === 'grok') {
    return 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300'
  }
  return 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
})

const planBadgeClass = computed(() => {
  if (normalizedPlanType.value === 'abnormal') {
    return 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400'
  }
  return typeClass.value
})

const expiresLabel = computed(() => {
  if (!props.subscriptionExpiresAt || !props.planType) return ''
  if (normalizedPlanType.value === 'free' || normalizedPlanType.value === 'basic') return ''
  try {
    const d = new Date(props.subscriptionExpiresAt)
    if (isNaN(d.getTime())) return ''
    const yyyy = d.getFullYear()
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    return `${t('admin.accounts.subscriptionExpires')} ${yyyy}-${mm}-${dd}`
  } catch {
    return ''
  }
})

const privacyBadge = computed(() => {
  if (props.type !== 'oauth' || !props.privacyMode) return null
  if (props.platform !== 'openai' && props.platform !== 'antigravity') return null

  const shieldCheck = 'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z'
  const shieldX = 'M12 9v3.75m0-10.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285zM12 18h.008v.008H12V18z'
  switch (props.privacyMode) {
    case 'training_off':
      return { label: 'Private', icon: shieldCheck, title: t('admin.accounts.privacyTrainingOff'), class: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400' }
    case 'training_set_cf_blocked':
      return { label: 'CF', icon: shieldX, title: t('admin.accounts.privacyCfBlocked'), class: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-400' }
    case 'training_set_failed':
      return { label: 'Fail', icon: shieldX, title: t('admin.accounts.privacyFailed'), class: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400' }
    case 'privacy_set':
      return { label: 'Private', icon: shieldCheck, title: t('admin.accounts.privacyAntigravitySet'), class: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400' }
    case 'privacy_set_failed':
      return { label: 'Fail', icon: shieldX, title: t('admin.accounts.privacyAntigravityFailed'), class: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400' }
    default:
      return null
  }
})
</script>
