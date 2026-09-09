<template>
  <div class="mt-4 border-t border-gray-200 pt-4 dark:border-dark-400">
    <div class="mb-3 flex items-start justify-between gap-3">
      <div class="min-w-0 flex-1">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t("admin.groups.codexModelsManifest.title") }}
        </label>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t("admin.groups.codexModelsManifest.hint") }}
        </p>
      </div>
      <Toggle
        :aria-label="t('admin.groups.codexModelsManifest.enable')"
        :model-value="config.enabled"
        @update:model-value="emitUpdate({ enabled: $event })"
      />
    </div>

    <div v-if="config.enabled">
      <p class="mb-2 text-xs text-gray-500 dark:text-gray-400">
        {{ t("admin.groups.codexModelsManifest.enabledHint") }}
      </p>
      <label for="codex-manifest-search" class="input-label">
        {{ t("admin.groups.codexModelsManifest.accounts") }}
      </label>
      <div v-if="config.account_ids.length" class="mb-2 flex flex-wrap gap-1.5">
        <span
          v-for="id in config.account_ids"
          :key="id"
          class="inline-flex max-w-full items-center gap-1 rounded bg-primary-100 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
        >
          <span class="min-w-0 break-all">{{ accountLabel(id) }}</span>
          <button
            type="button"
            class="ml-0.5 shrink-0 text-primary-500 hover:text-primary-700 dark:hover:text-primary-200"
            :aria-label="t('common.remove')"
            :title="t('common.remove')"
            @click="removeAccount(id)"
          >
            <Icon name="x" size="xs" />
          </button>
        </span>
      </div>

      <div ref="searchContainerRef" class="relative">
        <input
          id="codex-manifest-search"
          v-model="searchKeyword"
          type="text"
          class="input text-sm"
          :placeholder="t('admin.groups.codexModelsManifest.searchPlaceholder')"
          :disabled="config.account_ids.length >= maxAccounts"
          @input="searchAccounts"
          @focus="onSearchFocus"
          @keydown.esc="showDropdown = false"
        />
        <div
          v-if="showDropdown && (searchResults.length > 0 || searchKeyword.trim() !== '')"
          class="absolute z-50 mt-1 max-h-48 w-full overflow-auto rounded-lg border bg-white shadow-lg dark:border-dark-600 dark:bg-dark-800"
        >
          <p v-if="searchResults.length === 0" class="px-3 py-2 text-sm text-gray-400">
            {{ t("admin.groups.codexModelsManifest.searchEmpty") }}
          </p>
          <button
            v-for="account in searchResults"
            :key="account.id"
            type="button"
            class="w-full break-all px-3 py-2 text-left text-sm hover:bg-gray-100 disabled:opacity-50 dark:hover:bg-dark-700"
            :disabled="config.account_ids.includes(account.id) || config.account_ids.length >= maxAccounts"
            @click="selectAccount(account)"
          >
            <span>{{ account.name }}</span>
            <span class="ml-2 text-xs text-gray-400">#{{ account.id }}</span>
          </button>
        </div>
      </div>

      <div class="mt-3 flex items-start justify-between gap-3">
        <div class="min-w-0 flex-1">
          <label class="text-sm text-gray-700 dark:text-gray-300">
            {{ t("admin.groups.codexModelsManifest.fallback") }}
          </label>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t("admin.groups.codexModelsManifest.fallbackHint") }}
          </p>
        </div>
        <Toggle
          :aria-label="t('admin.groups.codexModelsManifest.fallback')"
          :model-value="config.fallback_to_scheduler"
          @update:model-value="emitUpdate({ fallback_to_scheduler: $event })"
        />
      </div>
      <p v-if="showValidationError" class="mt-2 text-xs text-red-600 dark:text-red-400" role="alert">
        {{ t(config.account_ids.length > maxAccounts ? "admin.groups.codexModelsManifest.maxAccounts" : "admin.groups.codexModelsManifest.selectAtLeastOne") }}
      </p>
    </div>
    <p v-else class="text-xs text-gray-500 dark:text-gray-400">
      {{ t("admin.groups.codexModelsManifest.disabledHint") }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import Toggle from "@/components/common/Toggle.vue";
import { useKeyedDebouncedSearch } from "@/composables/useKeyedDebouncedSearch";
import { adminAPI } from "@/api/admin";
import type { CodexModelsManifestConfig } from "@/types";

interface SimpleAccount {
  id: number;
  name: string;
}

const props = defineProps<{
  groupId: number;
  modelValue: CodexModelsManifestConfig;
  accountNames?: Record<number, string>;
}>();
const emit = defineEmits<{
  (event: "update:modelValue", value: CodexModelsManifestConfig): void;
}>();
const { t } = useI18n();
const maxAccounts = 10;
const localNames = ref<Record<number, string>>({});
const config = computed(() => props.modelValue);
const emitUpdate = (patch: Partial<CodexModelsManifestConfig>) =>
  emit("update:modelValue", { ...props.modelValue, ...patch });
const accountLabel = (id: number) =>
  props.accountNames?.[id] ?? localNames.value[id] ?? `#${id}`;
const searchKeyword = ref("");
const searchResults = ref<SimpleAccount[]>([]);
const showDropdown = ref(false);
const showValidationError = ref(false);
const searchContainerRef = ref<HTMLElement | null>(null);

const handleDocumentClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement;
  if (searchContainerRef.value && !searchContainerRef.value.contains(target)) {
    showDropdown.value = false;
  }
};
onMounted(() => document.addEventListener("click", handleDocumentClick));
onUnmounted(() => document.removeEventListener("click", handleDocumentClick));

const searchRunner = useKeyedDebouncedSearch<SimpleAccount[]>({
  delay: 300,
  search: async (keyword, { signal }) => {
    const res = await adminAPI.accounts.list(
      1,
      20,
      { search: keyword, platform: "openai", group: String(props.groupId), lite: "1" },
      { signal },
    );
    return res.items.map(({ id, name }) => ({ id, name }));
  },
  onSuccess: (_key, result) => { searchResults.value = result; },
  onError: () => { searchResults.value = []; },
});
const searchAccounts = () => {
  showDropdown.value = true;
  searchRunner.trigger("codex-manifest", searchKeyword.value);
};
const onSearchFocus = () => {
  showDropdown.value = true;
  if (searchResults.value.length === 0) searchAccounts();
};
const selectAccount = (account: SimpleAccount) => {
  if (config.value.account_ids.includes(account.id) || config.value.account_ids.length >= maxAccounts) return;
  localNames.value[account.id] = account.name;
  emitUpdate({ account_ids: [...config.value.account_ids, account.id] });
  searchKeyword.value = "";
  showDropdown.value = false;
  showValidationError.value = false;
};
const removeAccount = (id: number) =>
  emitUpdate({ account_ids: config.value.account_ids.filter(accountId => accountId !== id) });
const validate = (): boolean => {
  showValidationError.value = config.value.enabled &&
    (config.value.account_ids.length === 0 || config.value.account_ids.length > maxAccounts);
  return !showValidationError.value;
};
const resetValidation = () => { showValidationError.value = false; };
defineExpose({ validate, resetValidation });
</script>
