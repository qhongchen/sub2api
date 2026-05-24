<template>
  <AppLayout>
    <div class="space-y-5">
      <section class="cch-panel-card overflow-hidden">
        <div class="flex items-center justify-between gap-3 border-b border-gray-200/70 px-4 py-3 dark:border-white/[0.06]">
          <div class="flex min-w-0 items-center gap-2">
            <Icon name="bolt" size="sm" class="text-orange-500" :stroke-width="2" />
            <h2 class="text-base font-semibold text-gray-950 dark:text-white">
              {{ t('usage.activeSessions') }}
            </h2>
            <span class="truncate text-sm text-gray-500 dark:text-dark-400">
              {{ t('usage.activeSessionsSummary', { count: activeSessionRows.length, minutes: 5 }) }}
            </span>
          </div>
          <button
            type="button"
            class="inline-flex items-center gap-1 text-xs font-medium text-gray-500 transition-colors hover:text-gray-950 dark:text-dark-300 dark:hover:text-white"
            @click="filtersOpen = true"
          >
            {{ t('usage.viewAll') }}
            <Icon name="arrowRight" size="xs" :stroke-width="2" />
          </button>
        </div>

        <div v-if="loading && usageLogs.length === 0" class="divide-y divide-gray-100 dark:divide-white/[0.06]">
          <div v-for="idx in 2" :key="idx" class="flex items-center gap-3 px-4 py-3">
            <div class="h-4 w-4 animate-pulse rounded-full bg-gray-200 dark:bg-white/[0.08]" />
            <div class="h-4 flex-1 animate-pulse rounded bg-gray-200 dark:bg-white/[0.08]" />
            <div class="h-4 w-24 animate-pulse rounded bg-gray-200 dark:bg-white/[0.08]" />
          </div>
        </div>
        <div v-else-if="activeSessionRows.length" class="divide-y divide-gray-100 dark:divide-white/[0.06]">
          <div
            v-for="row in activeSessionRows"
            :key="row.id"
            class="flex items-center gap-3 px-4 py-3 text-sm"
          >
            <Icon name="checkCircle" size="sm" class="flex-shrink-0 text-emerald-500" :stroke-width="2" />
            <Icon name="user" size="xs" class="hidden flex-shrink-0 text-gray-500 dark:text-dark-400 sm:block" />
            <span class="min-w-0 flex-1 truncate font-semibold text-gray-950 dark:text-white">
              {{ row.api_key?.name || t('usage.unknownKey') }}
            </span>
            <span class="hidden items-center gap-1 text-gray-500 dark:text-dark-400 sm:inline-flex">
              <Icon name="key" size="xs" />
              {{ row.model || '-' }}
            </span>
            <span class="hidden max-w-[220px] truncate text-gray-500 dark:text-dark-400 md:inline">
              @ {{ formatUsageEndpoints(row) }}
            </span>
            <span class="ml-auto inline-flex items-center gap-1 text-gray-500 dark:text-dark-400">
              <Icon name="clock" size="xs" />
              {{ formatRelativeTime(row.created_at) }}
            </span>
          </div>
        </div>
        <div v-else class="flex min-h-[88px] items-center justify-center px-4 py-6 text-sm text-gray-500 dark:text-dark-400">
          {{ t('usage.noActiveSessions') }}
        </div>
      </section>

      <section class="space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200/70 bg-white/70 px-3 py-2 text-sm font-medium text-gray-600 shadow-sm transition-colors hover:border-gray-300 hover:text-gray-950 dark:border-white/[0.08] dark:bg-dark-900/60 dark:text-dark-300 dark:hover:text-white"
            @click="filtersOpen = !filtersOpen"
          >
            <Icon name="filter" size="sm" :stroke-width="2" />
            <span>{{ t('usage.filterCriteria') }}</span>
            <span
              v-if="activeFilterCount > 0"
              class="rounded-full bg-primary-50 px-1.5 py-0.5 text-[10px] font-semibold text-primary-600 dark:bg-primary-500/10 dark:text-primary-300"
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

          <div class="ml-auto flex items-center gap-1">
            <div class="relative">
              <button
                type="button"
                class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-800 shadow-sm transition-colors hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-100 dark:hover:bg-white/[0.04]"
                :title="t('usage.columnSettings')"
                @click="columnDropdownOpen = !columnDropdownOpen"
              >
                <Icon name="grid" size="sm" />
                <span>{{ visibleLogColumnCount }}/{{ logColumnOptions.length }}</span>
              </button>
              <div
                v-if="columnDropdownOpen"
                class="absolute right-0 top-full z-50 mt-2 w-52 rounded-lg border border-gray-200 bg-white p-1 shadow-lg dark:border-dark-700 dark:bg-dark-800"
              >
                <button
                  v-for="column in logColumnOptions"
                  :key="column.key"
                  type="button"
                  class="flex w-full items-center justify-between rounded-md px-3 py-2 text-left text-sm text-gray-600 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
                  @click="toggleLogColumn(column.key)"
                >
                  <span>{{ column.label }}</span>
                  <Icon v-if="isLogColumnVisible(column.key)" name="check" size="sm" class="text-primary-500" :stroke-width="2" />
                </button>
              </div>
            </div>

            <button
              type="button"
              class="consumer-icon-btn"
              :title="t('common.refresh')"
              :disabled="loading"
              @click="refreshData"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
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
            <div class="mb-5 flex flex-wrap items-center gap-2">
              <button
                v-for="preset in quickDatePresets"
                :key="preset.key"
                type="button"
                class="inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-sm font-semibold shadow-sm transition-colors"
                :class="activePreset === preset.key
                  ? 'border-orange-500 bg-orange-500 text-white'
                  : 'border-gray-200 bg-white text-gray-800 hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-100 dark:hover:bg-white/[0.04]'"
                @click="applyDatePreset(preset.key)"
              >
                <Icon :name="preset.icon" size="sm" />
                {{ preset.label }}
              </button>
            </div>

            <div class="grid gap-4 lg:grid-cols-2">
              <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]">
                <button
                  type="button"
                  class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                  @click="timeFiltersOpen = !timeFiltersOpen"
                >
                  <span class="flex items-center gap-3">
                    <span class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                      <Icon name="clock" size="sm" />
                    </span>
                    <span>
                      <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('usage.timeRange') }}</span>
                      <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('usage.timeRangeDescription') }}</span>
                    </span>
                  </span>
                  <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': timeFiltersOpen }" />
                </button>
                <div v-if="timeFiltersOpen" class="space-y-3 px-4 pb-4">
                  <div class="grid gap-3 sm:grid-cols-2">
                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.startDate') }}</span>
                      <input v-model="startDate" type="date" class="input" @change="syncDateRange" />
                    </label>
                    <label class="block">
                      <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.endDate') }}</span>
                      <input v-model="endDate" type="date" class="input" @change="syncDateRange" />
                    </label>
                  </div>
                </div>
              </div>

              <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]">
                <button
                  type="button"
                  class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                  @click="identityFiltersOpen = !identityFiltersOpen"
                >
                  <span class="flex items-center gap-3">
                    <span class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                      <Icon name="user" size="sm" />
                    </span>
                    <span>
                      <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('usage.identityInfo') }}</span>
                      <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('usage.identityDescription') }}</span>
                    </span>
                  </span>
                  <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': identityFiltersOpen }" />
                </button>
                <div v-if="identityFiltersOpen" class="grid gap-3 px-4 pb-4 sm:grid-cols-2">
                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.apiKeyFilter') }}</span>
                    <Select
                      v-model="apiKeyFilter"
                      :options="apiKeyOptions"
                      :placeholder="t('usage.allApiKeys')"
                    />
                  </label>
                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.model') }}</span>
                    <input
                      v-model="modelFilter"
                      type="text"
                      class="input"
                      :placeholder="t('usage.modelPlaceholder')"
                      @keyup.enter="applyFilters"
                    />
                  </label>
                </div>
              </div>

              <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]">
                <button
                  type="button"
                  class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                  @click="requestFiltersOpen = !requestFiltersOpen"
                >
                  <span class="flex items-center gap-3">
                    <span class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                      <Icon name="server" size="sm" />
                    </span>
                    <span>
                      <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('usage.requestParams') }}</span>
                      <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('usage.requestDescription') }}</span>
                    </span>
                  </span>
                  <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': requestFiltersOpen }" />
                </button>
                <div v-if="requestFiltersOpen" class="grid gap-3 px-4 pb-4 sm:grid-cols-2">
                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.type') }}</span>
                    <Select
                      v-model="requestTypeFilter"
                      :options="requestTypeOptions"
                      :placeholder="t('usage.allTypes')"
                    />
                  </label>
                  <label class="block">
                    <span class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.endpoint') }}</span>
                    <input
                      type="text"
                      class="input"
                      :value="t('usage.endpointFilterUnavailable')"
                      disabled
                    />
                  </label>
                </div>
              </div>

              <div class="rounded-xl border border-gray-200/70 bg-white/50 dark:border-white/[0.06] dark:bg-white/[0.02]">
                <button
                  type="button"
                  class="flex w-full items-center justify-between gap-3 px-4 py-3 text-left"
                  @click="statusFiltersOpen = !statusFiltersOpen"
                >
                  <span class="flex items-center gap-3">
                    <span class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-white/[0.06] dark:text-dark-300">
                      <Icon name="database" size="sm" />
                    </span>
                    <span>
                      <span class="block text-sm font-semibold text-gray-950 dark:text-white">{{ t('usage.statusInfo') }}</span>
                      <span class="block text-xs text-gray-500 dark:text-dark-400">{{ t('usage.statusDescription') }}</span>
                    </span>
                  </span>
                  <Icon name="chevronDown" size="sm" class="text-gray-400 transition-transform" :class="{ 'rotate-180': statusFiltersOpen }" />
                </button>
                <div v-if="statusFiltersOpen" class="px-4 pb-4 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('usage.statusFilterUnavailable') }}
                </div>
              </div>
            </div>

            <div class="mt-5 flex flex-wrap items-center gap-2">
              <button type="button" class="btn btn-primary" :disabled="loading" @click="applyFilters">
                {{ t('usage.applyFilters') }}
              </button>
              <button type="button" class="btn btn-secondary" @click="resetFilters">
                {{ t('common.reset') }}
              </button>
              <button type="button" class="btn btn-secondary" :disabled="exporting" @click="exportToCSV">
                <Icon name="download" size="sm" :class="{ 'animate-bounce': exporting }" />
                {{ exporting ? t('usage.exporting') : t('usage.exportCsv') }}
              </button>
            </div>
          </div>
        </Transition>
      </section>

      <section class="cch-panel-card overflow-hidden">
        <div class="flex items-center justify-between gap-3 border-b border-gray-200/70 px-4 py-3 dark:border-white/[0.06]">
          <span class="text-sm text-gray-500 dark:text-dark-400">
            {{ t('usage.loadedRecords', { count: usageLogs.length }) }}
          </span>
          <span class="text-xs text-gray-400 dark:text-dark-500">
            {{ usageStatsSummary }}
          </span>
        </div>

        <div v-if="!isDesktopTable" class="divide-y divide-gray-100 dark:divide-white/[0.06]">
          <div v-if="loading && usageLogs.length === 0" class="space-y-3 p-4">
            <div v-for="idx in 4" :key="idx" class="h-24 animate-pulse rounded-xl bg-gray-100 dark:bg-white/[0.06]" />
          </div>
          <div v-else-if="usageLogs.length === 0" class="flex min-h-[220px] flex-col items-center justify-center gap-3 p-6 text-center">
            <Icon name="inbox" size="xl" class="text-gray-300 dark:text-dark-500" />
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('usage.noRecords') }}</p>
          </div>
          <article
            v-for="row in usageLogs"
            v-else
            :key="row.id"
            class="space-y-3 p-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-gray-950 dark:text-white">{{ row.model || '-' }}</div>
                <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ formatRelativeTime(row.created_at) }} · {{ row.api_key?.name || t('usage.unknownKey') }}</div>
              </div>
              <span class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold" :class="getUsageStatusClass(row)">
                {{ getUsageStatusLabel(row) }}
              </span>
            </div>
            <div class="grid grid-cols-2 gap-3 text-sm">
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.tokens') }}</div>
                <div class="font-semibold text-gray-950 dark:text-white">{{ formatTokens(getTotalTokens(row)) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.cost') }}</div>
                <div class="font-semibold text-emerald-600 dark:text-emerald-300">${{ row.actual_cost.toFixed(6) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.requestId') }}</div>
                <div class="truncate font-mono text-xs text-gray-700 dark:text-dark-200">{{ shortRequestId(row.request_id) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.performance') }}</div>
                <div class="font-semibold text-gray-950 dark:text-white">{{ formatDuration(row.duration_ms) }}</div>
              </div>
            </div>
          </article>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[980px] divide-y divide-gray-200/70 dark:divide-white/[0.06]">
            <thead class="bg-gray-50/80 dark:bg-white/[0.03]">
              <tr>
                <th v-if="isLogColumnVisible('time')" class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.time') }}</th>
                <th v-if="isLogColumnVisible('api_key')" class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.apiKeyFilter') }}</th>
                <th v-if="isLogColumnVisible('request_id')" class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.requestId') }}</th>
                <th v-if="isLogColumnVisible('endpoint')" class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.endpoint') }}</th>
                <th v-if="isLogColumnVisible('model')" class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.billingModel') }}</th>
                <th v-if="isLogColumnVisible('tokens')" class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.tokens') }}</th>
                <th v-if="isLogColumnVisible('cache')" class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.cache') }}</th>
                <th v-if="isLogColumnVisible('cost')" class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.cost') }}</th>
                <th v-if="isLogColumnVisible('performance')" class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.performance') }}</th>
                <th v-if="isLogColumnVisible('status')" class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('usage.status') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white/40 dark:divide-white/[0.05] dark:bg-transparent">
              <template v-if="loading && usageLogs.length === 0">
                <tr v-for="idx in 5" :key="idx">
                  <td v-for="column in visibleLogColumns" :key="column.key" class="px-4 py-4">
                    <div class="h-4 w-24 animate-pulse rounded bg-gray-200 dark:bg-white/[0.08]" />
                  </td>
                </tr>
              </template>
              <tr v-else-if="usageLogs.length === 0">
                <td :colspan="visibleLogColumns.length" class="px-4 py-16 text-center">
                  <div class="flex flex-col items-center gap-3">
                    <Icon name="inbox" size="xl" class="text-gray-300 dark:text-dark-500" />
                    <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('usage.noRecords') }}</p>
                  </div>
                </td>
              </tr>
              <tr
                v-for="row in usageLogs"
                v-else
                :key="row.id"
                class="transition-colors hover:bg-orange-50/40 dark:hover:bg-white/[0.035]"
              >
                <td v-if="isLogColumnVisible('time')" class="whitespace-nowrap px-4 py-4 text-sm text-gray-700 dark:text-dark-200">
                  <span :title="formatDateTime(row.created_at)">{{ formatRelativeTime(row.created_at) }}</span>
                </td>
                <td v-if="isLogColumnVisible('api_key')" class="px-4 py-4">
                  <div class="flex min-w-0 items-center gap-2">
                    <Icon name="key" size="xs" class="flex-shrink-0 text-gray-400" />
                    <span class="max-w-[150px] truncate text-sm font-medium text-gray-950 dark:text-white">{{ row.api_key?.name || t('usage.unknownKey') }}</span>
                  </div>
                </td>
                <td v-if="isLogColumnVisible('request_id')" class="px-4 py-4">
                  <span class="font-mono text-xs text-gray-700 dark:text-dark-200" :title="row.request_id">
                    {{ shortRequestId(row.request_id) }}
                  </span>
                </td>
                <td v-if="isLogColumnVisible('endpoint')" class="px-4 py-4">
                  <span class="block max-w-[220px] truncate text-sm text-gray-600 dark:text-dark-300" :title="formatUsageEndpoints(row)">
                    {{ formatUsageEndpoints(row) }}
                  </span>
                </td>
                <td v-if="isLogColumnVisible('model')" class="px-4 py-4">
                  <div class="flex items-center gap-2">
                    <Icon name="cpu" size="xs" class="text-gray-500 dark:text-dark-400" />
                    <span class="max-w-[180px] truncate text-sm font-semibold text-gray-950 dark:text-white">{{ row.model || '-' }}</span>
                  </div>
                  <div v-if="row.reasoning_effort" class="mt-0.5 text-[10px] text-gray-500 dark:text-dark-400">
                    {{ formatReasoningEffort(row.reasoning_effort) }}
                  </div>
                </td>
                <td v-if="isLogColumnVisible('tokens')" class="px-4 py-4 text-right">
                  <div class="inline-flex items-center justify-end gap-1.5">
                    <div>
                      <div class="font-mono text-sm font-semibold text-gray-950 dark:text-white">{{ formatTokens(getTotalTokens(row)) }}</div>
                      <div class="text-xs text-gray-500 dark:text-dark-400">{{ formatTokens(row.output_tokens) }}</div>
                    </div>
                    <button
                      type="button"
                      class="rounded-full text-gray-400 transition-colors hover:text-blue-500"
                      @mouseenter="showTokenTooltip($event, row)"
                      @mouseleave="hideTokenTooltip"
                    >
                      <Icon name="infoCircle" size="xs" />
                    </button>
                  </div>
                </td>
                <td v-if="isLogColumnVisible('cache')" class="px-4 py-4 text-right">
                  <div class="font-mono text-sm text-gray-950 dark:text-white">
                    {{ getCacheTokens(row) > 0 ? formatTokens(getCacheTokens(row)) : '-' }}
                  </div>
                  <div v-if="row.cache_ttl_overridden" class="text-[10px] text-rose-500">TTL</div>
                </td>
                <td v-if="isLogColumnVisible('cost')" class="px-4 py-4 text-right">
                  <div class="inline-flex items-center justify-end gap-1.5">
                    <span class="font-mono text-sm font-semibold text-gray-950 dark:text-white">${{ row.actual_cost.toFixed(6) }}</span>
                    <button
                      type="button"
                      class="rounded-full text-gray-400 transition-colors hover:text-blue-500"
                      @mouseenter="showTooltip($event, row)"
                      @mouseleave="hideTooltip"
                    >
                      <Icon name="infoCircle" size="xs" />
                    </button>
                  </div>
                </td>
                <td v-if="isLogColumnVisible('performance')" class="px-4 py-4 text-right">
                  <div class="font-mono text-sm text-gray-950 dark:text-white">{{ formatDuration(row.duration_ms) }}</div>
                  <div class="text-xs text-gray-500 dark:text-dark-400">
                    TTFB {{ row.first_token_ms != null ? formatDuration(row.first_token_ms) : '-' }}
                  </div>
                </td>
                <td v-if="isLogColumnVisible('status')" class="px-4 py-4">
                  <span class="inline-flex rounded-full px-2.5 py-1 text-xs font-semibold" :class="getUsageStatusClass(row)">
                    {{ getUsageStatusLabel(row) }}
                  </span>
                  <div class="mt-1">
                    <span class="inline-flex rounded px-1.5 py-0.5 text-[10px] font-medium" :class="getRequestTypeBadgeClass(row)">
                      {{ getRequestTypeLabel(row) }}
                    </span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="pagination.total > 0" class="border-t border-gray-200/70 p-4 dark:border-white/[0.06]">
          <Pagination
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </section>
    </div>
  </AppLayout>

  <!-- Token Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tokenTooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: tokenTooltipPosition.x + 'px',
        top: tokenTooltipPosition.y + 'px'
      }"
    >
      <div
        class="whitespace-nowrap rounded-lg border border-gray-700 bg-gray-900 px-3 py-2.5 text-xs text-white shadow-xl dark:border-gray-600 dark:bg-gray-800"
      >
        <div class="space-y-1.5">
          <!-- Token Breakdown -->
          <div>
            <div class="text-xs font-semibold text-gray-300 mb-1">{{ t('usage.tokenDetails') }}</div>
            <div v-if="tokenTooltipData && tokenTooltipData.input_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.inputTokens') }}</span>
              <span class="font-medium text-white">{{ tokenTooltipData.input_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.output_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.outputTokens') }}</span>
              <span class="font-medium text-white">{{ tokenTooltipData.output_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_creation_tokens > 0">
              <!-- 有 5m/1h 明细时，展开显示 -->
              <template v-if="tokenTooltipData.cache_creation_5m_tokens > 0 || tokenTooltipData.cache_creation_1h_tokens > 0">
                <div v-if="tokenTooltipData.cache_creation_5m_tokens > 0" class="flex items-center justify-between gap-4">
                  <span class="text-gray-400 flex items-center gap-1.5">
                    {{ t('admin.usage.cacheCreation5mTokens') }}
                    <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-amber-500/20 text-amber-400 ring-1 ring-inset ring-amber-500/30">5m</span>
                  </span>
                  <span class="font-medium text-white">{{ tokenTooltipData.cache_creation_5m_tokens.toLocaleString() }}</span>
                </div>
                <div v-if="tokenTooltipData.cache_creation_1h_tokens > 0" class="flex items-center justify-between gap-4">
                  <span class="text-gray-400 flex items-center gap-1.5">
                    {{ t('admin.usage.cacheCreation1hTokens') }}
                    <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-orange-500/20 text-orange-400 ring-1 ring-inset ring-orange-500/30">1h</span>
                  </span>
                  <span class="font-medium text-white">{{ tokenTooltipData.cache_creation_1h_tokens.toLocaleString() }}</span>
                </div>
              </template>
              <!-- 无明细时，只显示聚合值 -->
              <div v-else class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('admin.usage.cacheCreationTokens') }}</span>
                <span class="font-medium text-white">{{ tokenTooltipData.cache_creation_tokens.toLocaleString() }}</span>
              </div>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_ttl_overridden" class="flex items-center justify-between gap-4">
              <span class="text-gray-400 flex items-center gap-1.5">
                {{ t('usage.cacheTtlOverriddenLabel') }}
                <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-rose-500/20 text-rose-400 ring-1 ring-inset ring-rose-500/30">R-{{ tokenTooltipData.cache_creation_1h_tokens > 0 ? '5m' : '1H' }}</span>
              </span>
              <span class="font-medium text-rose-400">{{ tokenTooltipData.cache_creation_1h_tokens > 0 ? t('usage.cacheTtlOverridden1h') : t('usage.cacheTtlOverridden5m') }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_read_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheReadTokens') }}</span>
              <span class="font-medium text-white">{{ tokenTooltipData.cache_read_tokens.toLocaleString() }}</span>
            </div>
          </div>
          <!-- Total -->
          <div class="flex items-center justify-between gap-6 border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('usage.totalTokens') }}</span>
            <span class="font-semibold text-blue-400">{{ ((tokenTooltipData?.input_tokens || 0) + (tokenTooltipData?.output_tokens || 0) + (tokenTooltipData?.cache_creation_tokens || 0) + (tokenTooltipData?.cache_read_tokens || 0)).toLocaleString() }}</span>
          </div>
        </div>
        <!-- Tooltip Arrow (left side) -->
        <div
          class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-gray-900 border-t-transparent dark:border-r-gray-800"
        ></div>
      </div>
    </div>
  </Teleport>

  <!-- Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: tooltipPosition.x + 'px',
        top: tooltipPosition.y + 'px'
      }"
    >
      <div
        class="whitespace-nowrap rounded-lg border border-gray-700 bg-gray-900 px-3 py-2.5 text-xs text-white shadow-xl dark:border-gray-600 dark:bg-gray-800"
      >
        <div class="space-y-1.5">
          <!-- Cost Breakdown -->
          <div class="mb-2 border-b border-gray-700 pb-1.5">
            <div class="text-xs font-semibold text-gray-300 mb-1">{{ t('usage.costDetails') }}</div>
            <div v-if="tooltipData && tooltipData.input_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.inputCost') }}</span>
              <span class="font-medium text-white">${{ tooltipData.input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.output_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.outputCost') }}</span>
              <span class="font-medium text-white">${{ tooltipData.output_cost.toFixed(6) }}</span>
            </div>
            <!-- Per-image billing: show image metadata and unit price -->
            <template v-if="tooltipData && isImageUsage(tooltipData)">
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageCount') }}</span>
                <span class="font-medium text-white">{{ tooltipData.image_count }}{{ t('usage.imageUnit') }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageBillingSize') }}</span>
                <span class="font-medium text-white">{{ formatImageBillingSize(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageSizeSource') }}</span>
                <span class="font-medium text-white">{{ formatImageSizeSource(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageInputSize') }}</span>
                <span class="font-medium text-white">{{ formatImageInputSize(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageOutputSize') }}</span>
                <span class="font-medium text-white">{{ formatImageOutputSize(tooltipData, t) }}</span>
              </div>
              <div v-if="formatImageSizeBreakdown(tooltipData)" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageSizeBreakdown') }}</span>
                <span class="font-medium text-white">{{ formatImageSizeBreakdown(tooltipData) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageUnitPrice') }}</span>
                <span class="font-medium text-sky-300">${{ imageUnitPrice(tooltipData).toFixed(6) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageTotalPrice') }}</span>
                <span class="font-medium text-white">${{ tooltipData.total_cost?.toFixed(6) || '0.000000' }}</span>
              </div>
            </template>
            <!-- Token billing: show unit prices per 1M tokens -->
            <template v-else-if="!getDisplayBillingMode(tooltipData) || getDisplayBillingMode(tooltipData) === BILLING_MODE_TOKEN">
              <div v-if="tooltipData && tooltipData.input_tokens > 0" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.inputTokenPrice') }}</span>
                <span class="font-medium text-sky-300">{{ formatTokenPricePerMillion(tooltipData.input_cost, tooltipData.input_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && tooltipData.output_tokens > 0" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.outputTokenPrice') }}</span>
                <span class="font-medium text-violet-300">{{ formatTokenPricePerMillion(tooltipData.output_cost, tooltipData.output_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
            </template>
            <div v-else class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.unitPrice') }}</span>
              <span class="font-medium text-sky-300">${{ tooltipData?.total_cost?.toFixed(6) || '0.000000' }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.cache_creation_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="font-medium text-white">${{ tooltipData.cache_creation_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.cache_read_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="font-medium text-white">${{ tooltipData.cache_read_cost.toFixed(6) }}</span>
            </div>
          </div>
          <!-- Rate and Summary -->
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.serviceTier') }}</span>
            <span class="font-semibold text-cyan-300">{{ getUsageServiceTierLabel(tooltipData?.service_tier, t) }}</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.rate') }}</span>
            <span class="font-semibold text-blue-400"
              >{{ formatMultiplier(tooltipData?.rate_multiplier || 1) }}x</span
            >
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.original') }}</span>
            <span class="font-medium text-white">${{ tooltipData?.total_cost.toFixed(6) }}</span>
          </div>
          <div class="flex items-center justify-between gap-6 border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('usage.billed') }}</span>
            <span class="font-semibold text-green-400"
              >${{ tooltipData?.actual_cost.toFixed(6) }}</span
            >
          </div>
        </div>
        <!-- Tooltip Arrow (left side) -->
        <div
          class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-gray-900 border-t-transparent dark:border-r-gray-800"
        ></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { usageAPI, keysAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { UsageLog, ApiKey, UsageQueryParams, UsageStatsResponse, UsageRequestType } from '@/types'
import { formatDateTime, formatReasoningEffort } from '@/utils/format'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { formatMultiplier } from '@/utils/formatters'
import { formatTokenPricePerMillion } from '@/utils/usagePricing'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_TOKEN,
  getBillingModeLabel,
} from '@/utils/billingMode'
import {
  formatImageBillingSize,
  formatImageInputSize,
  formatImageOutputSize,
  formatImageSizeBreakdown,
  formatImageSizeSource,
} from '@/utils/imageUsage'

const { t } = useI18n()
const appStore = useAppStore()

let abortController: AbortController | null = null
let autoRefreshTimer: ReturnType<typeof setInterval> | null = null

// Tooltip state
const tooltipVisible = ref(false)
const tooltipPosition = ref({ x: 0, y: 0 })
const tooltipData = ref<UsageLog | null>(null)

// Token tooltip state
const tokenTooltipVisible = ref(false)
const tokenTooltipPosition = ref({ x: 0, y: 0 })
const tokenTooltipData = ref<UsageLog | null>(null)

// Usage stats from API
const usageStats = ref<UsageStatsResponse | null>(null)

const usageLogs = ref<UsageLog[]>([])
const apiKeys = ref<ApiKey[]>([])
const loading = ref(false)
const exporting = ref(false)
const filtersOpen = ref(true)
const timeFiltersOpen = ref(true)
const identityFiltersOpen = ref(true)
const requestFiltersOpen = ref(false)
const statusFiltersOpen = ref(false)
const columnDropdownOpen = ref(false)
const autoRefreshEnabled = ref(false)
const isDesktopTable = ref(typeof window === 'undefined' ? true : window.innerWidth >= 768)

type DatePresetKey = 'today' | 'this_week' | 'last_7_days' | 'last_30_days'
type UsageLogColumnKey =
  | 'time'
  | 'api_key'
  | 'request_id'
  | 'endpoint'
  | 'model'
  | 'tokens'
  | 'cache'
  | 'cost'
  | 'performance'
  | 'status'

const activePreset = ref<DatePresetKey | null>('last_7_days')
const hiddenLogColumns = ref<UsageLogColumnKey[]>(['endpoint', 'cache'])

const logColumnOptions = computed<Array<{ key: UsageLogColumnKey; label: string }>>(() => [
  { key: 'time', label: t('usage.time') },
  { key: 'api_key', label: t('usage.apiKeyFilter') },
  { key: 'request_id', label: t('usage.requestId') },
  { key: 'endpoint', label: t('usage.endpoint') },
  { key: 'model', label: t('usage.billingModel') },
  { key: 'tokens', label: t('usage.tokens') },
  { key: 'cache', label: t('usage.cache') },
  { key: 'cost', label: t('usage.cost') },
  { key: 'performance', label: t('usage.performance') },
  { key: 'status', label: t('usage.status') }
])

const visibleLogColumns = computed(() => (
  logColumnOptions.value.filter((column) => isLogColumnVisible(column.key))
))

const visibleLogColumnCount = computed(() => visibleLogColumns.value.length)

const quickDatePresets = computed<Array<{ key: DatePresetKey; label: string; icon: 'calendar' | 'clock' }>>(() => [
  { key: 'today', label: t('usage.quickToday'), icon: 'calendar' },
  { key: 'this_week', label: t('usage.quickThisWeek'), icon: 'calendar' },
  { key: 'last_7_days', label: t('usage.quickLast7Days'), icon: 'clock' },
  { key: 'last_30_days', label: t('usage.quickLast30Days'), icon: 'clock' }
])

const requestTypeOptions = computed(() => [
  { value: null, label: t('usage.allTypes') },
  { value: 'sync', label: t('usage.sync') },
  { value: 'stream', label: t('usage.stream') },
  { value: 'ws_v2', label: t('usage.ws') }
])

const apiKeyOptions = computed(() => {
  return [
    { value: null, label: t('usage.allApiKeys') },
    ...apiKeys.value.map((key) => ({
      value: key.id,
      label: key.name
    }))
  ]
})

const apiKeyFilter = computed({
  get: () => filters.value.api_key_id ?? null,
  set: (value: string | number | boolean | null) => {
    filters.value.api_key_id = typeof value === 'number' ? value : undefined
  }
})

const modelFilter = computed({
  get: () => filters.value.model || '',
  set: (value: string) => {
    filters.value.model = value.trim() || undefined
  }
})

const requestTypeFilter = computed({
  get: () => filters.value.request_type ?? null,
  set: (value: string | number | boolean | null) => {
    filters.value.request_type = typeof value === 'string' && value
      ? value as UsageRequestType
      : undefined
  }
})

const activeFilterCount = computed(() => {
  let count = 0
  if (filters.value.api_key_id) count += 1
  if (filters.value.model) count += 1
  if (filters.value.request_type) count += 1
  if (filters.value.start_date || filters.value.end_date) count += 1
  return count
})

const activeSessionRows = computed(() => {
  const now = Date.now()
  const activeWindowMs = 5 * 60 * 1000

  return usageLogs.value
    .filter((row) => {
      const createdAt = new Date(row.created_at).getTime()
      if (Number.isNaN(createdAt)) return false
      return now - createdAt <= activeWindowMs || row.duration_ms === 0
    })
    .slice(0, 5)
})

const usageStatsSummary = computed(() => {
  const requests = usageStats.value?.total_requests || 0
  const tokens = usageStats.value?.total_tokens || 0
  const cost = usageStats.value?.total_actual_cost || 0
  return `${requests.toLocaleString()} ${t('usage.requestsShort')} · ${formatTokens(tokens)} ${t('usage.tokens')} · $${cost.toFixed(4)}`
})

// Helper function to format date in local timezone
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

// Initialize date range immediately
const now = new Date()
const weekAgo = new Date(now)
weekAgo.setDate(weekAgo.getDate() - 6)

// Date range state
const startDate = ref(formatLocalDate(weekAgo))
const endDate = ref(formatLocalDate(now))

const filters = ref<UsageQueryParams>({
  api_key_id: undefined,
  start_date: undefined,
  end_date: undefined
})

// Initialize filters with date range
filters.value.start_date = startDate.value
filters.value.end_date = endDate.value

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})
const sortState = reactive({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc'
})

const formatDuration = (ms: number): string => {
  if (ms < 1000) return `${ms.toFixed(0)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const imageUnitPrice = (row: UsageLog | null): number => {
  if (!row || row.image_count <= 0) return 0
  const total = row.total_cost ?? 0
  const price = total / row.image_count
  return Number.isFinite(price) ? price : 0
}

const isImageUsage = (row: Pick<UsageLog, 'image_count'> | null | undefined): boolean => {
  return (row?.image_count ?? 0) > 0
}

const getDisplayBillingMode = (row: Pick<UsageLog, 'billing_mode' | 'image_count'> | null | undefined): string | null | undefined => {
  if (isImageUsage(row)) {
    return BILLING_MODE_IMAGE
  }
  return row?.billing_mode
}

const getRequestTypeLabel = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const getRequestTypeBadgeClass = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'ws_v2') return 'bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-200'
  if (requestType === 'stream') return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
  if (requestType === 'sync') return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
}


const getRequestTypeExportText = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'ws_v2') return 'WS'
  if (requestType === 'stream') return 'Stream'
  if (requestType === 'sync') return 'Sync'
  return 'Unknown'
}

const formatUsageEndpoints = (log: UsageLog): string => {
  const inbound = log.inbound_endpoint?.trim()
  return inbound || '-'
}

const getTotalTokens = (row: UsageLog): number => {
  return row.input_tokens + row.output_tokens + row.cache_creation_tokens + row.cache_read_tokens
}

const getCacheTokens = (row: UsageLog): number => {
  return row.cache_creation_tokens + row.cache_read_tokens
}

const shortRequestId = (value: string): string => {
  if (!value) return '-'
  return value.length > 12 ? `${value.slice(0, 8)}...${value.slice(-4)}` : value
}

const formatRelativeTime = (value: string): string => {
  const timestamp = new Date(value).getTime()
  if (Number.isNaN(timestamp)) return '-'

  const diffSeconds = Math.max(0, Math.floor((Date.now() - timestamp) / 1000))
  if (diffSeconds < 30) return t('usage.justNow')
  if (diffSeconds < 60) return t('usage.secondsAgo', { count: diffSeconds })

  const diffMinutes = Math.floor(diffSeconds / 60)
  if (diffMinutes < 60) return t('usage.minutesAgo', { count: diffMinutes })

  const diffHours = Math.floor(diffMinutes / 60)
  if (diffHours < 24) return t('usage.hoursAgo', { count: diffHours })

  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) return t('usage.daysAgo', { count: diffDays })

  return formatDateTime(value)
}

const getUsageStatusLabel = (row: UsageLog): string => {
  if (row.duration_ms === 0) return t('usage.statusPending')
  if (row.actual_cost >= 0) return t('usage.statusSuccess')
  return t('usage.statusFailed')
}

const getUsageStatusClass = (row: UsageLog): string => {
  if (row.duration_ms === 0) {
    return 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-200'
  }
  if (row.actual_cost >= 0) {
    return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-200'
  }
  return 'bg-rose-100 text-rose-800 dark:bg-rose-500/15 dark:text-rose-200'
}

const isLogColumnVisible = (key: UsageLogColumnKey): boolean => {
  return !hiddenLogColumns.value.includes(key)
}

const toggleLogColumn = (key: UsageLogColumnKey) => {
  if (visibleLogColumnCount.value === 1 && isLogColumnVisible(key)) return

  hiddenLogColumns.value = isLogColumnVisible(key)
    ? [...hiddenLogColumns.value, key]
    : hiddenLogColumns.value.filter((column) => column !== key)
}

const startOfWeek = (date: Date): Date => {
  const start = new Date(date)
  const day = start.getDay() || 7
  start.setDate(start.getDate() - day + 1)
  return start
}

const syncDateRange = () => {
  filters.value.start_date = startDate.value || undefined
  filters.value.end_date = endDate.value || undefined
  activePreset.value = null
}

const applyDatePreset = (preset: DatePresetKey) => {
  const today = new Date()
  let start = new Date(today)

  if (preset === 'today') {
    start = new Date(today)
  } else if (preset === 'this_week') {
    start = startOfWeek(today)
  } else if (preset === 'last_7_days') {
    start.setDate(today.getDate() - 6)
  } else {
    start.setDate(today.getDate() - 29)
  }

  activePreset.value = preset
  startDate.value = formatLocalDate(start)
  endDate.value = formatLocalDate(today)
  filters.value.start_date = startDate.value
  filters.value.end_date = endDate.value
  applyFilters()
}

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

type UsageTableQueryParams = UsageQueryParams & {
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

const buildUsageQueryParams = (page: number, pageSize: number): UsageTableQueryParams => ({
  page,
  page_size: pageSize,
  ...filters.value,
  sort_by: sortState.sort_by,
  sort_order: sortState.sort_order
})

const loadUsageLogs = async () => {
  if (abortController) {
    abortController.abort()
  }
  const currentAbortController = new AbortController()
  abortController = currentAbortController
  const { signal } = currentAbortController
  loading.value = true
  try {
    const response = await usageAPI.query(
      buildUsageQueryParams(pagination.page, pagination.page_size),
      { signal }
    )
    if (signal.aborted) {
      return
    }
    usageLogs.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error) {
    if (signal.aborted) {
      return
    }
    const abortError = error as { name?: string; code?: string }
    if (abortError?.name === 'AbortError' || abortError?.code === 'ERR_CANCELED') {
      return
    }
    appStore.showError(t('usage.failedToLoad'))
  } finally {
    if (abortController === currentAbortController) {
      loading.value = false
    }
  }
}

const loadApiKeys = async () => {
  try {
    const response = await keysAPI.list(1, 100)
    apiKeys.value = response.items
  } catch (error) {
    console.error('Failed to load API keys:', error)
  }
}

const loadUsageStats = async () => {
  try {
    const apiKeyId = filters.value.api_key_id ? Number(filters.value.api_key_id) : undefined
    const stats = await usageAPI.getStatsByDateRange(
      filters.value.start_date || startDate.value,
      filters.value.end_date || endDate.value,
      apiKeyId
    )
    usageStats.value = stats
  } catch (error) {
    console.error('Failed to load usage stats:', error)
  }
}

const applyFilters = () => {
  pagination.page = 1
  loadUsageLogs()
  loadUsageStats()
}

const resetFilters = () => {
  filters.value = {
    api_key_id: undefined,
    model: undefined,
    request_type: undefined,
    start_date: undefined,
    end_date: undefined
  }
  const now = new Date()
  const weekAgo = new Date(now)
  weekAgo.setDate(weekAgo.getDate() - 6)
  activePreset.value = 'last_7_days'
  startDate.value = formatLocalDate(weekAgo)
  endDate.value = formatLocalDate(now)
  filters.value.start_date = startDate.value
  filters.value.end_date = endDate.value
  pagination.page = 1
  loadUsageLogs()
  loadUsageStats()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadUsageLogs()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadUsageLogs()
}

const refreshData = () => {
  loadUsageLogs()
  loadUsageStats()
}

const toggleAutoRefresh = () => {
  autoRefreshEnabled.value = !autoRefreshEnabled.value

  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }

  if (autoRefreshEnabled.value) {
    autoRefreshTimer = setInterval(() => {
      refreshData()
    }, 5000)
  }
}

const handleViewportResize = () => {
  isDesktopTable.value = window.innerWidth >= 768
}

/**
 * Escape CSV value to prevent injection and handle special characters
 */
const escapeCSVValue = (value: unknown): string => {
  if (value == null) return ''

  const str = String(value)
  const escaped = str.replace(/"/g, '""')

  // Prevent formula injection by prefixing dangerous characters with single quote
  if (/^[=+\-@\t\r]/.test(str)) {
    return `"\'${escaped}"`
  }

  // Escape values containing comma, quote, or newline
  if (/[,"\n\r]/.test(str)) {
    return `"${escaped}"`
  }

  return str
}

const exportToCSV = async () => {
  if (pagination.total === 0) {
    appStore.showWarning(t('usage.noDataToExport'))
    return
  }

  exporting.value = true
  appStore.showInfo(t('usage.preparingExport'))

  try {
    const allLogs: UsageLog[] = []
    const pageSize = 100 // Use a larger page size for export to reduce requests
    const totalRequests = Math.ceil(pagination.total / pageSize)

    for (let page = 1; page <= totalRequests; page++) {
      const response = await usageAPI.query(buildUsageQueryParams(page, pageSize))
      allLogs.push(...response.items)
    }

    if (allLogs.length === 0) {
      appStore.showWarning(t('usage.noDataToExport'))
      return
    }

    const headers = [
      'Time',
      'API Key Name',
      'Model',
      'Reasoning Effort',
      'Inbound Endpoint',
      'Type',
      'Billing Mode',
      'Input Tokens',
      'Output Tokens',
      'Cache Read Tokens',
      'Cache Creation Tokens',
      'Rate Multiplier',
      'Billed Cost',
      'Original Cost',
      'First Token (ms)',
      'Duration (ms)'
    ]
    const rows = allLogs.map((log) =>
      [
        log.created_at,
        log.api_key?.name || '',
        log.model,
        formatReasoningEffort(log.reasoning_effort),
        log.inbound_endpoint || '',
        getRequestTypeExportText(log),
        getBillingModeLabel(getDisplayBillingMode(log), t),
        log.input_tokens,
        log.output_tokens,
        log.cache_read_tokens,
        log.cache_creation_tokens,
        log.rate_multiplier,
        log.actual_cost.toFixed(8),
        log.total_cost.toFixed(8),
        log.first_token_ms ?? '',
        log.duration_ms
      ].map(escapeCSVValue)
    )

    const csvContent = [
      headers.map(escapeCSVValue).join(','),
      ...rows.map((row) => row.join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `usage_${filters.value.start_date}_to_${filters.value.end_date}.csv`
    link.click()
    window.URL.revokeObjectURL(url)

    appStore.showSuccess(t('usage.exportSuccess'))
  } catch (error) {
    appStore.showError(t('usage.exportFailed'))
    console.error('CSV Export failed:', error)
  } finally {
    exporting.value = false
  }
}

// Tooltip functions
const showTooltip = (event: MouseEvent, row: UsageLog) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()

  tooltipData.value = row
  // Position to the right of the icon, vertically centered
  tooltipPosition.value.x = rect.right + 8
  tooltipPosition.value.y = rect.top + rect.height / 2
  tooltipVisible.value = true
}

const hideTooltip = () => {
  tooltipVisible.value = false
  tooltipData.value = null
}

// Token tooltip functions
const showTokenTooltip = (event: MouseEvent, row: UsageLog) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()

  tokenTooltipData.value = row
  tokenTooltipPosition.value.x = rect.right + 8
  tokenTooltipPosition.value.y = rect.top + rect.height / 2
  tokenTooltipVisible.value = true
}

const hideTokenTooltip = () => {
  tokenTooltipVisible.value = false
  tokenTooltipData.value = null
}

onMounted(() => {
  loadApiKeys()
  loadUsageLogs()
  loadUsageStats()
  window.addEventListener('resize', handleViewportResize)
})

onUnmounted(() => {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
  window.removeEventListener('resize', handleViewportResize)
})
</script>
