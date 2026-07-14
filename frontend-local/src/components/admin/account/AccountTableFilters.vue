<template>
  <div class="account-filter-row">
    <SearchInput
      :model-value="searchQuery"
      :placeholder="t('admin.accounts.searchAccounts')"
      class="account-filter-search account-filter-control"
      @update:model-value="$emit('update:searchQuery', $event)"
      @search="$emit('change')"
    />

    <Select
      :model-value="filters.platform"
      class="account-filter-select"
      :options="pOpts"
      @update:model-value="updatePlatform"
    />
    <Select
      :model-value="filters.status"
      class="account-filter-select"
      :options="sOpts"
      @update:model-value="updateStatus"
    />
    <Select
      :model-value="filters.group"
      class="account-filter-select account-filter-group"
      :options="gOpts"
      @update:model-value="updateGroup"
    />
    <div class="account-filter-actions">
      <button
        type="button"
        class="btn btn-primary account-filter-apply"
        @click="$emit('change')"
      >
        {{ t('usage.applyFilters') }}
      </button>
      <button
        type="button"
        class="btn btn-secondary account-filter-reset"
        :disabled="!showReset"
        @click="$emit('reset')"
      >
        {{ t('common.reset') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import type { AdminGroup } from '@/types'

const props = defineProps<{ searchQuery: string; filters: Record<string, any>; groups?: AdminGroup[] }>()
const emit = defineEmits(['update:searchQuery', 'update:filters', 'change', 'reset'])

const { t } = useI18n()

const showReset = computed(() => {
  return Boolean(
    props.searchQuery ||
    props.filters.platform ||
    props.filters.status ||
    props.filters.group
  )
})

const updatePlatform = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, platform: value }) }
const updateStatus = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, status: value }) }
const updateGroup = (value: string | number | boolean | null) => { emit('update:filters', { ...props.filters, group: value }) }

const pOpts = computed(() => [{ value: '', label: t('admin.accounts.allPlatforms') }, { value: 'anthropic', label: 'Anthropic' }, { value: 'openai', label: 'OpenAI' }, { value: 'gemini', label: 'Gemini' }, { value: 'antigravity', label: 'Antigravity' }, { value: 'grok', label: 'Grok' }])
const sOpts = computed(() => [{ value: '', label: t('admin.accounts.allStatus') }, { value: 'active', label: t('admin.accounts.status.active') }, { value: 'inactive', label: t('admin.accounts.status.inactive') }, { value: 'error', label: t('admin.accounts.status.error') }, { value: 'rate_limited', label: t('admin.accounts.status.rateLimited') }, { value: 'temp_unschedulable', label: t('admin.accounts.status.tempUnschedulable') }, { value: 'unschedulable', label: t('admin.accounts.status.unschedulable') }])
const gOpts = computed(() => [
  { value: '', label: t('admin.accounts.allGroups') },
  { value: 'ungrouped', label: t('admin.accounts.ungroupedGroup') },
  ...(props.groups || []).map(g => ({ value: String(g.id), label: g.name }))
])
</script>

<style scoped>
.account-filter-row {
  @apply flex flex-wrap items-center gap-3;
}

.account-filter-search {
  @apply min-w-[180px] flex-none sm:w-56;
}

.account-filter-select {
  @apply w-full shrink-0 sm:w-[132px];
}

.account-filter-group {
  @apply sm:w-[148px];
}

.account-filter-actions {
  @apply flex shrink-0 items-center gap-2;
}

.account-filter-apply,
.account-filter-reset {
  @apply whitespace-nowrap;
}
</style>
