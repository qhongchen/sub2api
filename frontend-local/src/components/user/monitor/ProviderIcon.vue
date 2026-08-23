<template>
  <PlatformIcon
    v-if="normalizedProvider"
    :platform="normalizedProvider"
    size="md"
    :style="iconStyle"
    aria-hidden="true"
  />
  <span
    v-else
    class="inline-flex items-center justify-center font-bold text-gray-500"
    :style="fallbackStyle"
  >
    {{ fallbackText }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Provider } from '@/api/admin/channelMonitor'
import type { GroupPlatform } from '@/types'
import PlatformIcon from '@/components/common/PlatformIcon.vue'

const props = withDefaults(defineProps<{
  provider: Provider | string
  size?: number
}>(), {
  size: 20,
})

const supportedProviders = new Set<GroupPlatform>([
  'openai',
  'anthropic',
  'gemini',
  'grok',
  'antigravity',
  'kimi',
  'zhipu',
  'deepseek',
])

const normalizedProvider = computed<GroupPlatform | undefined>(() =>
  supportedProviders.has(props.provider as GroupPlatform)
    ? props.provider as GroupPlatform
    : undefined
)

const iconStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
}))

const fallbackStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  fontSize: `${Math.round(props.size * 0.5)}px`,
}))

const fallbackText = computed(() =>
  (props.provider || '?').charAt(0).toUpperCase()
)
</script>
