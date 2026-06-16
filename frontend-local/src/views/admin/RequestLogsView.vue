<template>
  <AppLayout>
    <div class="space-y-5">
      <section class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-semibold text-gray-950 dark:text-white">
            {{ t('admin.requestLogs.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">
            {{ t('admin.requestLogs.description') }}
          </p>
        </div>

      </section>

      <section
        v-if="opsDisabled"
        class="rounded-lg border border-amber-200 bg-amber-50 p-6 text-amber-900 shadow-sm dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-100"
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="flex gap-3">
            <span class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-200">
              <Icon name="exclamationTriangle" size="sm" :stroke-width="2" />
            </span>
            <div>
              <h2 class="text-base font-semibold">{{ t('admin.requestLogs.disabledTitle') }}</h2>
              <p class="mt-1 max-w-2xl text-sm text-amber-800 dark:text-amber-100/80">
                {{ t('admin.requestLogs.disabledDescription') }}
              </p>
            </div>
          </div>
          <router-link
            to="/admin/settings"
            class="inline-flex items-center justify-center gap-2 rounded-lg bg-amber-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-amber-700"
          >
            <Icon name="cog" size="sm" />
            {{ t('admin.requestLogs.openSettings') }}
          </router-link>
        </div>
      </section>

      <template v-else>
        <section class="space-y-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200/70 bg-white/70 px-3 py-2 text-sm font-medium text-gray-600 shadow-sm transition-colors hover:border-gray-300 hover:text-gray-950 dark:border-white/[0.08] dark:bg-dark-900/60 dark:text-dark-300 dark:hover:text-white"
              @click="filtersOpen = !filtersOpen"
            >
              <Icon name="filter" size="sm" :stroke-width="2" />
              <span>{{ t('admin.requestLogs.filters') }}</span>
              <span
                v-if="activeFilterCount > 0"
                class="rounded-full bg-orange-50 px-1.5 py-0.5 text-[10px] font-semibold text-orange-600 dark:bg-orange-500/10 dark:text-orange-300"
              >
                {{ activeFilterCount }}
              </span>
              <Icon
                name="chevronDown"
                size="xs"
                class="transition-transform"
                :class="{ 'rotate-180': filtersOpen }"
              />
            </button>

            <div class="ml-auto flex items-center gap-2">
              <div ref="columnDropdownRef" class="relative">
                <button
                  type="button"
                  class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-800 shadow-sm transition-colors hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-100 dark:hover:bg-white/[0.04]"
                  :title="t('admin.requestLogs.columnSettings')"
                  @click.stop="showColumnDropdown = !showColumnDropdown"
                >
                  <Icon name="grid" size="sm" />
                  <span>{{ visibleColumnCount }}/{{ allColumns.length }}</span>
                </button>

                <div
                  v-if="showColumnDropdown"
                  class="absolute right-0 top-full z-50 mt-2 max-h-96 w-60 overflow-y-auto rounded-lg border border-gray-200 bg-white p-1 shadow-lg dark:border-dark-700 dark:bg-dark-800"
                >
                  <div class="px-3 py-2 text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-dark-400">
                    {{ t('admin.requestLogs.columnSettings') }}
                  </div>
                  <button
                    v-for="col in toggleableColumns"
                    :key="col.key"
                    type="button"
                    class="flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-left text-sm text-gray-600 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
                    @click="toggleColumn(col.key)"
                  >
                    <span class="truncate">{{ col.label }}</span>
                    <Icon v-if="isColumnVisible(col.key)" name="check" size="sm" class="shrink-0 text-orange-500" :stroke-width="2" />
                  </button>
                  <div class="my-1 border-t border-gray-100 dark:border-white/[0.06]" />
                  <button
                    type="button"
                    class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-400 dark:hover:bg-dark-700 dark:hover:text-white"
                    @click="resetColumns"
                  >
                    <Icon name="refresh" size="xs" />
                    {{ t('admin.requestLogs.resetColumns') }}
                  </button>
                </div>
              </div>
              <button
                type="button"
                class="consumer-icon-btn"
                :title="t('admin.requestLogs.refresh')"
                :disabled="loading || refreshingSilently"
                @click="refreshData"
              >
                <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading || refreshingSilently }" />
              </button>

              <button
                type="button"
                class="inline-flex h-9 items-center gap-2 rounded-full px-2 text-sm text-gray-500 transition-colors hover:text-gray-950 dark:text-dark-300 dark:hover:text-white"
                :title="t('usage.autoRefresh')"
                @click="toggleAutoRefresh"
              >
                <span
                  class="h-1.5 w-1.5 rounded-full"
                  :class="autoRefreshEnabled ? 'bg-emerald-500' : 'bg-gray-300 dark:bg-dark-600'"
                />
                <span
                  class="relative inline-flex h-6 w-10 items-center rounded-full transition-colors"
                  :class="autoRefreshEnabled ? 'bg-orange-500' : 'bg-gray-200 dark:bg-dark-700'"
                >
                  <span
                    class="inline-block h-5 w-5 rounded-full bg-white shadow transition-transform"
                    :class="autoRefreshEnabled ? 'translate-x-4' : 'translate-x-0.5'"
                  />
                </span>
              </button>
            </div>
          </div>

          <Transition
            enter-active-class="transition duration-200 ease-out"
            enter-from-class="-translate-y-1 opacity-0"
            enter-to-class="translate-y-0 opacity-100"
            leave-active-class="transition duration-150 ease-in"
            leave-from-class="translate-y-0 opacity-100"
            leave-to-class="-translate-y-1 opacity-0"
          >
            <div v-if="filtersOpen" class="cch-panel-card p-4 md:p-5">
              <div class="grid gap-4 lg:grid-cols-2">
                <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]" data-test="request-log-filter-time">
                  <button
                    type="button"
                    class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                    @click="timeFiltersOpen = !timeFiltersOpen"
                  >
                    <span class="flex min-w-0 items-center gap-3">
                      <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                        <Icon name="clock" size="sm" />
                      </span>
                      <span>
                        <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('admin.requestLogs.filterGroups.time') }}</span>
                        <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.filterGroups.timeDescription') }}</span>
                      </span>
                    </span>
                    <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': timeFiltersOpen }" />
                  </button>

                  <div v-if="timeFiltersOpen" class="space-y-4 px-4 pb-4">
                    <div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap">
                      <button
                        v-for="option in timeRangeOptions"
                        :key="option.value"
                        type="button"
                        class="inline-flex items-center justify-center gap-2 rounded-lg border px-3 py-2 text-sm font-semibold shadow-sm transition-colors"
                        :class="filters.time_range === option.value
                          ? 'border-orange-500 bg-orange-500 text-white'
                          : 'border-gray-200 bg-white text-gray-800 hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-100 dark:hover:bg-white/[0.04]'"
                        @click="selectTimeRange(option.value)"
                      >
                        <Icon :name="option.icon" size="sm" />
                        {{ option.label }}
                      </button>
                    </div>

                    <div class="grid gap-3 sm:grid-cols-2">
                      <label class="block">
                        <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.startTime') }}</span>
                        <input v-model="customStartTime" type="datetime-local" class="input" @change="onCustomTimeChange" @keyup.enter="applyFilters" />
                      </label>
                      <label class="block">
                        <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.endTime') }}</span>
                        <input v-model="customEndTime" type="datetime-local" class="input" @change="onCustomTimeChange" @keyup.enter="applyFilters" />
                      </label>
                    </div>
                  </div>
                </div>

                <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]" data-test="request-log-filter-result">
                  <button
                    type="button"
                    class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                    @click="resultFiltersOpen = !resultFiltersOpen"
                  >
                    <span class="flex min-w-0 items-center gap-3">
                      <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                        <Icon name="checkCircle" size="sm" />
                      </span>
                      <span>
                        <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('admin.requestLogs.filterGroups.result') }}</span>
                        <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.filterGroups.resultDescription') }}</span>
                      </span>
                    </span>
                    <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': resultFiltersOpen }" />
                  </button>

                  <div v-if="resultFiltersOpen" class="grid gap-3 px-4 pb-4 sm:grid-cols-2">
                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.statusCode') }}</span>
                      <Select v-model="filters.statusPreset" :options="statusCodeOptions" />
                    </label>

                    <label v-if="filters.statusPreset === 'custom'" class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.customStatusCode') }}</span>
                      <input
                        v-model.trim="filters.customStatusCode"
                        type="number"
                        min="100"
                        max="599"
                        class="input"
                        placeholder="503"
                        @keyup.enter="applyFilters"
                      />
                    </label>

                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.sort') }}</span>
                      <Select v-model="filters.sort" :options="sortOptions" />
                    </label>
                  </div>
                </div>

                <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]" data-test="request-log-filter-identity">
                  <button
                    type="button"
                    class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                    @click="identityFiltersOpen = !identityFiltersOpen"
                  >
                    <span class="flex min-w-0 items-center gap-3">
                      <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                        <Icon name="user" size="sm" />
                      </span>
                      <span>
                        <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('admin.requestLogs.filterGroups.identity') }}</span>
                        <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.filterGroups.identityDescription') }}</span>
                      </span>
                    </span>
                    <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': identityFiltersOpen }" />
                  </button>

                  <div v-if="identityFiltersOpen" class="grid gap-3 px-4 pb-4 sm:grid-cols-2">
                    <label ref="userSearchRef" class="relative block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.user') }}</span>
                      <input
                        v-model="userKeyword"
                        type="text"
                        class="input pr-8"
                        :placeholder="t('admin.requestLogs.searchUserPlaceholder')"
                        @focus="showUserDropdown = true"
                        @input="debounceUserSearch"
                        @keyup.enter="applyFilters"
                      />
                      <button
                        v-if="filters.user_id"
                        type="button"
                        class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                        @click="clearUser"
                      >
                        <Icon name="x" size="xs" />
                      </button>
                      <div
                        v-if="showUserDropdown && userResults.length > 0"
                        class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                      >
                        <button
                          v-for="user in userResults"
                          :key="user.id"
                          type="button"
                          class="flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                          @click="selectUser(user)"
                        >
                          <span class="truncate">{{ getUserDisplayName(user) }}</span>
                          <span class="text-xs text-gray-400">#{{ user.id }}</span>
                        </button>
                      </div>
                    </label>

                    <label ref="apiKeySearchRef" class="relative block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.apiKey') }}</span>
                      <input
                        v-model="apiKeyKeyword"
                        type="text"
                        class="input pr-8"
                        :placeholder="t('admin.requestLogs.searchApiKeyPlaceholder')"
                        @focus="onApiKeyFocus"
                        @input="debounceApiKeySearch"
                        @keyup.enter="applyFilters"
                      />
                      <button
                        v-if="filters.api_key_id"
                        type="button"
                        class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                        @click="clearApiKey"
                      >
                        <Icon name="x" size="xs" />
                      </button>
                      <div
                        v-if="showApiKeyDropdown && apiKeyResults.length > 0"
                        class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                      >
                        <button
                          v-for="key in apiKeyResults"
                          :key="key.id"
                          type="button"
                          class="flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                          @click="selectApiKey(key)"
                        >
                          <span class="truncate">{{ key.name || `#${key.id}` }}</span>
                          <span class="text-xs text-gray-400">#{{ key.id }}</span>
                        </button>
                      </div>
                    </label>

                    <label ref="accountSearchRef" class="relative block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.account') }}</span>
                      <input
                        v-model="accountKeyword"
                        type="text"
                        class="input pr-8"
                        :placeholder="t('admin.requestLogs.searchAccountPlaceholder')"
                        @focus="showAccountDropdown = true"
                        @input="debounceAccountSearch"
                        @keyup.enter="applyFilters"
                      />
                      <button
                        v-if="filters.account_id"
                        type="button"
                        class="absolute right-2 top-8 text-gray-400 transition-colors hover:text-gray-700 dark:hover:text-gray-200"
                        @click="clearAccount"
                      >
                        <Icon name="x" size="xs" />
                      </button>
                      <div
                        v-if="showAccountDropdown && accountResults.length > 0"
                        class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
                      >
                        <button
                          v-for="account in accountResults"
                          :key="account.id"
                          type="button"
                          class="flex w-full items-center justify-between gap-2 px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                          @click="selectAccount(account)"
                        >
                          <span class="truncate">{{ account.name || `#${account.id}` }}</span>
                          <span class="text-xs text-gray-400">#{{ account.id }}</span>
                        </button>
                      </div>
                    </label>

                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.group') }}</span>
                      <Select v-model="filters.group_id" :options="groupFilterOptions" searchable />
                    </label>
                  </div>
                </div>

                <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]" data-test="request-log-filter-routing">
                  <button
                    type="button"
                    class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                    @click="requestFiltersOpen = !requestFiltersOpen"
                  >
                    <span class="flex min-w-0 items-center gap-3">
                      <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                        <Icon name="server" size="sm" />
                      </span>
                      <span>
                        <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('admin.requestLogs.filterGroups.routing') }}</span>
                        <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.filterGroups.routingDescription') }}</span>
                      </span>
                    </span>
                    <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': requestFiltersOpen }" />
                  </button>

                  <div v-if="requestFiltersOpen" class="grid gap-3 px-4 pb-4 sm:grid-cols-2">
                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.model') }}</span>
                      <input
                        v-model.trim="filters.model"
                        type="text"
                        class="input"
                        :placeholder="t('admin.requestLogs.modelPlaceholder')"
                        @keyup.enter="applyFilters"
                      />
                    </label>

                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.requestId') }}</span>
                      <input
                        v-model.trim="filters.request_id"
                        type="text"
                        class="input font-mono"
                        :placeholder="t('admin.requestLogs.requestIdPlaceholder')"
                        @keyup.enter="applyFilters"
                      />
                    </label>

                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('admin.requestLogs.sessionId') }}</span>
                      <input
                        v-model.trim="filters.session_id"
                        type="text"
                        class="input font-mono"
                        :placeholder="t('admin.requestLogs.sessionIdPlaceholder')"
                        @keyup.enter="applyFilters"
                      />
                    </label>
                  </div>
                </div>
              </div>

              <div class="mt-5 flex flex-wrap items-center gap-2">
                <button type="button" class="btn btn-primary" :disabled="loading" @click="applyFilters">
                  {{ t('admin.requestLogs.apply') }}
                </button>
                <button type="button" class="btn btn-secondary" @click="resetFilters">
                  {{ t('admin.requestLogs.reset') }}
                </button>
              </div>
            </div>
          </Transition>
        </section>

        <TablePanel
          :error="errorMessage"
          :summary="loadedRecordsSummary"
          :summary-items="requestLogStatsSummaryItems"
        >
          <UsageTable
            :data="logs"
            :loading="loading"
            :columns="visibleColumns"
            shell-class="overflow-hidden"
            :server-side-sort="false"
            :default-sort-key="'created_at'"
            :default-sort-order="'desc'"
            :show-request-type-badge="false"
            :show-diagnostics-action="true"
            @session-click="filterBySession"
            @diagnostics-click="openDiagnostics"
          />

          <template #footer>
            <div class="flex min-h-10 items-center justify-center text-sm text-gray-500 dark:text-dark-400">
              <span v-if="loadingMore" class="inline-flex items-center gap-2">
                <Icon name="refresh" size="sm" class="animate-spin" />
                {{ t('admin.requestLogs.loadingMore') }}
              </span>
              <span v-else-if="hasMoreLogs">
                {{ t('admin.requestLogs.scrollLoadMore') }}
              </span>
              <span v-else-if="logs.length > 0">
                {{ t('admin.requestLogs.allRecordsLoaded') }}
              </span>
            </div>
            <div ref="loadMoreSentinelRef" class="h-px" aria-hidden="true" />
          </template>
        </TablePanel>
      </template>
    </div>

    <BaseDialog
      :show="Boolean(selectedLog)"
      :title="t('admin.requestLogs.diagnostics.title')"
      width="full"
      :close-on-click-outside="true"
      @close="closeDiagnostics"
    >
      <div v-if="selectedLog" class="space-y-5" data-testid="request-diagnostics-dialog">
        <section class="rounded-lg border border-gray-200 bg-gray-50/70 p-4 dark:border-white/[0.08] dark:bg-white/[0.03]">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="min-w-0 space-y-3">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold"
                  :class="getFinalStatusClass(selectedLog)"
                  data-testid="diagnostics-final-status"
                >
                  {{ getFinalStatusLabel(selectedLog) }}
                </span>
                <span
                  v-if="hasFailedUpstreamAttempt(selectedLog) && isFinalSuccess(selectedLog)"
                  class="inline-flex items-center rounded-full bg-sky-50 px-2.5 py-1 text-xs font-semibold text-sky-700 ring-1 ring-inset ring-sky-200/70 dark:bg-sky-500/10 dark:text-sky-200 dark:ring-sky-400/20"
                >
                  {{ t('admin.requestLogs.diagnostics.retryRecovered') }}
                </span>
                <span
                  v-if="isPendingRequest(selectedLog)"
                  class="inline-flex items-center rounded-full bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-700 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/10 dark:text-amber-200 dark:ring-amber-400/20"
                >
                  {{ t('admin.requestLogs.diagnostics.pendingHint') }}
                </span>
                <span
                  v-if="selectedLog.claude_context_1m"
                  class="inline-flex items-center rounded-full bg-purple-50 px-2.5 py-1 text-xs font-semibold text-purple-700 ring-1 ring-inset ring-purple-200/70 dark:bg-purple-500/10 dark:text-purple-200 dark:ring-purple-400/20"
                  title="1M Context"
                >
                  1M
                </span>
              </div>
              <div class="space-y-1">
                <p class="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-dark-400">
                  {{ t('admin.requestLogs.diagnostics.requestId') }}
                </p>
                <p class="break-all font-mono text-sm font-semibold text-gray-950 dark:text-white">
                  {{ selectedLog.request_id }}
                </p>
              </div>
              <p class="max-w-3xl text-sm text-gray-600 dark:text-dark-300">
                {{ getDiagnosticsSummary(selectedLog) }}
              </p>
            </div>

            <div class="flex shrink-0 flex-wrap gap-2">
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-semibold text-gray-700 shadow-sm transition-colors hover:border-gray-300 hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-200 dark:hover:bg-white/[0.04]"
                @click="copyValue(selectedLog.request_id, t('admin.requestLogs.requestIdCopied'))"
              >
                <Icon name="copy" size="xs" :stroke-width="2" />
                {{ t('admin.requestLogs.diagnostics.copyRequestId') }}
              </button>
              <button
                v-if="selectedLog.session_id"
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-semibold text-gray-700 shadow-sm transition-colors hover:border-gray-300 hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-200 dark:hover:bg-white/[0.04]"
                @click="copyValue(selectedLog.session_id, t('admin.requestLogs.diagnostics.sessionCopied'))"
              >
                <Icon name="copy" size="xs" :stroke-width="2" />
                {{ t('admin.requestLogs.diagnostics.copySessionId') }}
              </button>
              <button
                v-if="selectedLog.session_id"
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg bg-orange-600 px-3 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-orange-700"
                @click="filterSelectedSession"
              >
                <Icon name="filter" size="xs" :stroke-width="2" />
                {{ t('admin.requestLogs.diagnostics.filterSession') }}
              </button>
            </div>
          </div>
        </section>

        <section v-if="getErrorSummary(selectedLog)" class="rounded-lg border border-rose-200 bg-rose-50 p-4 text-rose-900 dark:border-rose-400/20 dark:bg-rose-500/10 dark:text-rose-100">
          <div class="flex items-start gap-3">
            <Icon name="exclamationTriangle" size="sm" class="mt-0.5 shrink-0" :stroke-width="2" />
            <div class="min-w-0">
              <h4 class="text-sm font-semibold">{{ t('admin.requestLogs.diagnostics.finalError') }}</h4>
              <p class="mt-1 whitespace-pre-wrap break-words text-sm leading-6">
                {{ getErrorSummary(selectedLog) }}
              </p>
            </div>
          </div>
        </section>

        <div class="grid gap-4 xl:grid-cols-2">
          <section class="diagnostics-section">
            <div class="diagnostics-section-heading">
              <Icon name="clipboard" size="sm" />
              <h4>{{ t('admin.requestLogs.diagnostics.overview') }}</h4>
            </div>
            <dl class="diagnostics-grid">
              <div v-for="field in overviewFields" :key="field.key" class="diagnostics-field" :class="{ 'md:col-span-2': field.wide }">
                <dt>{{ field.label }}</dt>
                <dd :class="{ 'font-mono': field.mono, 'whitespace-pre-wrap': field.multiline }">{{ field.value }}</dd>
              </div>
            </dl>
          </section>

          <section class="diagnostics-section">
            <div class="diagnostics-section-heading">
              <Icon name="userCircle" size="sm" />
              <h4>{{ t('admin.requestLogs.diagnostics.identity') }}</h4>
            </div>
            <dl class="diagnostics-grid">
              <div v-for="field in identityFields" :key="field.key" class="diagnostics-field" :class="{ 'md:col-span-2': field.wide }">
                <dt>{{ field.label }}</dt>
                <dd :class="{ 'font-mono': field.mono, 'whitespace-pre-wrap': field.multiline }">{{ field.value }}</dd>
              </div>
            </dl>
          </section>

          <section class="diagnostics-section">
            <div class="diagnostics-section-heading">
              <Icon name="cpu" size="sm" />
              <h4>{{ t('admin.requestLogs.diagnostics.routing') }}</h4>
            </div>
            <dl class="diagnostics-grid">
              <div v-for="field in routingFields" :key="field.key" class="diagnostics-field" :class="{ 'md:col-span-2': field.wide }">
                <dt>{{ field.label }}</dt>
                <dd :class="{ 'font-mono': field.mono, 'whitespace-pre-wrap': field.multiline }">{{ field.value }}</dd>
              </div>
            </dl>
          </section>

          <section class="diagnostics-section">
            <div class="diagnostics-section-heading">
              <Icon name="chatBubble" size="sm" />
              <h4>{{ t('admin.requestLogs.diagnostics.session') }}</h4>
            </div>
            <dl class="diagnostics-grid">
              <div v-for="field in sessionFields" :key="field.key" class="diagnostics-field" :class="{ 'md:col-span-2': field.wide }">
                <dt>{{ field.label }}</dt>
                <dd :class="{ 'font-mono': field.mono, 'whitespace-pre-wrap': field.multiline }">{{ field.value }}</dd>
              </div>
            </dl>
          </section>
        </div>

        <section class="diagnostics-section">
          <div class="diagnostics-section-heading">
            <Icon name="chartBar" size="sm" />
            <h4>{{ t('admin.requestLogs.diagnostics.resourceBilling') }}</h4>
          </div>
          <p
            v-if="selectedLog.billable === false"
            class="mb-3 rounded-lg bg-gray-50 px-3 py-2 text-sm text-gray-600 ring-1 ring-inset ring-gray-200 dark:bg-white/[0.04] dark:text-dark-300 dark:ring-white/[0.08]"
            data-testid="diagnostics-non-billable-note"
          >
            {{ t('admin.requestLogs.diagnostics.nonBillableNote') }}
          </p>
          <p
            v-else-if="selectedLog.billable && selectedLog.actual_cost == null"
            class="mb-3 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-800 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/10 dark:text-amber-200 dark:ring-amber-400/20"
          >
            {{ t('admin.requestLogs.diagnostics.billableCostPendingNote') }}
          </p>
          <dl class="diagnostics-grid xl:grid-cols-4">
            <div v-for="field in resourceBillingFields" :key="field.key" class="diagnostics-field" :class="{ 'md:col-span-2': field.wide }">
              <dt>{{ field.label }}</dt>
              <dd :class="{ 'font-mono': field.mono, 'whitespace-pre-wrap': field.multiline }">{{ field.value }}</dd>
            </div>
          </dl>
        </section>

        <section class="diagnostics-section">
          <div class="diagnostics-section-heading">
            <Icon name="sync" size="sm" />
            <h4>{{ t('admin.requestLogs.diagnostics.upstreamTimeline') }}</h4>
          </div>

          <div v-if="selectedLog.upstream_attempts?.length" class="space-y-3" data-testid="diagnostics-upstream-timeline">
            <div
              v-for="(attempt, index) in selectedLog.upstream_attempts"
              :key="`${index}:${attempt.upstream_request_id || attempt.upstream_status_code || attempt.account_id || attempt.at_unix_ms || 'attempt'}`"
              class="relative pl-7"
            >
              <div
                class="absolute left-0 top-1 flex h-5 w-5 items-center justify-center rounded-full ring-4 ring-white dark:ring-dark-800"
                :class="getAttemptStatusClass(attempt)"
              >
                <span class="h-1.5 w-1.5 rounded-full bg-current" />
              </div>
              <div class="rounded-lg border border-gray-200 bg-white p-3 dark:border-white/[0.08] dark:bg-dark-900">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div class="flex min-w-0 flex-wrap items-center gap-2">
                    <span class="text-xs font-semibold text-gray-400 dark:text-dark-400">
                      #{{ index + 1 }}
                    </span>
                    <span class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold" :class="getAttemptBadgeClass(attempt)">
                      {{ formatAttemptStatus(attempt) }}
                    </span>
                    <span class="text-sm font-semibold text-gray-900 dark:text-white">
                      {{ formatAttemptTarget(attempt) }}
                    </span>
                  </div>
                  <span class="text-xs text-gray-500 dark:text-dark-400">
                    {{ formatAttemptTime(attempt.at_unix_ms) }}
                  </span>
                </div>
                <div class="mt-2 grid gap-2 text-xs text-gray-600 dark:text-dark-300 md:grid-cols-2">
                  <div v-if="attempt.upstream_request_id" class="min-w-0">
                    <span class="text-gray-400 dark:text-dark-500">{{ t('admin.requestLogs.diagnostics.upstreamRequestId') }}</span>
                    <p class="break-all font-mono">{{ attempt.upstream_request_id }}</p>
                  </div>
                  <div v-if="attempt.upstream_url" class="min-w-0">
                    <span class="text-gray-400 dark:text-dark-500">{{ t('admin.requestLogs.diagnostics.upstreamUrl') }}</span>
                    <p class="break-all font-mono">{{ attempt.upstream_url }}</p>
                  </div>
                </div>
                <p v-if="attempt.message || attempt.detail" class="mt-2 whitespace-pre-wrap break-words text-sm leading-6 text-gray-700 dark:text-dark-200">
                  {{ attempt.detail || attempt.message }}
                </p>
              </div>
            </div>
          </div>

          <div v-else class="rounded-lg border border-dashed border-gray-200 px-4 py-5 text-center text-sm text-gray-500 dark:border-white/[0.08] dark:text-dark-400">
            {{ t('admin.requestLogs.diagnostics.noUpstreamAttempts') }}
          </div>
        </section>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import TablePanel from '@/components/common/TablePanel.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import { useInfinitePagedList } from '@/composables/useInfinitePagedList'
import { adminAPI } from '@/api'
import type {
  AdminRequestLog,
  AdminRequestLogQueryParams,
  AdminRequestLogResponse,
  AdminRequestLogSort,
  AdminRequestLogSummary,
  AdminRequestLogUpstreamAttempt
} from '@/api/admin/requestRecords'
import type { SimpleApiKey, SimpleUser } from '@/api/admin/usage'
import { useAdminSettingsStore, useAppStore } from '@/stores'
import { formatDateTime, formatNumber, formatReasoningEffort } from '@/utils/format'
import { getBillingModeLabel } from '@/utils/billingMode'
import type { Column } from '@/components/common/types'

type TimeRangePreset = 'today' | 'yesterday' | 'this_week' | 'this_month' | 'custom'
type StatusPreset = '' | '200' | '!200' | '4xx' | '5xx' | '503' | '429' | 'custom'
type RequestLogColumn = Column & { key: RequestLogColumnKey }
type RequestLogColumnKey =
  | 'time'
  | 'user'
  | 'account'
  | 'session'
  | 'model'
  | 'tokens'
  | 'cache'
  | 'cost'
  | 'performance'
  | 'user_agent'
  | 'status'
  | 'api_key'
  | 'request_id'
  | 'ip_address'
  | 'platform'
  | 'group'
  | 'endpoint'

interface RequestLogFilters {
  time_range: TimeRangePreset | ''
  statusPreset: StatusPreset
  customStatusCode: string
  model: string
  user_id?: number
  api_key_id?: number
  account_id?: number
  group_id: string
  request_id: string
  session_id: string
  sort: AdminRequestLogSort
}

interface SimpleAccount {
  id: number
  name: string
}

interface SimpleGroup {
  id: number
  name: string
}

interface DiagnosticField {
  key: string
  label: string
  value: string
  mono?: boolean
  multiline?: boolean
  wide?: boolean
}

const { t } = useI18n()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()

const errorMessage = ref('')
const opsDisabled = ref(false)
const filtersOpen = ref(false)
const timeFiltersOpen = ref(true)
const resultFiltersOpen = ref(true)
const identityFiltersOpen = ref(true)
const requestFiltersOpen = ref(false)
const autoRefreshEnabled = ref(false)
const loadMoreSentinelRef = ref<HTMLElement | null>(null)
const summary = ref<AdminRequestLogSummary>({
  total_requests: 0,
  success_requests: 0,
  error_requests: 0,
  error_rate: 0
})

const filters = reactive<RequestLogFilters>({
  time_range: '',
  statusPreset: '',
  customStatusCode: '',
  model: '',
  group_id: '',
  request_id: '',
  session_id: '',
  sort: 'created_at_desc'
})

const customStartTime = ref('')
const customEndTime = ref('')
const userKeyword = ref('')
const apiKeyKeyword = ref('')
const accountKeyword = ref('')
const userResults = ref<SimpleUser[]>([])
const apiKeyResults = ref<SimpleApiKey[]>([])
const accountResults = ref<SimpleAccount[]>([])
const groupOptions = ref<SimpleGroup[]>([])
const showUserDropdown = ref(false)
const showApiKeyDropdown = ref(false)
const showAccountDropdown = ref(false)
const userSearchRef = ref<HTMLElement | null>(null)
const apiKeySearchRef = ref<HTMLElement | null>(null)
const accountSearchRef = ref<HTMLElement | null>(null)
const columnDropdownRef = ref<HTMLElement | null>(null)
const showColumnDropdown = ref(false)
const selectedLog = ref<AdminRequestLog | null>(null)

const {
  items: logs,
  loading,
  loadingMore,
  refreshingSilently,
  pagination,
  hasMore: hasMoreLogs,
  load: loadRequestLogs,
  reset: resetRequestLogs,
  refreshFirstPageSilently,
  startObserver: startLoadMoreObserver,
  stopObserver: stopLoadMoreObserver
} = useInfinitePagedList<AdminRequestLog, AdminRequestLogResponse>({
  pageSize: 20,
  sentinelRef: loadMoreSentinelRef,
  fetchPage: (pageState, options) => {
    errorMessage.value = ''
    return adminAPI.requestRecords.list(buildParams(pageState), options)
  },
  onSuccess: (result) => {
    summary.value = result.summary || createEmptyRequestLogSummary()
  },
  onError: (error) => {
    if (isOpsDisabledError(error)) {
      opsDisabled.value = true
      adminSettingsStore.setOpsMonitoringEnabledLocal(false)
      return
    }
    console.error('[RequestLogsView] Failed to load request logs', error)
    errorMessage.value = extractErrorMessage(error)
    appStore.showError(errorMessage.value)
  },
  isCanceled,
  canLoadMore: () => !opsDisabled.value
})

let userSearchTimeout: ReturnType<typeof setTimeout> | null = null
let apiKeySearchTimeout: ReturnType<typeof setTimeout> | null = null
let accountSearchTimeout: ReturnType<typeof setTimeout> | null = null
let autoRefreshTimer: ReturnType<typeof setInterval> | null = null

const timeRangeOptions = computed<Array<{ value: TimeRangePreset; label: string; icon: 'calendar' }>>(() => [
  { value: 'today', label: t('admin.dashboard.rangeToday'), icon: 'calendar' },
  { value: 'yesterday', label: t('admin.dashboard.rangeYesterday'), icon: 'calendar' },
  { value: 'this_week', label: t('admin.dashboard.rangeThisWeek'), icon: 'calendar' },
  { value: 'this_month', label: t('admin.dashboard.rangeThisMonth'), icon: 'calendar' },
  { value: 'custom', label: t('admin.requestLogs.custom'), icon: 'calendar' }
])

const statusCodeOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.requestLogs.allStatusCodes') },
  { value: '200', label: '200' },
  { value: '!200', label: t('admin.requestLogs.excludeStatus200') },
  { value: '4xx', label: '4xx' },
  { value: '5xx', label: '5xx' },
  { value: '503', label: '503' },
  { value: '429', label: '429' },
  { value: 'custom', label: t('admin.requestLogs.customStatusCode') }
])

const sortOptions = computed<SelectOption[]>(() => [
  { value: 'created_at_desc', label: t('admin.requestLogs.sortNewest') },
  { value: 'duration_desc', label: t('admin.requestLogs.sortDuration') }
])

const groupFilterOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.requestLogs.allGroups') },
  ...groupOptions.value.map((group) => ({
    value: String(group.id),
    label: group.name
  }))
])

const DEFAULT_VISIBLE_COLUMNS: RequestLogColumnKey[] = [
  'time',
  'user',
  'account',
  'session',
  'model',
  'tokens',
  'cache',
  'cost',
  'performance',
  'user_agent',
  'status'
]
const HIDDEN_COLUMNS_KEY = 'admin-request-logs-hidden-columns-v2'
const hiddenColumns = reactive<Set<RequestLogColumnKey>>(new Set())

const allColumns = computed<RequestLogColumn[]>(() => [
  { key: 'time', label: t('admin.requestLogs.table.time') },
  { key: 'user', label: t('admin.requestLogs.table.user') },
  { key: 'account', label: t('admin.requestLogs.table.account') },
  { key: 'session', label: t('admin.requestLogs.table.session') },
  { key: 'model', label: t('admin.requestLogs.table.billingModel') },
  { key: 'tokens', label: t('admin.requestLogs.table.tokens') },
  { key: 'cache', label: t('admin.requestLogs.table.cache') },
  { key: 'cost', label: t('admin.requestLogs.table.cost') },
  { key: 'performance', label: t('admin.requestLogs.table.performance') },
  { key: 'user_agent', label: t('admin.requestLogs.table.userAgent') },
  { key: 'status', label: t('admin.requestLogs.table.status') },
  { key: 'api_key', label: t('admin.requestLogs.table.apiKey') },
  { key: 'request_id', label: t('admin.requestLogs.table.requestId') },
  { key: 'ip_address', label: t('admin.requestLogs.table.sourceIp') },
  { key: 'platform', label: t('admin.requestLogs.table.platform') },
  { key: 'group', label: t('admin.requestLogs.table.group') },
  { key: 'endpoint', label: t('admin.requestLogs.table.inboundEndpoint') }
])

const toggleableColumns = computed<RequestLogColumn[]>(() => allColumns.value)
const visibleColumns = computed<RequestLogColumn[]>(() => allColumns.value.filter((column) => !hiddenColumns.has(column.key)))
const visibleColumnCount = computed(() => visibleColumns.value.length)
const allColumnKeys = computed(() => new Set(allColumns.value.map((column) => column.key)))
const loadedRecordsSummary = computed(() =>
  t('admin.requestLogs.loadedRecordsSummary', {
    count: formatNumber(logs.value.length),
    total: formatNumber(pagination.total)
  })
)

function isRequestLogColumnKey(value: string): value is RequestLogColumnKey {
  return allColumnKeys.value.has(value as RequestLogColumnKey)
}

function defaultHiddenColumns(): RequestLogColumnKey[] {
  return allColumns.value
    .map((column) => column.key)
    .filter((key): key is RequestLogColumnKey => !DEFAULT_VISIBLE_COLUMNS.includes(key))
}

function saveHiddenColumns() {
  try {
    localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
  } catch (error) {
    console.error('[RequestLogsView] Failed to save column visibility', error)
  }
}

function loadSavedColumns() {
  hiddenColumns.clear()
  const defaults = defaultHiddenColumns()
  try {
    const raw = localStorage.getItem(HIDDEN_COLUMNS_KEY)
    const parsed = raw ? JSON.parse(raw) : defaults
    const keys = Array.isArray(parsed) ? parsed : defaults
    keys.forEach((key) => {
      if (typeof key === 'string' && isRequestLogColumnKey(key)) {
        hiddenColumns.add(key)
      }
    })
  } catch {
    defaults.forEach((key) => hiddenColumns.add(key))
  }
}

function isColumnVisible(key: RequestLogColumnKey): boolean {
  return !hiddenColumns.has(key)
}

function toggleColumn(key: RequestLogColumnKey) {
  if (hiddenColumns.has(key)) {
    hiddenColumns.delete(key)
  } else {
    hiddenColumns.add(key)
  }
  saveHiddenColumns()
}

function resetColumns() {
  hiddenColumns.clear()
  defaultHiddenColumns().forEach((key) => hiddenColumns.add(key))
  saveHiddenColumns()
}

const requestLogStatsSummaryItems = computed(() => [
  `${formatNumber(summary.value.total_requests)} ${t('admin.requestLogs.totalRequests')}`,
  `${formatNumber(summary.value.success_requests)} ${t('admin.requestLogs.successRequests')}`,
  `${formatNumber(summary.value.error_requests)} ${t('admin.requestLogs.errorRequests')}`,
  `${(summary.value.error_rate * 100).toFixed(2)}% ${t('admin.requestLogs.errorRate')}`
])

function createEmptyRequestLogSummary(): AdminRequestLogSummary {
  return {
    total_requests: 0,
    success_requests: 0,
    error_requests: 0,
    error_rate: 0
  }
}

const activeFilterCount = computed(() => {
  let count = 0
  if (filters.time_range || customStartTime.value || customEndTime.value) count += 1
  if (filters.statusPreset) count += 1
  if (filters.model) count += 1
  if (filters.user_id) count += 1
  if (filters.api_key_id) count += 1
  if (filters.account_id) count += 1
  if (filters.group_id) count += 1
  if (filters.request_id) count += 1
  if (filters.session_id) count += 1
  return count
})

const overviewFields = computed<DiagnosticField[]>(() => {
  const log = selectedLog.value
  if (!log) return []
  return [
    field('outcome', t('admin.requestLogs.diagnostics.finalOutcome'), getFinalStatusLabel(log)),
    field('status_code', t('admin.requestLogs.diagnostics.clientStatusCode'), formatStatusCode(log.status_code), true),
    field('created_at', t('admin.requestLogs.diagnostics.createdAt'), formatDateValue(log.created_at)),
    field('completed_at', t('admin.requestLogs.diagnostics.completedAt'), formatDateValue(log.completed_at)),
    field('duration_ms', t('admin.requestLogs.diagnostics.duration'), formatDuration(log.duration_ms), true),
    field('first_token_ms', t('admin.requestLogs.diagnostics.firstToken'), formatDuration(log.first_token_ms), true),
    field('request_type', t('admin.requestLogs.diagnostics.requestType'), formatRequestType(log)),
    field('stream', t('admin.requestLogs.diagnostics.stream'), formatBoolean(log.stream)),
    field('client_request_id', t('admin.requestLogs.diagnostics.clientRequestId'), valueOrDash(log.client_request_id), true, { wide: true }),
  ]
})

const identityFields = computed<DiagnosticField[]>(() => {
  const log = selectedLog.value
  if (!log) return []
  return [
    field('user', t('admin.requestLogs.diagnostics.user'), formatIdentity(log.user, log.user_id), true),
    field('api_key', t('admin.requestLogs.diagnostics.apiKey'), formatIdentity(log.api_key, log.api_key_id), true),
    field('account', t('admin.requestLogs.diagnostics.account'), formatIdentity(log.account, log.account_id), true),
    field('group', t('admin.requestLogs.diagnostics.group'), formatIdentity(log.group, log.group_id), true),
    field('ip_address', t('admin.requestLogs.diagnostics.ipAddress'), valueOrDash(log.ip_address), true),
    field('user_agent', t('admin.requestLogs.diagnostics.userAgent'), valueOrDash(log.user_agent), false, { wide: true, multiline: true }),
  ]
})

const routingFields = computed<DiagnosticField[]>(() => {
  const log = selectedLog.value
  if (!log) return []
  return [
    field('platform', t('admin.requestLogs.diagnostics.platform'), valueOrDash(log.platform)),
    field('model', t('admin.requestLogs.diagnostics.model'), valueOrDash(log.model), true),
    field('requested_model', t('admin.requestLogs.diagnostics.requestedModel'), valueOrDash(log.requested_model), true),
    field('upstream_model', t('admin.requestLogs.diagnostics.upstreamModel'), valueOrDash(log.upstream_model), true),
    field('reasoning_effort', t('admin.requestLogs.diagnostics.reasoningEffort'), formatReasoningEffort(log.reasoning_effort)),
    field('service_tier', t('admin.requestLogs.diagnostics.serviceTier'), valueOrDash(log.service_tier)),
    field('inbound_endpoint', t('admin.requestLogs.diagnostics.inboundEndpoint'), valueOrDash(log.inbound_endpoint), true, { wide: true }),
    field('upstream_endpoint', t('admin.requestLogs.diagnostics.upstreamEndpoint'), valueOrDash(log.upstream_endpoint), true, { wide: true }),
    field('model_mapping_chain', t('admin.requestLogs.diagnostics.modelMappingChain'), valueOrDash(log.model_mapping_chain), false, { wide: true, multiline: true }),
  ]
})

const sessionFields = computed<DiagnosticField[]>(() => {
  const log = selectedLog.value
  if (!log) return []
  return [
    field('session_id', t('admin.requestLogs.diagnostics.sessionId'), valueOrDash(log.session_id), true, { wide: true }),
    field('session_source', t('admin.requestLogs.diagnostics.sessionSource'), valueOrDash(log.session_source)),
    field('client_session_id', t('admin.requestLogs.diagnostics.clientSessionId'), valueOrDash(log.client_session_id), true, { wide: true }),
  ]
})

const resourceBillingFields = computed<DiagnosticField[]>(() => {
  const log = selectedLog.value
  if (!log) return []
  return [
    field('billable', t('admin.requestLogs.diagnostics.billable'), log.billable ? t('usage.billable') : t('usage.nonBillable')),
    field('billing_mode', t('admin.requestLogs.diagnostics.billingMode'), getBillingModeLabel(log.billing_mode, t)),
    field('input_tokens', t('admin.requestLogs.diagnostics.inputTokens'), formatTokenCount(log.input_tokens), true),
    field('output_tokens', t('admin.requestLogs.diagnostics.outputTokens'), formatTokenCount(log.output_tokens), true),
    field('cache_creation_tokens', t('admin.requestLogs.diagnostics.cacheCreationTokens'), formatTokenCount(log.cache_creation_tokens), true),
    field('cache_read_tokens', t('admin.requestLogs.diagnostics.cacheReadTokens'), formatTokenCount(log.cache_read_tokens), true),
    field('image_count', t('admin.requestLogs.diagnostics.imageCount'), formatTokenCount(log.image_count), true),
    field('cache_ttl_overridden', t('admin.requestLogs.diagnostics.cacheTtlOverridden'), formatBoolean(log.cache_ttl_overridden)),
    field('claude_context_1m', t('admin.requestLogs.diagnostics.claudeContext1M'), formatBoolean(log.claude_context_1m)),
    field('total_cost', t('admin.requestLogs.diagnostics.totalCost'), formatCost(log.total_cost), true),
    field('actual_cost', t('admin.requestLogs.diagnostics.actualCost'), formatCost(log.actual_cost), true),
    field('account_stats_cost', t('admin.requestLogs.diagnostics.accountStatsCost'), formatCost(log.account_stats_cost), true),
    field('rate_multiplier', t('admin.requestLogs.diagnostics.rateMultiplier'), formatMultiplierValue(log.rate_multiplier), true),
    field('account_rate_multiplier', t('admin.requestLogs.diagnostics.accountRateMultiplier'), formatMultiplierValue(log.account_rate_multiplier), true),
  ]
})

function field(
  key: string,
  label: string,
  value: unknown,
  mono = false,
  options: Pick<DiagnosticField, 'multiline' | 'wide'> = {},
): DiagnosticField {
  return {
    key,
    label,
    value: valueOrDash(value),
    mono,
    ...options,
  }
}

function valueOrDash(value: unknown): string {
  if (value == null) return '-'
  if (typeof value === 'string') {
    const normalized = value.trim()
    return normalized || '-'
  }
  return String(value)
}

function formatDateValue(value?: string | null): string {
  return value ? valueOrDash(formatDateTime(value)) : '-'
}

function formatStatusCode(statusCode?: number | null): string {
  if (statusCode == null || statusCode <= 0) return '-'
  return String(statusCode)
}

function formatDuration(ms?: number | null): string {
  if (ms == null) return '-'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatTokenCount(value?: number | null): string {
  if (value == null) return '-'
  return value.toLocaleString()
}

function formatCost(value?: number | null): string {
  if (value == null) return '-'
  return `$${value.toFixed(6)}`
}

function formatMultiplierValue(value?: number | null): string {
  if (value == null) return '-'
  return `${value.toFixed(2)}x`
}

function formatBoolean(value?: boolean | null): string {
  if (value == null) return '-'
  return value ? t('admin.requestLogs.diagnostics.yes') : t('admin.requestLogs.diagnostics.no')
}

function formatIdentity(
  identity: AdminRequestLog['user'] | AdminRequestLog['api_key'] | AdminRequestLog['account'] | AdminRequestLog['group'],
  id?: number | null,
): string {
  const name = identity?.username?.trim() || identity?.name?.trim() || identity?.email?.trim()
  const parts = [
    name,
    identity?.email && identity.email !== name ? identity.email : '',
    id || identity?.id ? `#${id || identity?.id}` : '',
  ].filter(Boolean)
  return parts.join(' · ') || '-'
}

function formatRequestType(log: AdminRequestLog): string {
  if (log.request_type === 'stream') return t('admin.requestLogs.requestTypes.stream')
  if (log.request_type === 'ws_v2') return t('admin.requestLogs.requestTypes.ws_v2')
  if (log.request_type === 'sync') return t('admin.requestLogs.requestTypes.sync')
  return t('admin.requestLogs.requestTypes.unknown')
}

function isPendingRequest(log: AdminRequestLog): boolean {
  return log.outcome === 'pending' || log.kind === 'pending'
}

function isFinalSuccess(log: AdminRequestLog): boolean {
  if (isPendingRequest(log)) return false
  if (log.outcome === 'success' || log.outcome === 'non_billable' || log.kind === 'success') return true
  return log.status_code >= 200 && log.status_code < 300
}

function isFinalError(log: AdminRequestLog): boolean {
  if (isPendingRequest(log) || isFinalSuccess(log)) return false
  return log.outcome === 'error' || log.kind === 'error' || log.status_code >= 400
}

function getFinalStatusLabel(log: AdminRequestLog): string {
  if (isPendingRequest(log)) return t('usage.statusPending')
  if (log.status_code > 0) return String(log.status_code)
  if (isFinalError(log)) return t('usage.statusFailed')
  return t('usage.statusSuccess')
}

function getFinalStatusClass(log: AdminRequestLog): string {
  if (isPendingRequest(log)) {
    return 'bg-amber-50 text-amber-700 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/10 dark:text-amber-200 dark:ring-amber-400/20'
  }
  if (isFinalSuccess(log)) {
    return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-200'
  }
  if (log.status_code >= 500) {
    return 'bg-rose-100 text-rose-800 dark:bg-rose-500/15 dark:text-rose-200'
  }
  return 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-200'
}

function hasFailedUpstreamAttempt(log: AdminRequestLog): boolean {
  return Boolean(log.upstream_attempts?.some((attempt) => {
    const code = attempt.upstream_status_code
    if (code != null && code >= 400) return true
    const kind = attempt.kind?.toLowerCase()
    return kind === 'error' || kind === 'failed'
  }))
}

function getDiagnosticsSummary(log: AdminRequestLog): string {
  if (isPendingRequest(log)) return t('admin.requestLogs.diagnostics.summaryPending')
  if (isFinalSuccess(log) && hasFailedUpstreamAttempt(log)) {
    return t('admin.requestLogs.diagnostics.summaryRecovered')
  }
  if (isFinalSuccess(log)) return t('admin.requestLogs.diagnostics.summarySuccess')
  return t('admin.requestLogs.diagnostics.summaryError')
}

function getErrorSummary(log: AdminRequestLog): string {
  if (!isFinalError(log)) return ''
  return (
    log.error_message?.trim() ||
    log.message?.trim() ||
    log.phase?.trim() ||
    ''
  )
}

function isAttemptSuccess(attempt: AdminRequestLogUpstreamAttempt): boolean {
  const code = attempt.upstream_status_code
  if (code != null) return code >= 200 && code < 300
  const kind = attempt.kind?.toLowerCase()
  return kind === 'success'
}

function formatAttemptStatus(attempt: AdminRequestLogUpstreamAttempt): string {
  if (attempt.upstream_status_code != null) return String(attempt.upstream_status_code)
  if (attempt.kind) return attempt.kind
  return '-'
}

function getAttemptBadgeClass(attempt: AdminRequestLogUpstreamAttempt): string {
  if (isAttemptSuccess(attempt)) {
    return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-200'
  }
  const code = attempt.upstream_status_code
  if (code != null && code >= 500) {
    return 'bg-rose-100 text-rose-800 dark:bg-rose-500/15 dark:text-rose-200'
  }
  if (code != null && code >= 400) {
    return 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-200'
  }
  return 'bg-gray-100 text-gray-700 dark:bg-white/[0.06] dark:text-dark-200'
}

function getAttemptStatusClass(attempt: AdminRequestLogUpstreamAttempt): string {
  if (isAttemptSuccess(attempt)) return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-300'
  const code = attempt.upstream_status_code
  if (code != null && code >= 500) return 'bg-rose-50 text-rose-600 dark:bg-rose-500/10 dark:text-rose-300'
  if (code != null && code >= 400) return 'bg-amber-50 text-amber-600 dark:bg-amber-500/10 dark:text-amber-300'
  return 'bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300'
}

function formatAttemptTarget(attempt: AdminRequestLogUpstreamAttempt): string {
  const account = attempt.account_name?.trim() || (attempt.account_id ? `#${attempt.account_id}` : '')
  const platform = attempt.platform?.trim()
  return [platform, account].filter(Boolean).join(' · ') || '-'
}

function formatAttemptTime(value?: number): string {
  if (!value) return '-'
  return formatDateValue(new Date(value).toISOString())
}

function openDiagnostics(row: unknown) {
  selectedLog.value = row as AdminRequestLog
}

function closeDiagnostics() {
  selectedLog.value = null
}

async function copyValue(value: string | undefined, successMessage: string) {
  if (!value) return
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(value)
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = value
      textarea.style.cssText = 'position:fixed;left:-9999px;top:-9999px'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    appStore.showSuccess(successMessage)
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

function filterSelectedSession() {
  const sessionId = selectedLog.value?.session_id
  if (!sessionId) return
  closeDiagnostics()
  filterBySession(sessionId)
}

function toRFC3339(localValue: string): string | undefined {
  if (!localValue) return undefined
  const d = new Date(localValue)
  if (Number.isNaN(d.getTime())) return undefined
  return d.toISOString()
}

function toDateTimeLocalValue(date: Date): string {
  const offsetMs = date.getTimezoneOffset() * 60 * 1000
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16)
}

function startOfLocalDay(date: Date): Date {
  const out = new Date(date)
  out.setHours(0, 0, 0, 0)
  return out
}

function endOfLocalDay(date: Date): Date {
  const out = new Date(date)
  out.setHours(23, 59, 59, 999)
  return out
}

function startOfLocalWeek(date: Date): Date {
  const out = startOfLocalDay(date)
  const day = out.getDay()
  const diff = day === 0 ? -6 : 1 - day
  out.setDate(out.getDate() + diff)
  return out
}

function startOfLocalMonth(date: Date): Date {
  const out = startOfLocalDay(date)
  out.setDate(1)
  return out
}

function getPresetDateRange(range: Exclude<TimeRangePreset, 'custom'>, now: Date): { start: Date; end: Date } {
  const end = endOfLocalDay(now)

  if (range === 'today') {
    return { start: startOfLocalDay(now), end }
  }
  if (range === 'yesterday') {
    const yesterday = new Date(now)
    yesterday.setDate(yesterday.getDate() - 1)
    return { start: startOfLocalDay(yesterday), end: endOfLocalDay(yesterday) }
  }
  if (range === 'this_week') {
    return { start: startOfLocalWeek(now), end }
  }
  if (range === 'this_month') {
    return { start: startOfLocalMonth(now), end }
  }
  return { start: startOfLocalDay(now), end }
}

function buildParams(pageState: { page: number; page_size: number }): AdminRequestLogQueryParams {
  const params: AdminRequestLogQueryParams = {
    page: pageState.page,
    page_size: pageState.page_size,
    sort: filters.sort
  }
  const selectedTimeRange = filters.time_range

  if (customStartTime.value || customEndTime.value) {
    const startTime = toRFC3339(customStartTime.value)
    const endTime = toRFC3339(customEndTime.value)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
  } else if (selectedTimeRange === 'yesterday') {
    const { start, end } = getPresetDateRange('yesterday', new Date())
    params.start_time = start.toISOString()
    params.end_time = end.toISOString()
  } else if (selectedTimeRange && selectedTimeRange !== 'custom') {
    params.time_range = selectedTimeRange
  }

  if (filters.statusPreset === '!200') {
    params.status_codes_other = true
  } else if (filters.statusPreset === 'custom') {
    const code = filters.customStatusCode.trim()
    if (code) params.status_codes = code
  } else if (filters.statusPreset) {
    params.status_codes = filters.statusPreset
  }
  if (filters.model.trim()) params.model = filters.model.trim()
  if (filters.user_id) params.user_id = filters.user_id
  if (filters.api_key_id) params.api_key_id = filters.api_key_id
  if (filters.account_id) params.account_id = filters.account_id
  if (filters.group_id) params.group_id = Number(filters.group_id)
  if (filters.request_id.trim()) params.request_id = filters.request_id.trim()
  if (filters.session_id.trim()) params.session_id = filters.session_id.trim()

  return params
}

function isCanceled(error: unknown): boolean {
  return !!error && typeof error === 'object' && 'code' in error && (error as { code?: string }).code === 'ERR_CANCELED'
}

function extractErrorMessage(error: unknown): string {
  if (error && typeof error === 'object' && 'message' in error) {
    const message = String((error as { message?: unknown }).message || '')
    if (message) return message
  }
  return t('admin.requestLogs.loadFailed')
}

function isOpsDisabledError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false
  const record = error as { code?: unknown; message?: unknown; status?: unknown }
  return record.code === 'OPS_DISABLED' || record.message === 'Ops monitoring is disabled'
}

async function loadGroupOptions() {
  try {
    const res = await adminAPI.groups.list(1, 1000)
    groupOptions.value = (res.items || []).map((group: any) => ({
      id: Number(group.id),
      name: group.name || `#${group.id}`
    }))
  } catch (error) {
    console.warn('[RequestLogsView] Failed to load group options', error)
  }
}

function applyFilters() {
  void resetRequestLogs()
}

function refreshData() {
  void resetRequestLogs()
}

function stopAutoRefresh() {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

function startAutoRefresh() {
  stopAutoRefresh()
  autoRefreshTimer = setInterval(() => {
    if (!loading.value && !loadingMore.value && !refreshingSilently.value && !opsDisabled.value) {
      void refreshFirstPageSilently()
    }
  }, 5000)
}

function toggleAutoRefresh() {
  autoRefreshEnabled.value = !autoRefreshEnabled.value
  if (autoRefreshEnabled.value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

function filterBySession(sessionId: string) {
  const normalized = sessionId.trim()
  if (!normalized) return
  filters.session_id = normalized
  void resetRequestLogs()
}

function resetFilters() {
  filters.time_range = ''
  filters.statusPreset = ''
  filters.customStatusCode = ''
  filters.model = ''
  filters.user_id = undefined
  filters.api_key_id = undefined
  filters.account_id = undefined
  filters.group_id = ''
  filters.request_id = ''
  filters.session_id = ''
  filters.sort = 'created_at_desc'
  customStartTime.value = ''
  customEndTime.value = ''
  userKeyword.value = ''
  apiKeyKeyword.value = ''
  accountKeyword.value = ''
  userResults.value = []
  apiKeyResults.value = []
  accountResults.value = []
  applyFilters()
}

function selectTimeRange(range: TimeRangePreset) {
  filters.time_range = range
  if (range === 'custom') return

  const { start, end } = getPresetDateRange(range, new Date())
  customStartTime.value = toDateTimeLocalValue(start)
  customEndTime.value = toDateTimeLocalValue(end)
}

function onCustomTimeChange() {
  filters.time_range = customStartTime.value || customEndTime.value ? 'custom' : ''
}

function debounceUserSearch() {
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  userSearchTimeout = setTimeout(async () => {
    if (!userKeyword.value.trim()) {
      userResults.value = []
      return
    }
    try {
      userResults.value = await adminAPI.usage.searchUsers(userKeyword.value.trim())
    } catch {
      userResults.value = []
    }
  }, 300)
}

function debounceApiKeySearch() {
  if (apiKeySearchTimeout) clearTimeout(apiKeySearchTimeout)
  apiKeySearchTimeout = setTimeout(async () => {
    try {
      apiKeyResults.value = await adminAPI.usage.searchApiKeys(filters.user_id, apiKeyKeyword.value.trim())
    } catch {
      apiKeyResults.value = []
    }
  }, 300)
}

function debounceAccountSearch() {
  if (accountSearchTimeout) clearTimeout(accountSearchTimeout)
  accountSearchTimeout = setTimeout(async () => {
    if (!accountKeyword.value.trim()) {
      accountResults.value = []
      return
    }
    try {
      const res = await adminAPI.accounts.list(1, 20, { search: accountKeyword.value.trim() })
      accountResults.value = (res.items || []).map((account: any) => ({
        id: Number(account.id),
        name: account.name || `#${account.id}`
      }))
    } catch {
      accountResults.value = []
    }
  }, 300)
}

function getUserDisplayName(user: SimpleUser): string {
  return user.username?.trim() || user.email || `#${user.id}`
}

async function selectUser(user: SimpleUser) {
  filters.user_id = user.id
  userKeyword.value = getUserDisplayName(user)
  showUserDropdown.value = false
  clearApiKey()
  try {
    apiKeyResults.value = await adminAPI.usage.searchApiKeys(user.id, '')
  } catch {
    apiKeyResults.value = []
  }
}

function clearUser() {
  filters.user_id = undefined
  userKeyword.value = ''
  userResults.value = []
  showUserDropdown.value = false
  clearApiKey()
}

function selectApiKey(key: SimpleApiKey) {
  filters.api_key_id = key.id
  apiKeyKeyword.value = key.name || `#${key.id}`
  showApiKeyDropdown.value = false
}

function clearApiKey() {
  filters.api_key_id = undefined
  apiKeyKeyword.value = ''
  apiKeyResults.value = []
  showApiKeyDropdown.value = false
}

function onApiKeyFocus() {
  showApiKeyDropdown.value = true
  if (apiKeyResults.value.length === 0) debounceApiKeySearch()
}

function selectAccount(account: SimpleAccount) {
  filters.account_id = account.id
  accountKeyword.value = account.name || `#${account.id}`
  showAccountDropdown.value = false
}

function clearAccount() {
  filters.account_id = undefined
  accountKeyword.value = ''
  accountResults.value = []
  showAccountDropdown.value = false
}

function onDocumentClick(event: MouseEvent) {
  const target = event.target as Node | null
  if (!target) return
  if (!(userSearchRef.value?.contains(target) ?? false)) showUserDropdown.value = false
  if (!(apiKeySearchRef.value?.contains(target) ?? false)) showApiKeyDropdown.value = false
  if (!(accountSearchRef.value?.contains(target) ?? false)) showAccountDropdown.value = false
  if (!(columnDropdownRef.value?.contains(target) ?? false)) showColumnDropdown.value = false
}

onMounted(async () => {
  loadSavedColumns()
  document.addEventListener('click', onDocumentClick)
  await adminSettingsStore.fetch()
  opsDisabled.value = !adminSettingsStore.opsMonitoringEnabled
  if (opsDisabled.value) return
  void loadGroupOptions()
  await loadRequestLogs()
  startLoadMoreObserver()
})

onBeforeUnmount(() => {
  stopLoadMoreObserver()
  stopAutoRefresh()
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  if (apiKeySearchTimeout) clearTimeout(apiKeySearchTimeout)
  if (accountSearchTimeout) clearTimeout(accountSearchTimeout)
  document.removeEventListener('click', onDocumentClick)
})
</script>

<style scoped>
.diagnostics-section {
  @apply rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-white/[0.08] dark:bg-dark-900;
}

.diagnostics-section-heading {
  @apply mb-3 flex items-center gap-2 text-sm font-semibold text-gray-950 dark:text-white;
}

.diagnostics-section-heading svg {
  @apply shrink-0 text-gray-400 dark:text-dark-400;
}

.diagnostics-grid {
  @apply grid gap-3 md:grid-cols-2;
}

.diagnostics-field {
  @apply min-w-0 rounded-lg bg-gray-50 px-3 py-2 ring-1 ring-inset ring-gray-100 dark:bg-white/[0.035] dark:ring-white/[0.06];
}

.diagnostics-field dt {
  @apply text-xs font-medium text-gray-500 dark:text-dark-400;
}

.diagnostics-field dd {
  @apply mt-1 break-words text-sm font-medium text-gray-900 dark:text-dark-100;
  overflow-wrap: anywhere;
}

</style>
