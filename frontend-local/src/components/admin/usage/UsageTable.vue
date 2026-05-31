<template>
  <div :class="shellClass">
    <div class="divide-y divide-gray-100 md:hidden dark:divide-white/[0.06]">
      <div v-if="loading && data.length === 0" class="space-y-3 p-4">
        <div v-for="idx in 4" :key="idx" class="h-24 animate-pulse rounded-xl bg-gray-100 dark:bg-white/[0.06]" />
      </div>
      <div v-else-if="data.length === 0" class="flex min-h-[220px] flex-col items-center justify-center gap-3 p-6 text-center">
        <Icon name="inbox" size="xl" class="text-gray-300 dark:text-dark-500" />
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('usage.noRecords') }}</p>
      </div>
      <article
        v-for="row in data"
        v-else
        :key="rowKey(row)"
        class="space-y-3 p-4"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div
              class="truncate text-sm font-semibold text-gray-950 dark:text-white"
              @mouseenter="showLongTextTooltip($event, row.model)"
              @mouseleave="hideLongTextTooltip"
            >
              {{ row.model || '-' }}
            </div>
            <div
              class="mt-1 text-xs text-gray-500 dark:text-dark-400"
              @mouseenter="showLongTextTooltip($event, formatDateTime(row.created_at), { force: true })"
              @mouseleave="hideLongTextTooltip"
            >
              {{ formatRelativeTime(row.created_at) }}
              <template v-if="hasColumn('user')"> · {{ getUserDisplayName(row) }}</template>
              <template v-else> · {{ row.api_key?.name || t('usage.unknownKey') }}</template>
            </div>
            <div
              v-if="hasColumn('account')"
              class="mt-1 truncate text-xs text-gray-500 dark:text-dark-400"
              @mouseenter="showLongTextTooltip($event, getAccountDisplayName(row))"
              @mouseleave="hideLongTextTooltip"
            >
              {{ getAccountDisplayName(row) }}
            </div>
            <div
              v-if="hasColumn('user_agent')"
              class="mt-1 line-clamp-2 break-words text-xs leading-4 text-gray-400 dark:text-dark-500"
              @mouseenter="showLongTextTooltip($event, row.user_agent)"
              @mouseleave="hideLongTextTooltip"
            >
              {{ row.user_agent || '-' }}
            </div>
          </div>
          <span
            v-if="hasColumn('status')"
            class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold"
            :class="getUsageStatusClass(row)"
            @mouseenter="showLongTextTooltip($event, getStatusTooltipContent(row), { force: true })"
            @mouseleave="hideLongTextTooltip"
          >
            {{ getUsageStatusLabel(row) }}
          </span>
        </div>
        <div class="grid grid-cols-2 gap-3 text-sm">
          <div>
            <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.tokens') }}</div>
            <div class="font-semibold text-gray-950 dark:text-white">{{ hasUsageFacts(row) ? formatTokens(getTotalTokens(row)) : '-' }}</div>
          </div>
          <div>
            <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('usage.cost') }}</div>
            <div class="font-semibold text-emerald-600 dark:text-emerald-300">{{ hasBillableCost(row) ? `$${num(row.actual_cost).toFixed(6)}` : '-' }}</div>
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

    <div class="hidden overflow-x-auto md:block">
      <table class="w-full min-w-[980px] divide-y divide-gray-200/70 dark:divide-white/[0.06]">
        <thead class="bg-gray-50/80 dark:bg-white/[0.03]">
          <tr>
            <th
              v-for="column in visibleColumns"
              :key="column.key"
              class="px-4 py-3 text-xs font-medium text-gray-500 dark:text-dark-400"
              :class="getHeaderClass(column.key)"
            >
              <button
                v-if="column.sortable"
                type="button"
                class="inline-flex items-center gap-1 transition-colors hover:text-gray-950 dark:hover:text-white"
                :class="isRightAligned(column.key) ? 'justify-end' : 'justify-start'"
                @click="toggleSort(column.key)"
              >
                {{ column.label }}
                <Icon
                  name="chevronDown"
                  size="xs"
                  class="transition-transform"
                  :class="{ 'rotate-180': currentSortKey === column.key && currentSortOrder === 'asc' }"
                />
              </button>
              <span v-else>{{ column.label }}</span>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 bg-white/40 dark:divide-white/[0.05] dark:bg-transparent">
          <template v-if="loading && data.length === 0">
            <tr v-for="idx in 5" :key="idx">
              <td v-for="column in visibleColumns" :key="column.key" class="px-4 py-4">
                <div class="h-4 w-24 animate-pulse rounded bg-gray-200 dark:bg-white/[0.08]" />
              </td>
            </tr>
          </template>
          <tr v-else-if="data.length === 0">
            <td :colspan="visibleColumns.length" class="px-4 py-16 text-center">
              <div class="flex flex-col items-center gap-3">
                <Icon name="inbox" size="xl" class="text-gray-300 dark:text-dark-500" />
                <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('usage.noRecords') }}</p>
              </div>
            </td>
          </tr>
          <tr
            v-for="row in data"
            v-else
            :key="rowKey(row)"
            class="transition-colors hover:bg-orange-50/40 dark:hover:bg-white/[0.035]"
          >
            <td v-for="column in visibleColumns" :key="column.key" class="px-4 py-4" :class="getCellClass(column.key)">
              <template v-if="column.key === 'user'">
                <span
                  v-if="row.user"
                  class="block max-w-[180px] truncate text-sm font-semibold text-gray-950 dark:text-white"
                  @mouseenter="showLongTextTooltip($event, getUserTooltip(row), { force: hasExtraUserDetails(row) })"
                  @mouseleave="hideLongTextTooltip"
                >
                  {{ getUserDisplayName(row) }}
                </span>
                <span v-else class="text-sm font-semibold text-gray-950 dark:text-white">#{{ row.user_id || '-' }}</span>
              </template>

              <template v-else-if="column.key === 'time' || column.key === 'created_at'">
                <span
                  class="whitespace-nowrap text-sm text-gray-700 dark:text-dark-200"
                  @mouseenter="showLongTextTooltip($event, formatDateTime(row.created_at), { force: true })"
                  @mouseleave="hideLongTextTooltip"
                >
                  {{ formatRelativeTime(row.created_at) }}
                </span>
              </template>

              <template v-else-if="column.key === 'api_key'">
                <div class="flex min-w-0 items-center gap-2">
                  <Icon name="key" size="xs" class="flex-shrink-0 text-gray-400" />
                  <span class="max-w-[150px] truncate text-sm font-medium text-gray-950 dark:text-white">
                    {{ row.api_key?.name || (row.api_key_id ? `#${row.api_key_id}` : t('usage.unknownKey')) }}
                  </span>
                </div>
              </template>

              <template v-else-if="column.key === 'account'">
                <div class="flex min-w-0 items-center gap-2">
                  <Icon name="server" size="xs" class="flex-shrink-0 text-gray-400" />
                  <span
                    class="max-w-[150px] truncate text-sm font-medium text-gray-950 dark:text-white"
                    @mouseenter="showLongTextTooltip($event, getAccountDisplayName(row))"
                    @mouseleave="hideLongTextTooltip"
                  >
                    {{ getAccountDisplayName(row) }}
                  </span>
                </div>
              </template>

              <template v-else-if="column.key === 'session'">
                <button
                  v-if="row.session_id"
                  type="button"
                  class="block max-w-[190px] text-left"
                  @click="emit('session-click', row.session_id)"
                  @mouseenter="showLongTextTooltip($event, formatSessionTitle(row), { force: hasExtraSessionDetails(row) })"
                  @mouseleave="hideLongTextTooltip"
                >
                  <span class="block truncate font-mono text-xs font-semibold text-gray-800 dark:text-dark-100">{{ row.session_id }}</span>
                  <span class="block truncate text-[10px] text-gray-500 dark:text-dark-400">{{ row.session_source || '-' }}</span>
                </button>
                <span v-else class="text-sm text-gray-500 dark:text-dark-400">-</span>
              </template>

              <template v-else-if="column.key === 'request_id'">
                <span
                  class="font-mono text-xs text-gray-700 dark:text-dark-200"
                  @mouseenter="showLongTextTooltip($event, row.request_id, { force: true })"
                  @mouseleave="hideLongTextTooltip"
                >
                  {{ shortRequestId(row.request_id) }}
                </span>
              </template>

              <template v-else-if="column.key === 'endpoint'">
                <span
                  class="block max-w-[220px] truncate text-sm text-gray-600 dark:text-dark-300"
                  @mouseenter="showLongTextTooltip($event, formatUsageEndpoints(row))"
                  @mouseleave="hideLongTextTooltip"
                >
                  {{ formatUsageEndpoints(row) }}
                </span>
              </template>

              <template v-else-if="column.key === 'model'">
                <div class="flex min-w-0 items-center gap-2">
                  <Icon name="cpu" size="xs" class="text-gray-500 dark:text-dark-400" />
                  <span
                    class="max-w-[180px] truncate text-sm font-semibold text-gray-950 dark:text-white"
                    @mouseenter="showLongTextTooltip($event, row.model)"
                    @mouseleave="hideLongTextTooltip"
                  >{{ row.model || '-' }}</span>
                </div>
                <div v-if="row.model_mapping_chain && row.model_mapping_chain.includes('→')" class="mt-0.5 space-y-0.5 text-[10px]">
                  <div
                    v-for="(step, i) in row.model_mapping_chain.split('→')"
                    :key="i"
                    class="max-w-[180px] truncate"
                    :class="i === 0 ? 'text-gray-500 dark:text-dark-400' : 'text-gray-400 dark:text-dark-500'"
                    :style="i > 0 ? `padding-left: ${i * 0.5}rem` : ''"
                    @mouseenter="showLongTextTooltip($event, step.trim())"
                    @mouseleave="hideLongTextTooltip"
                  >
                    <span v-if="i > 0" class="mr-0.5">↳</span>{{ step.trim() }}
                  </div>
                </div>
                <div
                  v-else-if="row.upstream_model && row.upstream_model !== row.model"
                  class="mt-0.5 max-w-[180px] truncate text-[10px] text-gray-500 dark:text-dark-400"
                  @mouseenter="showLongTextTooltip($event, row.upstream_model)"
                  @mouseleave="hideLongTextTooltip"
                >
                  ↳ {{ row.upstream_model }}
                </div>
                <div v-if="row.reasoning_effort" class="mt-0.5 text-[10px] text-gray-500 dark:text-dark-400">
                  {{ formatReasoningEffort(row.reasoning_effort) }}
                </div>
              </template>

              <template v-else-if="column.key === 'tokens'">
                <span v-if="!hasUsageFacts(row)" class="text-sm text-gray-500 dark:text-dark-400">-</span>
                <div v-else-if="isImageUsage(row)" class="inline-flex items-center justify-end gap-2">
                  <Icon name="sparkles" size="xs" class="text-pink-500 dark:text-pink-300" />
                  <div>
                    <div class="text-sm font-semibold text-gray-950 dark:text-white">
                      {{ num(row.image_count) }}{{ t('usage.imageUnit') }}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-dark-400">
                      {{ getBillingModeLabel(getDisplayBillingMode(row), t) }} · {{ formatImageBillingSize(row, t) }}
                    </div>
                  </div>
                </div>
                <div v-else class="inline-flex items-center justify-end gap-1.5">
                  <div>
                    <div class="font-mono text-sm font-semibold text-gray-950 dark:text-white">{{ formatTokens(getTotalTokens(row)) }}</div>
                    <div class="text-xs text-gray-500 dark:text-dark-400">{{ formatTokens(num(row.output_tokens)) }}</div>
                  </div>
                  <button
                    type="button"
                    :aria-label="t('usage.tokenDetails')"
                    class="rounded-full text-gray-400 transition-colors hover:text-blue-500"
                    @mouseenter="showTokenTooltip($event, row)"
                    @focus="showTokenTooltip($event, row)"
                    @mouseleave="hideTokenTooltip"
                    @blur="hideTokenTooltip"
                  >
                    <Icon name="infoCircle" size="xs" />
                  </button>
                </div>
              </template>

              <template v-else-if="column.key === 'cache'">
                <div v-if="hasUsageFacts(row)" class="font-mono text-sm text-gray-950 dark:text-white">
                  {{ getCacheTokens(row) > 0 ? formatTokens(getCacheTokens(row)) : '-' }}
                </div>
                <span v-else class="text-sm text-gray-500 dark:text-dark-400">-</span>
                <div v-if="row.cache_ttl_overridden" class="text-[10px] text-rose-500">TTL</div>
              </template>

              <template v-else-if="column.key === 'billing_mode'">
                <span class="inline-flex rounded-full px-2.5 py-1 text-xs font-semibold" :class="getBillingModeBadgeClass(getDisplayBillingMode(row))">
                  {{ getBillingModeLabel(getDisplayBillingMode(row), t) }}
                </span>
              </template>

              <template v-else-if="column.key === 'group'">
                <span v-if="row.group?.name" class="inline-flex rounded-full bg-indigo-50 px-2.5 py-1 text-xs font-semibold text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-200">
                  {{ row.group.name }}
                </span>
                <span v-else class="text-sm text-gray-500 dark:text-dark-400">-</span>
              </template>

              <template v-else-if="column.key === 'user_agent'">
                <span
                  class="line-clamp-2 max-w-[180px] break-words text-sm leading-5 text-gray-600 dark:text-dark-300"
                  @mouseenter="showLongTextTooltip($event, row.user_agent)"
                  @mouseleave="hideLongTextTooltip"
                >
                  {{ row.user_agent || '-' }}
                </span>
              </template>

              <template v-else-if="column.key === 'ip_address'">
                <span class="font-mono text-xs text-gray-700 dark:text-dark-200">
                  {{ row.ip_address || '-' }}
                </span>
              </template>

              <template v-else-if="column.key === 'platform'">
                <span class="text-sm text-gray-700 dark:text-dark-200">
                  {{ row.platform || '-' }}
                </span>
              </template>

              <template v-else-if="column.key === 'cost'">
                <span v-if="!hasBillableCost(row)" class="text-sm text-gray-500 dark:text-dark-400">-</span>
                <div v-else class="inline-flex items-center justify-end gap-1.5">
                  <span class="font-mono text-sm font-semibold text-gray-950 dark:text-white">${{ num(row.actual_cost).toFixed(6) }}</span>
                  <button
                    type="button"
                    :aria-label="t('usage.costDetails')"
                    class="rounded-full text-gray-400 transition-colors hover:text-blue-500"
                    @mouseenter="showTooltip($event, row)"
                    @focus="showTooltip($event, row)"
                    @mouseleave="hideTooltip"
                    @blur="hideTooltip"
                  >
                    <Icon name="infoCircle" size="xs" />
                  </button>
                </div>
              </template>

              <template v-else-if="column.key === 'billable'">
                <span
                  class="inline-flex rounded-full px-2.5 py-1 text-xs font-semibold"
                  :class="row.billable ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-200' : 'bg-gray-100 text-gray-700 dark:bg-white/[0.06] dark:text-dark-200'"
                >
                  {{ row.billable ? t('usage.billable') : t('usage.nonBillable') }}
                </span>
              </template>

              <template v-else-if="column.key === 'performance'">
                <div class="font-mono text-sm text-gray-950 dark:text-white">{{ formatDuration(row.duration_ms) }}</div>
                <div class="text-xs text-gray-500 dark:text-dark-400">
                  TTFB {{ row.first_token_ms != null ? formatDuration(row.first_token_ms) : '-' }}
                </div>
              </template>

              <template v-else-if="column.key === 'status'">
                <div class="inline-flex flex-col items-start gap-1">
                  <span
                    class="inline-flex items-center justify-center rounded-full px-2.5 py-1 text-xs font-semibold"
                    :class="getUsageStatusClass(row)"
                    @mouseenter="showLongTextTooltip($event, getStatusTooltipContent(row), { force: true })"
                    @mouseleave="hideLongTextTooltip"
                  >
                    {{ getUsageStatusLabel(row) }}
                  </span>
                  <span
                    v-if="hasUpstreamAttempts(row)"
                    data-testid="upstream-attempts-badge"
                    class="inline-flex max-w-[150px] items-center gap-1 rounded-full bg-sky-50 px-2 py-0.5 text-[10px] font-medium text-sky-700 ring-1 ring-inset ring-sky-200/70 dark:bg-sky-500/10 dark:text-sky-200 dark:ring-sky-400/20"
                    @mouseenter="showLongTextTooltip($event, formatUpstreamAttempts(row), { force: true })"
                    @mouseleave="hideLongTextTooltip"
                  >
                    <Icon name="infoCircle" size="xs" />
                    <span class="truncate">{{ formatUpstreamAttemptCount(row) }}</span>
                  </span>
                </div>
                <div v-if="showRequestTypeBadge" class="mt-1">
                  <span class="inline-flex rounded px-1.5 py-0.5 text-[10px] font-medium" :class="getRequestTypeBadgeClass(row)">
                    {{ getRequestTypeLabel(row) }}
                  </span>
                </div>
              </template>

              <template v-else>
                <span class="text-sm text-gray-500 dark:text-dark-400">-</span>
              </template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <Teleport to="body">
    <div
      v-if="longTextTooltip.show"
      role="tooltip"
      class="usage-table-long-tooltip"
      :class="{
        'usage-table-long-tooltip-top': longTextTooltip.placement === 'top',
        'usage-table-long-tooltip-bottom': longTextTooltip.placement === 'bottom'
      }"
      :style="{ top: `${longTextTooltip.top}px`, left: `${longTextTooltip.left}px` }"
    >
      {{ longTextTooltip.content }}
      <span
        class="usage-table-long-tooltip-arrow"
        :class="{
          'usage-table-long-tooltip-arrow-top': longTextTooltip.placement === 'top',
          'usage-table-long-tooltip-arrow-bottom': longTextTooltip.placement === 'bottom'
        }"
        :style="{ left: `${longTextTooltip.arrowLeft}px` }"
      />
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="tokenTooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{ left: tokenTooltipPosition.x + 'px', top: tokenTooltipPosition.y + 'px' }"
    >
      <div class="whitespace-nowrap rounded-lg border border-gray-700 bg-gray-900 px-3 py-2.5 text-xs text-white shadow-xl dark:border-gray-600 dark:bg-gray-800">
        <div class="space-y-1.5">
          <div>
            <div class="mb-1 text-xs font-semibold text-gray-300">{{ t('usage.tokenDetails') }}</div>
            <div v-if="tokenTooltipData && num(tokenTooltipData.input_tokens) > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.inputTokens') }}</span>
              <span class="font-medium text-white">{{ num(tokenTooltipData.input_tokens).toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && num(tokenTooltipData.output_tokens) > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.outputTokens') }}</span>
              <span class="font-medium text-white">{{ num(tokenTooltipData.output_tokens).toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && num(tokenTooltipData.cache_creation_tokens) > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheCreationTokens') }}</span>
              <span class="font-medium text-white">{{ num(tokenTooltipData.cache_creation_tokens).toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && num(tokenTooltipData.cache_read_tokens) > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheReadTokens') }}</span>
              <span class="font-medium text-white">{{ num(tokenTooltipData.cache_read_tokens).toLocaleString() }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between gap-6 border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('usage.totalTokens') }}</span>
            <span class="font-semibold text-blue-400">{{ getTotalTokens(tokenTooltipData).toLocaleString() }}</span>
          </div>
        </div>
        <div class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-gray-900 border-t-transparent dark:border-r-gray-800"></div>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="tooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{ left: tooltipPosition.x + 'px', top: tooltipPosition.y + 'px' }"
    >
      <div class="whitespace-nowrap rounded-lg border border-gray-700 bg-gray-900 px-3 py-2.5 text-xs text-white shadow-xl dark:border-gray-600 dark:bg-gray-800">
        <div class="space-y-1.5">
          <div class="mb-2 border-b border-gray-700 pb-1.5">
            <div class="mb-1 text-xs font-semibold text-gray-300">{{ t('usage.costDetails') }}</div>
            <template v-if="tooltipData && isImageUsage(tooltipData)">
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageCount') }}</span>
                <span class="font-medium text-white">{{ num(tooltipData.image_count) }}{{ t('usage.imageUnit') }}</span>
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
            </template>
            <template v-else-if="!getDisplayBillingMode(tooltipData) || getDisplayBillingMode(tooltipData) === BILLING_MODE_TOKEN">
              <div v-if="tooltipData && num(tooltipData.input_tokens) > 0" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.inputTokenPrice') }}</span>
                <span class="font-medium text-sky-300">{{ formatTokenPricePerMillion(num(tooltipData.input_cost), num(tooltipData.input_tokens)) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && num(tooltipData.output_tokens) > 0" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.outputTokenPrice') }}</span>
                <span class="font-medium text-violet-300">{{ formatTokenPricePerMillion(num(tooltipData.output_cost), num(tooltipData.output_tokens)) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
            </template>
            <div v-if="tooltipData && num(tooltipData.cache_creation_cost) > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="font-medium text-white">${{ num(tooltipData.cache_creation_cost).toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && num(tooltipData.cache_read_cost) > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="font-medium text-white">${{ num(tooltipData.cache_read_cost).toFixed(6) }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.serviceTier') }}</span>
            <span class="font-semibold text-cyan-300">{{ getUsageServiceTierLabel(tooltipData?.service_tier, t) }}</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.rate') }}</span>
            <span class="font-semibold text-blue-400">{{ formatMultiplier(num(tooltipData?.rate_multiplier, 1)) }}x</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.original') }}</span>
            <span class="font-medium text-white">${{ num(tooltipData?.total_cost).toFixed(6) }}</span>
          </div>
          <div class="flex items-center justify-between gap-6 border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('usage.userBilled') }}</span>
            <span class="font-semibold text-green-400">${{ num(tooltipData?.actual_cost).toFixed(6) }}</span>
          </div>
          <template v-if="tooltipData?.account_rate_multiplier != null || tooltipData?.account_stats_cost != null">
            <div class="flex items-center justify-between gap-6">
              <span class="text-gray-400">{{ t('usage.accountMultiplier') }}</span>
              <span class="font-semibold text-blue-400">{{ formatMultiplier(tooltipData.account_rate_multiplier ?? 1) }}x</span>
            </div>
            <div class="flex items-center justify-between gap-6">
              <span class="text-gray-400">{{ t('usage.accountBilled') }}</span>
              <span class="font-semibold text-green-400">${{ accountBilled(tooltipData).toFixed(6) }}</span>
            </div>
          </template>
        </div>
        <div class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-gray-900 border-t-transparent dark:border-r-gray-800"></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatDateTime, formatReasoningEffort } from '@/utils/format'
import { formatMultiplier } from '@/utils/formatters'
import { formatTokenPricePerMillion } from '@/utils/usagePricing'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import { getBillingModeBadgeClass, getBillingModeLabel, BILLING_MODE_IMAGE, BILLING_MODE_TOKEN } from '@/utils/billingMode'
import {
  formatImageBillingSize,
  formatImageInputSize,
  formatImageOutputSize,
  formatImageSizeBreakdown,
  formatImageSizeSource,
} from '@/utils/imageUsage'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUsageLog, UsageLog } from '@/types'
import type { Column } from '@/components/common/types'

type UsageTableIdentity = {
  id: number
  name?: string
  username?: string
  email?: string
}

type UsageTableUpstreamAttempt = {
  at_unix_ms?: number
  platform?: string
  account_id?: number
  account_name?: string
  upstream_status_code?: number
  upstream_request_id?: string
  upstream_url?: string
  kind?: string
  message?: string
  detail?: string
}

type NullableUsageNumberField =
  | 'user_id'
  | 'api_key_id'
  | 'account_id'
  | 'group_id'
  | 'input_tokens'
  | 'output_tokens'
  | 'cache_creation_tokens'
  | 'cache_read_tokens'
  | 'cache_creation_5m_tokens'
  | 'cache_creation_1h_tokens'
  | 'input_cost'
  | 'output_cost'
  | 'cache_creation_cost'
  | 'cache_read_cost'
  | 'total_cost'
  | 'actual_cost'
  | 'rate_multiplier'
  | 'billing_type'
  | 'duration_ms'
  | 'image_count'

type NullableUsageStringField =
  | 'inbound_endpoint'
  | 'upstream_endpoint'
  | 'image_size'
  | 'image_input_size'
  | 'image_output_size'
  | 'billing_mode'
  | 'user_agent'

type UsageTableRow = Omit<
  Partial<AdminUsageLog & UsageLog>,
  | 'account'
  | 'api_key'
  | 'group'
  | 'user'
  | NullableUsageNumberField
  | NullableUsageStringField
  | 'cache_ttl_overridden'
  | 'image_size_breakdown'
  | 'image_size_source'
> & {
  id?: number | string
  status_code?: number | null
  kind?: 'success' | 'pending' | 'error'
  billable?: boolean
  outcome?: string
  session_id?: string
  session_source?: string
  client_session_id?: string
  error_message?: string
  upstream_attempts?: UsageTableUpstreamAttempt[]
  phase?: string
  message?: string
  platform?: string | null
  cache_ttl_overridden?: boolean | null
  image_size_breakdown?: UsageLog['image_size_breakdown'] | null
  image_size_source?: UsageLog['image_size_source'] | null
  account?: UsageTableIdentity | null
  api_key?: UsageTableIdentity | null
  group?: UsageTableIdentity | null
  user?: UsageTableIdentity | null
} & Partial<Record<NullableUsageNumberField, number | null>>
  & Partial<Record<NullableUsageStringField, string | null>>

interface Props {
  data: UsageTableRow[]
  loading?: boolean
  columns: Column[]
  shellClass?: string
  serverSideSort?: boolean
  defaultSortKey?: string
  defaultSortOrder?: 'asc' | 'desc'
  showRequestTypeBadge?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  shellClass: 'overflow-hidden',
  serverSideSort: false,
  defaultSortKey: '',
  defaultSortOrder: 'desc',
  showRequestTypeBadge: true,
})

const emit = defineEmits<{
  sort: [key: string, order: 'asc' | 'desc']
  'session-click': [sessionId: string]
}>()

const { t } = useI18n()

const visibleColumns = computed(() => props.columns)
const currentSortKey = ref(props.defaultSortKey)
const currentSortOrder = ref<'asc' | 'desc'>(props.defaultSortOrder)

const tooltipVisible = ref(false)
const tooltipPosition = ref({ x: 0, y: 0 })
const tooltipData = ref<UsageTableRow | null>(null)
const tokenTooltipVisible = ref(false)
const tokenTooltipPosition = ref({ x: 0, y: 0 })
const tokenTooltipData = ref<UsageTableRow | null>(null)
const longTextTooltip = reactive({
  show: false,
  content: '',
  top: 0,
  left: 0,
  arrowLeft: 0,
  placement: 'top' as 'top' | 'bottom',
})

const num = (value: unknown, fallback = 0): number => {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback
}

const hasColumn = (key: string): boolean => props.columns.some((column) => column.key === key)

const rowKey = (row: UsageTableRow): string | number => {
  return row.id ?? row.request_id ?? `${row.created_at || 'unknown'}:${row.model || 'unknown'}`
}

const isRightAligned = (key: string): boolean => ['tokens', 'cache', 'cost', 'performance'].includes(key)

const getHeaderClass = (key: string): string => {
  return isRightAligned(key) ? 'text-right' : 'text-left'
}

const getCellClass = (key: string): string => {
  return isRightAligned(key) ? 'text-right' : ''
}

const toggleSort = (key: string) => {
  const nextOrder = currentSortKey.value === key && currentSortOrder.value === 'desc' ? 'asc' : 'desc'
  currentSortKey.value = key
  currentSortOrder.value = nextOrder
  emit('sort', key, nextOrder)
}

const formatDuration = (ms: number | null | undefined): string => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms.toFixed(0)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(2)}K`
  return value.toLocaleString()
}

const shortRequestId = (value?: string | null): string => {
  if (!value) return '-'
  return value.length > 12 ? `${value.slice(0, 8)}...${value.slice(-4)}` : value
}

const formatRelativeTime = (value?: string | null): string => {
  if (!value) return '-'
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

const formatUsageEndpoints = (log: UsageTableRow): string => {
  const inbound = log.inbound_endpoint?.trim()
  return inbound || '-'
}

const formatSessionTitle = (row: UsageTableRow): string => {
  const parts = [row.session_id, row.session_source, row.client_session_id].filter(Boolean)
  return parts.join(' · ') || '-'
}

const hasExtraSessionDetails = (row: UsageTableRow): boolean => {
  const sessionId = row.session_id?.trim()
  const clientSessionId = row.client_session_id?.trim()
  return Boolean(clientSessionId && clientSessionId !== sessionId)
}

const getUserDisplayName = (row: UsageTableRow): string => {
  return row.user?.username?.trim() || row.user?.email?.trim() || (row.user_id ? `#${row.user_id}` : '-')
}

const getUserTooltip = (row: UsageTableRow): string => {
  if (!row.user) return row.user_id ? `#${row.user_id}` : '-'
  const name = getUserDisplayName(row)
  return row.user.email && row.user.email !== name ? `${name} · ${row.user.email}` : name
}

const hasExtraUserDetails = (row: UsageTableRow): boolean => {
  if (!row.user?.email) return false
  return row.user.email !== getUserDisplayName(row)
}

const getAccountDisplayName = (row: UsageTableRow): string => {
  return row.account?.name?.trim() || (row.account_id ? `#${row.account_id}` : '-')
}

const getTotalTokens = (row: UsageTableRow | null): number => {
  return num(row?.input_tokens) + num(row?.output_tokens) + num(row?.cache_creation_tokens) + num(row?.cache_read_tokens)
}

const getCacheTokens = (row: UsageTableRow): number => {
  return num(row.cache_creation_tokens) + num(row.cache_read_tokens)
}

const hasUsageFacts = (row: UsageTableRow): boolean => {
  if (row.kind === 'error') return false
  return (
    row.input_tokens != null ||
    row.output_tokens != null ||
    row.cache_creation_tokens != null ||
    row.cache_read_tokens != null ||
    row.actual_cost != null ||
    row.image_count != null
  )
}

const hasBillableCost = (row: UsageTableRow): boolean => {
  return row.billable !== false && row.kind !== 'error' && row.actual_cost != null
}

const isImageUsage = (row: Pick<UsageTableRow, 'image_count'> | null | undefined): boolean => {
  return num(row?.image_count) > 0
}

const getDisplayBillingMode = (row: Pick<UsageTableRow, 'billing_mode' | 'image_count'> | null | undefined): string | null | undefined => {
  if (isImageUsage(row)) return BILLING_MODE_IMAGE
  return row?.billing_mode
}

const imageUnitPrice = (row: UsageTableRow | null): number => {
  const count = num(row?.image_count)
  if (!row || count <= 0) return 0
  const price = num(row.total_cost) / count
  return Number.isFinite(price) ? price : 0
}

const accountBilled = (row: UsageTableRow | null): number => {
  if (!row) return 0
  const base = row.account_stats_cost != null ? row.account_stats_cost : num(row.total_cost)
  const result = base * (row.account_rate_multiplier ?? 1)
  return Number.isFinite(result) ? result : 0
}

const getRequestTypeLabel = (row: UsageTableRow): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const getRequestTypeBadgeClass = (row: UsageTableRow): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'ws_v2') return 'bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-200'
  if (requestType === 'stream') return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
  if (requestType === 'sync') return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
}

const getUsageStatusLabel = (row: UsageTableRow): string => {
  if (row.outcome === 'pending' || row.kind === 'pending') return t('usage.statusPending')
  if (row.status_code != null) return String(row.status_code)
  if (row.kind === 'error') return t('usage.statusFailed')
  if (row.duration_ms === 0) return t('usage.statusPending')
  return t('usage.statusSuccess')
}

const isFinalSuccessRow = (row: UsageTableRow): boolean => {
  if (row.outcome === 'success' || row.outcome === 'non_billable' || row.kind === 'success') return true
  return row.status_code != null && row.status_code >= 200 && row.status_code < 300
}

const getStatusTooltipContent = (row: UsageTableRow): string => {
  if (isFinalSuccessRow(row)) return ''
  const errorMessage = row.error_message?.trim()
  if (errorMessage) return errorMessage
  const message = row.message?.trim()
  if (message) return message
  const phase = row.phase?.trim()
  if (phase) return phase
  return ''
}

const hasUpstreamAttempts = (row: UsageTableRow): boolean => {
  return Array.isArray(row.upstream_attempts) && row.upstream_attempts.length > 0
}

const formatUpstreamAttemptCount = (row: UsageTableRow): string => {
  const count = row.upstream_attempts?.length || 0
  return t('usage.upstreamAttemptCount', { count })
}

const formatUpstreamAttempt = (attempt: UsageTableUpstreamAttempt, index: number): string => {
  const parts: string[] = []
  const account = attempt.account_name?.trim() || (attempt.account_id ? `#${attempt.account_id}` : '')
  if (account) parts.push(account)
  if (attempt.upstream_status_code) {
    parts.push(t('usage.upstreamAttemptStatus', { status: attempt.upstream_status_code }))
  }
  if (attempt.kind?.trim()) parts.push(attempt.kind.trim())
  if (attempt.message?.trim()) parts.push(attempt.message.trim())
  if (attempt.upstream_url?.trim()) parts.push(attempt.upstream_url.trim())
  const summary = parts.join(' · ')
  return `${index + 1}. ${summary || '-'}`
}

const formatUpstreamAttempts = (row: UsageTableRow): string => {
  const attempts = row.upstream_attempts || []
  if (attempts.length === 0) return ''
  return [
    t('usage.upstreamAttempts'),
    ...attempts.map((attempt, index) => formatUpstreamAttempt(attempt, index)),
  ].join('\n')
}

const getUsageStatusClass = (row: UsageTableRow): string => {
  if (row.outcome === 'pending' || row.kind === 'pending') {
    return 'whitespace-nowrap bg-amber-50 px-2 py-0.5 text-[11px] font-medium text-amber-700 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/10 dark:text-amber-200 dark:ring-amber-400/20'
  }
  if (row.status_code != null) {
    if (row.status_code >= 200 && row.status_code < 300) return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-200'
    if (row.status_code >= 500) return 'bg-rose-100 text-rose-800 dark:bg-rose-500/15 dark:text-rose-200'
    return 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-200'
  }
  if (row.kind === 'error') return 'bg-rose-100 text-rose-800 dark:bg-rose-500/15 dark:text-rose-200'
  if (row.duration_ms === 0) return 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-200'
  return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-200'
}

const showTooltip = (event: MouseEvent | FocusEvent, row: UsageTableRow) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  tooltipData.value = row
  tooltipPosition.value.x = rect.right + 8
  tooltipPosition.value.y = rect.top + rect.height / 2
  tooltipVisible.value = true
}

const hideTooltip = () => {
  tooltipVisible.value = false
  tooltipData.value = null
}

const showTokenTooltip = (event: MouseEvent | FocusEvent, row: UsageTableRow) => {
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

const normalizeTooltipContent = (value: unknown): string => {
  if (value == null) return ''
  const content = String(value).trim()
  return content === '-' ? '' : content
}

type LongTextTooltipOptions = {
  force?: boolean
}

const isTooltipTargetOverflowing = (target: HTMLElement): boolean => {
  return target.scrollWidth > target.clientWidth + 1 || target.scrollHeight > target.clientHeight + 1
}

const estimateTooltipWidth = (content: string): number => {
  const padding = 24
  const estimatedTextWidth = Math.max(48, content.length * 7)
  const maxViewportWidth = Math.max(48, window.innerWidth - padding)
  return Math.min(360, maxViewportWidth, Math.max(48, estimatedTextWidth + padding))
}

const showLongTextTooltip = (
  event: MouseEvent | FocusEvent,
  value: unknown,
  options: LongTextTooltipOptions = {},
) => {
  const content = normalizeTooltipContent(value)
  if (!content) {
    hideLongTextTooltip()
    return
  }

  const target = event.currentTarget as HTMLElement | null
  if (!target) return
  if (!options.force && !isTooltipTargetOverflowing(target)) {
    hideLongTextTooltip()
    return
  }

  const rect = target.getBoundingClientRect()
  const padding = 12
  const gap = 8
  const tooltipWidth = estimateTooltipWidth(content)
  const tooltipHeightEstimate = Math.min(160, Math.max(56, Math.ceil(content.length / 40) * 18 + 28))
  const preferredLeft = rect.left + rect.width / 2 - tooltipWidth / 2
  const maxLeft = Math.max(padding, window.innerWidth - tooltipWidth - padding)
  const left = Math.max(padding, Math.min(preferredLeft, maxLeft))
  const hasTopSpace = rect.top >= tooltipHeightEstimate + gap + padding
  const placement = hasTopSpace ? 'top' : 'bottom'
  const top = placement === 'top'
    ? Math.max(padding, rect.top - gap)
    : Math.min(window.innerHeight - padding, rect.bottom + gap)
  const arrowLeft = Math.max(12, Math.min(rect.left + rect.width / 2 - left, tooltipWidth - 12))

  longTextTooltip.show = true
  longTextTooltip.content = content
  longTextTooltip.top = top
  longTextTooltip.left = left
  longTextTooltip.arrowLeft = arrowLeft
  longTextTooltip.placement = placement
}

const hideLongTextTooltip = () => {
  longTextTooltip.show = false
}
</script>

<style scoped>
.usage-table-long-tooltip {
  @apply pointer-events-none fixed z-[99999] max-w-[min(22.5rem,calc(100vw-1.5rem))] rounded-md bg-gray-900 px-3 py-2 text-left text-xs font-normal leading-relaxed text-white shadow-xl ring-1 ring-white/10 dark:bg-gray-700;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
  width: max-content;
}

.usage-table-long-tooltip-top {
  transform: translateY(-100%);
}

.usage-table-long-tooltip-bottom {
  transform: translateY(0);
}

.usage-table-long-tooltip-arrow {
  @apply absolute h-2 w-2 -translate-x-1/2 rotate-45 bg-gray-900 dark:bg-gray-700;
}

.usage-table-long-tooltip-arrow-top {
  @apply -bottom-1;
}

.usage-table-long-tooltip-arrow-bottom {
  @apply -top-1;
}
</style>
