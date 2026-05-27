<template>
  <div class="flex flex-wrap items-center gap-2">
    <slot name="before"></slot>
    <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary btn-sm px-2 md:px-2.5">
      <Icon name="refresh" size="sm" :class="[loading ? 'animate-spin' : '']" />
      <span class="hidden md:inline">{{ t('common.refresh') }}</span>
    </button>
    <slot name="after"></slot>
    <slot name="beforeCreate"></slot>
    <button @click="$emit('create')" class="btn btn-primary btn-sm px-2.5 md:px-3">
      <Icon name="plus" size="sm" class="md:mr-1.5" />
      <span>{{ t('admin.accounts.createAccount') }}</span>
    </button>
    <slot name="afterCreate"></slot>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps(['loading'])
defineEmits(['refresh', 'create'])

const { t } = useI18n()
</script>
