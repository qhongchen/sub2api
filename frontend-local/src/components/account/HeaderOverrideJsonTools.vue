<template>
  <button
    type="button"
    class="rounded-lg bg-primary-50 px-3 py-1 text-xs text-primary-700 transition-colors hover:bg-primary-100 dark:bg-primary-900/30 dark:text-primary-400 dark:hover:bg-primary-900/50"
    @click="toggleImportPanel"
  >
    {{ message('importJson', 'Import JSON') }}
  </button>
  <button
    type="button"
    class="rounded-lg bg-primary-50 px-3 py-1 text-xs text-primary-700 transition-colors hover:bg-primary-100 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-primary-900/30 dark:text-primary-400 dark:hover:bg-primary-900/50"
    :disabled="!hasNamedRows"
    @click="copyAsJson"
  >
    {{ message('copyJson', 'Copy JSON') }}
  </button>

  <div v-if="showImportPanel" ref="importPanelRef" class="w-full space-y-2">
    <textarea
      ref="importTextareaRef"
      v-model="importText"
      rows="5"
      class="input font-mono text-xs"
      placeholder='{"user-agent":"my-client/1.0","x-relay-token":"..."}'
    />
    <div class="flex gap-2">
      <button type="button" class="rounded-lg bg-primary-600 px-3 py-1 text-xs text-white hover:bg-primary-700" @click="applyImport">
        {{ message('importJsonApply', 'Apply') }}
      </button>
      <button type="button" class="rounded-lg bg-gray-100 px-3 py-1 text-xs text-gray-600 hover:bg-gray-200 dark:bg-dark-600 dark:text-gray-400" @click="closeImportPanel">
        {{ message('importJsonCancel', 'Cancel') }}
      </button>
    </div>
    <p class="text-xs text-gray-500 dark:text-gray-400">
      {{ message('importJsonHint', 'Paste a flat JSON object. Values may be strings, numbers, or booleans.') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { parseHeaderOverridesJson, serializeHeaderOverrideRows, type HeaderOverrideRow } from './credentialsBuilder'

const props = defineProps<{ rows: HeaderOverrideRow[] }>()
const emit = defineEmits<{ (event: 'update:rows', rows: HeaderOverrideRow[]): void }>()
const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const showImportPanel = ref(false)
const importText = ref('')
const importPanelRef = ref<HTMLElement | null>(null)
const importTextareaRef = ref<HTMLTextAreaElement | null>(null)
const hasNamedRows = computed(() => props.rows.some((row) => row.name.trim()))

const message = (suffix: string, fallback: string) => {
  const key = `admin.accounts.headerOverride.${suffix}`
  const translated = t(key)
  return translated === key ? fallback : translated
}

const toggleImportPanel = async () => {
  showImportPanel.value = !showImportPanel.value
  if (!showImportPanel.value) return
  await nextTick()
  importPanelRef.value?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
  importTextareaRef.value?.focus({ preventScroll: true })
}

const closeImportPanel = () => {
  showImportPanel.value = false
  importText.value = ''
}

const applyImport = () => {
  const rows = parseHeaderOverridesJson(importText.value)
  if (rows === null) {
    appStore.showError(message('importJsonInvalid', 'Invalid header JSON'))
    return
  }
  emit('update:rows', rows)
  closeImportPanel()
}

const copyAsJson = () => {
  void copyToClipboard(serializeHeaderOverrideRows(props.rows))
}
</script>
