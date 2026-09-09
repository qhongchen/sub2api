<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import HelpTooltip from '@/components/common/HelpTooltip.vue'

const props = defineProps<{
  platform?: string
  type?: string
}>()
const model = defineModel<string>({ default: '' })
const { t } = useI18n()

interface HeaderExample {
  label: string
  header: string
  note?: string
}

const OFFICIAL_REQUEST_ID_HEADERS: Record<string, { platform: string; header: string }> = {
  openai: { platform: 'OpenAI', header: 'x-request-id' },
  anthropic: { platform: 'Anthropic', header: 'request-id' },
  gemini: { platform: 'Gemini', header: 'x-goog-request-id' },
  antigravity: { platform: 'Antigravity', header: 'x-goog-request-id' },
  grok: { platform: 'Grok', header: 'xai-request-id' }
}

const examples = computed<HeaderExample[]>(() => {
  const items: HeaderExample[] = []
  if (props.type === 'apikey') {
    items.push({
      label: 'sub2api',
      header: 'X-Client-Request-ID',
      note: t('admin.accounts.upstreamRequestIdHeaderHelp.sub2apiNote')
    })
    items.push({ label: 'new-api / one-api', header: 'X-Oneapi-Request-Id' })
  }
  const official = props.platform ? OFFICIAL_REQUEST_ID_HEADERS[props.platform] : undefined
  if (official) {
    items.push({
      label: t('admin.accounts.upstreamRequestIdHeaderHelp.official', { platform: official.platform }),
      header: official.header
    })
  }
  return items
})
</script>

<template>
  <div>
    <div class="flex items-center">
      <label for="upstream-request-id-header" class="input-label">{{ t('admin.accounts.upstreamRequestIdHeader') }}</label>
      <HelpTooltip width-class="w-80 max-w-[calc(100vw-2rem)]">
        <p>{{ t('admin.accounts.upstreamRequestIdHeaderHelp.intro') }}</p>
        <template v-if="examples.length">
          <p class="mt-2 font-medium">{{ t('admin.accounts.upstreamRequestIdHeaderHelp.examplesTitle') }}</p>
          <ul class="mt-1 space-y-1.5">
            <li v-for="item in examples" :key="item.label">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <span class="text-gray-300">{{ item.label }}</span>
                <code class="select-all rounded bg-white/10 px-1.5 py-0.5 font-mono text-[11px] text-white">{{ item.header }}</code>
              </div>
              <p v-if="item.note" class="mt-0.5 text-gray-400">{{ item.note }}</p>
            </li>
          </ul>
        </template>
      </HelpTooltip>
    </div>
    <input
      id="upstream-request-id-header"
      v-model="model"
      data-testid="upstream-request-id-header"
      type="text"
      class="input"
      :placeholder="t('admin.accounts.upstreamRequestIdHeaderPlaceholder')"
    />
  </div>
</template>
