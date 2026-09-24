<template>
  <div>
    <div class="mb-2 flex items-center justify-between gap-2">
      <label class="input-label mb-0">{{ t('admin.accounts.opencodeGo.protocolRules.title') }}</label>
      <button
        type="button"
        class="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400"
        @click="restoreDefaults"
      >
        {{ t('admin.accounts.opencodeGo.protocolRules.restoreDefaults') }}
      </button>
    </div>
    <p class="input-hint mb-2">{{ t('admin.accounts.opencodeGo.protocolRules.hint') }}</p>
    <div v-if="rows.length > 0" class="mb-2 space-y-2">
      <div v-for="(row, index) in rows" :key="getRowKey(row)" class="flex items-center gap-2">
        <input
          v-model="row.pattern"
          type="text"
          class="input min-w-0 flex-1 font-mono text-sm"
          :placeholder="t('admin.accounts.opencodeGo.protocolRules.patternPlaceholder')"
        />
        <select v-model="row.protocol" class="input w-36 shrink-0 sm:w-44">
          <option value="chat_completions">{{ t('admin.accounts.cnProviders.apiProtocol.chatCompletions') }}</option>
          <option value="responses">{{ t('admin.accounts.cnProviders.apiProtocol.responses') }}</option>
          <option value="anthropic">{{ t('admin.accounts.cnProviders.apiProtocol.anthropic') }}</option>
        </select>
        <button
          type="button"
          class="rounded p-2 text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"
          :aria-label="t('admin.accounts.opencodeGo.protocolRules.remove')"
          @click="removeRow(index)"
        >
          <Icon name="trash" size="sm" />
        </button>
      </div>
    </div>
    <div class="mb-2 flex items-center gap-2 border-t border-dashed border-gray-200 py-2 text-xs text-gray-500 dark:border-dark-600 dark:text-gray-400">
      <span class="flex-1 font-mono">*</span>
      <span>{{ t('admin.accounts.opencodeGo.protocolRules.fallback') }}</span>
    </div>
    <button
      type="button"
      class="w-full rounded border border-dashed border-gray-300 px-4 py-2 text-sm text-gray-600 hover:border-gray-400 dark:border-dark-500 dark:text-gray-400"
      @click="addRow"
    >
      {{ t('admin.accounts.opencodeGo.protocolRules.add') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { createStableObjectKeyResolver } from '@/utils/stableObjectKey'
import {
  cloneOpenCodeGoProtocolRules,
  defaultOpenCodeProtocolRules,
  type OpenCodeAccountMode,
  type OpenCodeGoProtocolRule
} from './credentialsBuilder'

const props = withDefaults(defineProps<{
  rows: OpenCodeGoProtocolRule[]
  plan?: OpenCodeAccountMode
}>(), { plan: 'go' })
const emit = defineEmits<{
  (e: 'update:rows', rows: OpenCodeGoProtocolRule[]): void
}>()
const { t } = useI18n()
const getRowKey = createStableObjectKeyResolver<OpenCodeGoProtocolRule>('opencode-go-protocol-rule')

function addRow(): void {
  emit('update:rows', [...props.rows, { pattern: '', protocol: 'chat_completions' }])
}
function removeRow(index: number): void {
  emit('update:rows', props.rows.filter((_, i) => i !== index))
}
function restoreDefaults(): void {
  emit('update:rows', cloneOpenCodeGoProtocolRules(defaultOpenCodeProtocolRules(props.plan)))
}
</script>
