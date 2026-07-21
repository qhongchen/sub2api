<template>
  <AppLayout>
    <div class="space-y-5">
      <section class="flex flex-wrap items-start justify-between gap-4">
        <div class="min-w-0">
          <h1 class="text-2xl font-semibold text-gray-950 dark:text-white">
            {{ t('admin.accounts.title') }}
          </h1>
          <p class="mt-1 max-w-2xl text-sm text-gray-500 dark:text-dark-300">
            {{ t('admin.accounts.description') }}
          </p>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <button @click="showCreate = true" class="btn btn-primary">
            <Icon name="plus" size="md" class="mr-2" />
            {{ t('admin.accounts.createAccount') }}
          </button>

          <div class="relative" ref="accountToolsDropdownRef">
            <button
              @click="
                showAccountToolsDropdown = !showAccountToolsDropdown;
                showColumnDropdown = false
              "
              class="btn btn-secondary px-2 md:px-3"
              :title="t('admin.accounts.moreActions')"
            >
              <Icon name="more" size="md" />
            </button>
            <div
              v-if="showAccountToolsDropdown"
              class="absolute right-0 z-50 mt-2 w-60 origin-top-right overflow-hidden rounded-lg border border-gray-200 bg-white p-1 shadow-lg dark:border-dark-700 dark:bg-dark-800"
            >
              <button class="account-tools-menu-item" @click="openSyncFromCrs">
                <span class="truncate">{{ t('admin.accounts.syncFromCrs') }}</span>
              </button>
              <button class="account-tools-menu-item" @click="openImportData">
                <span class="truncate">{{ t('admin.accounts.dataImport') }}</span>
              </button>
              <button class="account-tools-menu-item" @click="openExportDataDialogFromMenu">
                <span class="truncate">{{ t('admin.accounts.dataExport') }}</span>
              </button>
              <button class="account-tools-menu-item" @click="openErrorPassthrough">
                <span class="truncate">{{ t('admin.errorPassthrough.title') }}</span>
              </button>
              <button class="account-tools-menu-item" @click="openTLSFingerprintProfiles">
                <span class="truncate">{{ t('admin.tlsFingerprintProfiles.title') }}</span>
              </button>
              <div class="my-1 border-t border-gray-100 dark:border-dark-700" />
              <div class="space-y-3 px-3 py-2" @click.stop>
                <div class="flex items-center justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate text-sm font-medium text-gray-700 dark:text-gray-200">
                      {{ t('admin.accounts.upstreamBilling.autoProbeSettings') }}
                    </p>
                  </div>
                  <Toggle
                    v-model="upstreamBillingProbeSettings.enabled"
                    :aria-label="t('admin.accounts.upstreamBilling.autoProbeSettings')"
                  />
                </div>
                <label class="block text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.accounts.upstreamBilling.intervalMinutes') }}
                  <input
                    v-model.number="upstreamBillingProbeSettings.interval_minutes"
                    type="number"
                    min="5"
                    max="1440"
                    class="input mt-1 h-9"
                  />
                </label>
                <button
                  type="button"
                  class="btn btn-secondary btn-sm w-full"
                  :disabled="upstreamBillingSettingsLoading || upstreamBillingSettingsSaving"
                  @click="saveUpstreamBillingProbeSettings"
                >
                  {{ t('common.save') }}
                </button>
                <button
                  type="button"
                  class="btn btn-secondary btn-sm w-full"
                  :disabled="probingUpstreamBilling.size > 0"
                  @click="handleBulkProbeUpstreamBilling"
                >
                  {{ t('admin.accounts.bulkActions.probeUpstreamBilling') }}
                </button>
              </div>
            </div>
          </div>
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
                @click.stop="
                  showColumnDropdown = !showColumnDropdown;
                  showAccountToolsDropdown = false
                "
              >
                <Icon name="grid" size="sm" />
                <span>{{ visibleColumnSettingCount }}/{{ cardVisibilityOptions.length }}</span>
              </button>

              <div
                v-if="showColumnDropdown"
                class="absolute right-0 top-full z-50 mt-2 max-h-96 w-60 overflow-y-auto rounded-lg border border-gray-200 bg-white p-1 shadow-lg dark:border-dark-700 dark:bg-dark-800"
              >
                <div class="px-3 py-2 text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-dark-400">
                  {{ t('admin.requestLogs.columnSettings') }}
                </div>
                <button
                  v-for="field in cardVisibilityOptions"
                  :key="field.key"
                  type="button"
                  class="flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-left text-sm text-gray-600 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
                  @click="toggleCardField(field.key)"
                >
                  <span class="truncate">{{ field.label }}</span>
                  <Icon v-if="isCardFieldVisible(field.key)" name="check" size="sm" class="shrink-0 text-orange-500" :stroke-width="2" />
                </button>
                <div class="my-1 border-t border-gray-100 dark:border-white/[0.06]" />
                <button
                  type="button"
                  class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-400 dark:hover:bg-dark-700 dark:hover:text-white"
                  @click="resetCardVisibility"
                >
                  <Icon name="refresh" size="xs" />
                  {{ t('admin.requestLogs.resetColumns') }}
                </button>
              </div>
            </div>

            <button
              type="button"
              class="consumer-icon-btn"
              :title="t('common.refresh')"
              :disabled="accountListRefreshing"
              @click="handleManualRefresh"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': accountListRefreshing }" />
            </button>

            <button
              type="button"
              class="inline-flex h-9 items-center gap-2 rounded-full px-2 text-sm text-gray-500 transition-colors hover:text-gray-950 dark:text-dark-300 dark:hover:text-white"
              :title="autoRefreshTitle"
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
          <div v-if="filtersOpen" class="account-filter-card cch-panel-card p-4 md:p-5">
            <AccountTableFilters
              v-model:searchQuery="params.search"
              :filters="params"
              :groups="groups"
              @update:filters="(newFilters) => Object.assign(params, newFilters)"
              @change="debouncedReload"
              @reset="resetFilters"
            />
          </div>
        </Transition>

        <div
          v-if="hasPendingListSync"
          class="basis-full rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-800 dark:border-amber-700/40 dark:bg-amber-900/20 dark:text-amber-200 sm:flex sm:items-center sm:justify-between"
        >
          <span>{{ t('admin.accounts.listPendingSyncHint') }}</span>
          <button
            class="btn btn-secondary btn-sm px-2 py-1"
            @click="syncPendingListChanges"
          >
            {{ t('admin.accounts.listPendingSyncAction') }}
          </button>
        </div>
      </section>

      <TablePanel
        :summary="accountTableSummary"
        :summary-items="accountTableSummaryItems"
      >
        <div class="account-table-scroll">
          <table class="account-table">
            <thead class="account-table-header">
              <tr>
                <th
                  v-for="column in accountTableColumns"
                  :key="column.key"
                  class="account-table-head-cell"
                  :title="column.key === 'upstream_billing_rate'
                    ? t('admin.accounts.upstreamBilling.trustWarning')
                    : undefined"
                >
                  {{ column.label }}
                </th>
              </tr>
            </thead>
            <tbody>
              <template v-if="loading && accounts.length === 0">
                <tr
                  v-for="index in 5"
                  :key="`skeleton-${index}`"
                  class="account-table-row account-rich-row-skeleton"
                >
                  <td
                    v-for="column in accountTableColumns"
                    :key="`${column.key}-${index}`"
                    class="account-table-cell"
                  >
                    <div
                      class="h-4 rounded bg-gray-100 dark:bg-dark-700"
                      :class="column.key === 'account' ? 'w-56' : 'w-20'"
                    />
                  </td>
                </tr>
              </template>

              <tr v-else-if="accounts.length === 0">
                <td :colspan="accountTableColumnCount" class="account-table-empty-cell">
                  <div class="account-empty-state">
                    <div class="account-empty-icon">
                      <Icon name="server" size="lg" />
                    </div>
                    <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                      {{ t('admin.accounts.noAccountsYet') }}
                    </h3>
                    <p class="text-sm text-gray-500 dark:text-gray-400">
                      {{ t('admin.accounts.createFirstAccount') }}
                    </p>
                  </div>
                </td>
              </tr>

              <template v-else>
                <tr
                  v-for="account in accounts"
                  :key="account.id"
                  :class="['account-table-row', accountAccentClass(account)]"
                >
                  <td class="account-table-cell account-table-account-cell">
                    <div class="account-card-main">
                      <div class="account-card-title-row">
                        <span class="account-card-title" :title="account.name">{{ account.name }}</span>
                      </div>

                      <div class="account-card-subline">
                        <span v-if="getAccountEmail(account)" class="truncate" :title="getAccountEmail(account)">
                          {{ getAccountEmail(account) }}
                        </span>
                        <span v-if="account.proxy" class="account-inline-meta">
                          <Icon name="shield" size="xs" />
                          <span class="truncate">{{ account.proxy.name }}</span>
                          <span v-if="account.proxy.country_code">({{ account.proxy.country_code }})</span>
                        </span>
                        <span v-if="account.proxy && account.proxy.expires_at" :class="proxyExpiryBadge(account.proxy)">
                          {{ proxyExpiryText(account.proxy) }}
                        </span>
                        <span v-if="account.notes" class="truncate" :title="account.notes">
                          {{ account.notes }}
                        </span>
                      </div>

                      <div class="account-card-badges">
                        <button
                          v-if="isAccountTempUnschedulable(account)"
                          type="button"
                          :class="['account-status-chip', getAccountStatusClass(account)]"
                          @mouseenter="showAccountStatusTooltip(account, $event)"
                          @mouseleave="hideAccountStatusTooltip"
                          @focus="showAccountStatusTooltip(account, $event)"
                          @blur="hideAccountStatusTooltip"
                          @click="handleShowTempUnsched(account)"
                        >
                          {{ getAccountStatusText(account) }}
                        </button>
                        <span
                          v-else
                          :class="['account-status-chip', getAccountStatusClass(account)]"
                          tabindex="0"
                          @mouseenter="showAccountStatusTooltip(account, $event)"
                          @mouseleave="hideAccountStatusTooltip"
                          @focus="showAccountStatusTooltip(account, $event)"
                          @blur="hideAccountStatusTooltip"
                        >
                          {{ getAccountStatusText(account) }}
                        </span>
                        <PlatformTypeBadge
                          v-if="isCardFieldVisible('platform_type')"
                          mode="platformType"
                          :platform="account.platform"
                          :type="account.type"
                          :auth-mode="getOpenAIAuthMode(account)"
                        />
                        <span
                          v-if="isCardFieldVisible('platform_type') && getAntigravityTierLabel(account)"
                          :class="['account-tiny-badge', getAntigravityTierClass(account)]"
                        >
                          {{ getAntigravityTierLabel(account) }}
                        </span>
                        <div
                          v-if="isCardFieldVisible('platform_type') && getOpenAICompactMeta(account)"
                          :class="[
                            'account-compact-badge',
                            getOpenAICompactMeta(account)?.className
                          ]"
                          :title="getOpenAICompactTitle(account)"
                        >
                          <span :class="['h-1.5 w-1.5 rounded-full', getOpenAICompactMeta(account)?.dotClass]" />
                          <span>{{ getOpenAICompactMeta(account)?.label }}</span>
                        </div>
                        <AccountGroupsCell
                          v-if="!authStore.isSimpleMode && isCardFieldVisible('groups')"
                          :groups="account.groups"
                          :max-display="3"
                        />
                        <PlatformTypeBadge
                          v-if="isCardFieldVisible('account_plan') && hasAccountPlanDisplay(account)"
                          mode="planPrivacy"
                          :platform="account.platform"
                          :type="account.type"
                          :plan-type="getCredentialString(account, 'plan_type') || getAccountParentString(account, 'parent_plan_type')"
                          :privacy-mode="getExtraString(account, 'privacy_mode') || getAccountParentString(account, 'parent_privacy_mode')"
                          :subscription-expires-at="getCredentialString(account, 'subscription_expires_at') || getAccountParentString(account, 'parent_subscription_expires_at')"
                        />
                        <span v-if="isExpired(account.expires_at)" class="account-state-badge account-state-badge-warning">
                          {{ t('admin.accounts.expired') }}
                        </span>
                        <span
                          v-if="account.proxy_fallback_origin_id"
                          class="account-state-badge account-state-badge-warning"
                          :title="t('admin.accounts.fallbackActiveTip', { origin: account.proxy_fallback_origin_name || '-' })"
                        >
                          {{ t('admin.accounts.fallbackActive') }}
                        </span>
                        <button
                          v-if="account.proxy_fallback_origin_id"
                          type="button"
                          class="account-state-badge border border-amber-200 bg-white text-amber-700 transition hover:bg-amber-50 dark:border-amber-500/30 dark:bg-dark-900 dark:text-amber-300 dark:hover:bg-amber-500/10"
                          @click.stop="onRevertFallback(account)"
                        >
                          {{ t('admin.accounts.revertProxy') }}
                        </button>
                        <span
                          v-if="account.auto_pause_on_expired && account.expires_at"
                          class="account-state-badge account-state-badge-success"
                        >
                          {{ t('admin.accounts.autoPauseOnExpired') }}
                        </span>
                        <span v-if="account.expires_at" class="account-inline-meta">
                          <Icon name="clock" size="xs" />
                          {{ t('admin.accounts.columns.expiresAt') }} {{ formatExpiresAt(account.expires_at) }}
                        </span>
                      </div>
                    </div>
                  </td>

                  <td v-if="isCardFieldVisible('usage_windows')" class="account-table-cell account-table-usage-cell">
                    <AccountUsageCell
                      v-if="hasUsageWindowDisplay(account)"
                      :account="account"
                      :today-stats="todayStatsByAccountId[String(account.id)] ?? null"
                      :today-stats-loading="todayStatsLoading"
                      :manual-refresh-token="usageManualRefreshToken"
                      compact
                      @usage-refreshed="handleUsageRefreshed"
                    />
                    <span v-else class="account-table-muted">-</span>
                  </td>

                  <td v-if="isCardFieldVisible('priority')" class="account-table-cell">
                    <span class="account-table-metric-value">{{ account.priority }}</span>
                  </td>

                  <td v-if="isCardFieldVisible('scheduler_score')" class="account-table-cell">
                    <div
                      v-if="getSchedulerScoreRows(account).length"
                      class="flex min-w-[8rem] flex-col gap-1 font-mono text-[11px] leading-4"
                      :title="t('admin.accounts.schedulerScore.hint')"
                    >
                      <div
                        v-for="score in getSchedulerScoreRows(account)"
                        :key="score.group_id ?? 'ungrouped'"
                        class="whitespace-nowrap"
                      >
                        <span class="font-sans text-gray-500 dark:text-dark-300">
                          {{ formatSchedulerScoreGroup(score) }}
                        </span>
                        <span class="ml-1 text-gray-900 dark:text-dark-100">
                          {{ t('admin.accounts.schedulerScore.baseShort') }}
                          {{ formatSchedulerScore(score.base_score) }}
                        </span>
                        <span
                          v-if="score.sticky_weighted_enabled"
                          class="ml-1 text-orange-600 dark:text-orange-300"
                        >
                          {{ t('admin.accounts.schedulerScore.stickyShort') }}
                          {{ formatStickySchedulerScore(score) }}
                        </span>
                      </div>
                    </div>
                    <span v-else class="account-table-muted">-</span>
                  </td>

                  <td v-if="isCardFieldVisible('capacity')" class="account-table-cell">
                    <span class="account-table-metric-value">
                      {{ account.current_concurrency ?? 0 }}/{{ account.concurrency }}
                    </span>
                  </td>

                  <td v-if="isCardFieldVisible('billing_rate')" class="account-table-cell">
                    <span class="account-table-metric-value">{{ formatMultiplier(account.rate_multiplier) }}</span>
                  </td>

                  <td v-if="isCardFieldVisible('upstream_billing_rate')" class="account-table-cell">
                    <UpstreamBillingRateCell
                      :account="account"
                      :now="upstreamBillingNow"
                      :probing="probingUpstreamBilling.has(account.id)"
                      @probe="handleProbeUpstreamBilling(account)"
                    />
                  </td>

                  <td v-if="isCardFieldVisible('today_stats')" class="account-table-cell">
                    <div class="account-table-today-stat" tabindex="0">
                      <div class="account-table-metric-value">{{ formatTodayCost(account) }}</div>
                      <div class="account-metric-subvalue account-metric-subvalue-muted">
                        {{ t('admin.accounts.columns.lastUsed') }} {{ formatRelativeTime(account.last_used_at) }}
                      </div>
                      <span class="account-metric-detail-popover">
                        {{ formatTodayRequests(account) }} · {{ formatTodayTokens(account) }}
                      </span>
                    </div>
                  </td>

                  <td class="account-table-cell">
                    <button
                      :disabled="togglingSchedulable === account.id"
                      class="account-list-switch"
                      :class="[account.schedulable ? 'account-list-switch-on' : 'account-list-switch-off']"
                      :title="account.schedulable ? t('admin.accounts.schedulableEnabled') : t('admin.accounts.schedulableDisabled')"
                      @click="handleToggleSchedulable(account)"
                    >
                      <span :class="['account-list-switch-thumb', account.schedulable ? 'translate-x-4' : 'translate-x-0']" />
                    </button>
                  </td>

                  <td class="account-table-cell">
                    <div class="account-table-actions">
                      <button
                        type="button"
                        class="account-row-action"
                        :aria-label="t('common.edit')"
                        :title="t('common.edit')"
                        @click="handleEdit(account)"
                      >
                        <Icon name="edit" size="sm" />
                        <span class="account-row-action-label">{{ t('common.edit') }}</span>
                      </button>
                      <button
                        type="button"
                        class="account-row-action account-row-action-danger"
                        :aria-label="t('common.delete')"
                        :title="t('common.delete')"
                        @click="handleDelete(account)"
                      >
                        <Icon name="trash" size="sm" />
                        <span class="account-row-action-label">{{ t('common.delete') }}</span>
                      </button>
                      <button
                        type="button"
                        class="account-row-action"
                        :aria-label="t('common.more')"
                        :title="t('common.more')"
                        @click="openMenu(account, $event)"
                      >
                        <Icon name="more" size="sm" />
                        <span class="account-row-action-label">{{ t('common.more') }}</span>
                      </button>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </TablePanel>
    </div>
    <CreateAccountModal v-if="showCreate" :show="showCreate" :proxies="proxies" :groups="groups" @close="showCreate = false" @created="reload" />
    <EditAccountModal v-if="showEdit" :show="showEdit" :account="edAcc" :proxies="proxies" :groups="groups" @close="showEdit = false" @updated="handleAccountUpdated" />
    <ReAuthAccountModal :show="showReAuth" :account="reAuthAcc" @close="closeReAuthModal" @reauthorized="handleAccountUpdated" />
    <AccountTestModal :show="showTest" :account="testingAcc" @close="closeTestModal" />
    <AccountStatsModal v-if="showStats" :show="showStats" :account="statsAcc" @close="closeStatsModal" />
    <ScheduledTestsPanel :show="showSchedulePanel" :account-id="scheduleAcc?.id ?? null" :model-options="scheduleModelOptions" @close="closeSchedulePanel" />
    <AccountActionMenu :show="menu.show" :account="menu.acc" :position="menu.pos" @close="menu.show = false" @test="handleTest" @stats="handleViewStats" @schedule="handleSchedule" @duplicate="handleDuplicateAccount" @reauth="handleReAuth" @refresh-token="handleRefresh" @recover-state="handleRecoverState" @reset-quota="handleResetQuota" @set-privacy="handleSetPrivacy" @create-spark-shadow="handleCreateSparkShadow" />
    <Teleport to="body">
      <div
        v-if="accountStatusTooltip.show"
        role="tooltip"
        class="account-status-tooltip"
        :class="{
          'account-status-tooltip-top': accountStatusTooltip.placement === 'top',
          'account-status-tooltip-bottom': accountStatusTooltip.placement === 'bottom'
        }"
        :style="{ top: `${accountStatusTooltip.top}px`, left: `${accountStatusTooltip.left}px` }"
      >
        {{ accountStatusTooltip.content }}
        <span
          class="account-status-tooltip-arrow"
          :class="{
            'account-status-tooltip-arrow-top': accountStatusTooltip.placement === 'top',
            'account-status-tooltip-arrow-bottom': accountStatusTooltip.placement === 'bottom'
          }"
          :style="{ left: `${accountStatusTooltip.arrowLeft}px` }"
        />
      </div>
    </Teleport>
    <SyncFromCrsModal v-if="showSync" :show="showSync" @close="showSync = false" @synced="reload" />
    <ImportDataModal :show="showImportData" @close="showImportData = false" @imported="handleDataImported" />
    <TempUnschedStatusModal v-if="showTempUnsched" :show="showTempUnsched" :account="tempUnschedAcc" @close="showTempUnsched = false" @reset="handleTempUnschedReset" />
    <ConfirmDialog :show="showDeleteDialog" :title="t('admin.accounts.deleteAccount')" :message="t('admin.accounts.deleteConfirm', { name: deletingAcc?.name })" :confirm-text="t('common.delete')" :cancel-text="t('common.cancel')" :danger="true" @confirm="confirmDelete" @cancel="showDeleteDialog = false" />
    <ConfirmDialog :show="showCreateShadowDialog" :title="t('admin.accounts.createSparkShadow')" :message="t('admin.accounts.createSparkShadowConfirm', { name: creatingShadowAcc?.name })" @confirm="confirmCreateSparkShadow" @cancel="showCreateShadowDialog = false" />
    <ConfirmDialog :show="showExportDataDialog" :title="t('admin.accounts.dataExport')" :message="t('admin.accounts.dataExportConfirmMessage')" :confirm-text="t('admin.accounts.dataExportConfirm')" :cancel-text="t('common.cancel')" @confirm="handleExportData" @cancel="showExportDataDialog = false">
      <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
        <input type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" v-model="includeProxyOnExport" />
        <span>{{ t('admin.accounts.dataExportIncludeProxies') }}</span>
      </label>
    </ConfirmDialog>
    <ErrorPassthroughRulesModal :show="showErrorPassthrough" @close="showErrorPassthrough = false" />
    <TLSFingerprintProfilesModal :show="showTLSFingerprintProfiles" @close="showTLSFingerprintProfiles = false" />
    <TotpStepUpDialog :controller="accountExportStepUp" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, defineAsyncComponent, onMounted, onUnmounted, watch } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { adminAPI } from '@/api/admin'
import { useTableLoader } from '@/composables/useTableLoader'
import { useStepUp, isStepUpBlocked, isStepUpCancelled, stepUpBlockReason } from '@/composables/useStepUp'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import TablePanel from '@/components/common/TablePanel.vue'
import Toggle from '@/components/common/Toggle.vue'
import TotpStepUpDialog from '@/components/auth/TotpStepUpDialog.vue'
import AccountTableFilters from '@/components/admin/account/AccountTableFilters.vue'
import AccountActionMenu from '@/components/admin/account/AccountActionMenu.vue'
import ImportDataModal from '@/components/admin/account/ImportDataModal.vue'
import ReAuthAccountModal from '@/components/admin/account/ReAuthAccountModal.vue'
import AccountTestModal from '@/components/admin/account/AccountTestModal.vue'
import ScheduledTestsPanel from '@/components/admin/account/ScheduledTestsPanel.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import AccountGroupsCell from '@/components/account/AccountGroupsCell.vue'
import AccountUsageCell from '@/components/account/AccountUsageCell.vue'
import UpstreamBillingRateCell from '@/components/account/UpstreamBillingRateCell.vue'
import PlatformTypeBadge from '@/components/common/PlatformTypeBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import ErrorPassthroughRulesModal from '@/components/admin/ErrorPassthroughRulesModal.vue'
import TLSFingerprintProfilesModal from '@/components/admin/TLSFingerprintProfilesModal.vue'
import { buildOpenAIUsageRefreshKey } from '@/utils/accountUsageRefresh'
import { formatCountdown, formatCurrency, formatDateTime, formatNumber, formatRelativeTime, formatTokensK } from '@/utils/format'
import { proxyExpiryBadgeClass, proxyExpiryLabelKey } from '@/utils/proxyExpiry'
import { extractApiErrorMessage } from '@/utils/apiError'
import type {
  Account,
  AccountSchedulerGroupScore,
  Proxy as AccountProxy,
  AdminGroup,
  WindowStats,
  ClaudeModel,
  UpstreamBillingProbeSettings,
  UpstreamBillingProbeSnapshot
} from '@/types'

const CreateAccountModal = defineAsyncComponent(() => import('@/components/account/CreateAccountModal.vue'))
const EditAccountModal = defineAsyncComponent(() => import('@/components/account/EditAccountModal.vue'))
const SyncFromCrsModal = defineAsyncComponent(() => import('@/components/account/SyncFromCrsModal.vue'))
const TempUnschedStatusModal = defineAsyncComponent(() => import('@/components/account/TempUnschedStatusModal.vue'))
const AccountStatsModal = defineAsyncComponent(() => import('@/components/admin/account/AccountStatsModal.vue'))

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const proxies = ref<AccountProxy[]>([])
const groups = ref<AdminGroup[]>([])

const showCreate = ref(false)
const showEdit = ref(false)
const showSync = ref(false)
const showImportData = ref(false)
const showExportDataDialog = ref(false)
const includeProxyOnExport = ref(true)
const showTempUnsched = ref(false)
const showDeleteDialog = ref(false)
const showCreateShadowDialog = ref(false)
const showReAuth = ref(false)
const showTest = ref(false)
const showStats = ref(false)
const showErrorPassthrough = ref(false)
const showTLSFingerprintProfiles = ref(false)
const edAcc = ref<Account | null>(null)
const tempUnschedAcc = ref<Account | null>(null)
const deletingAcc = ref<Account | null>(null)
const creatingShadowAcc = ref<Account | null>(null)
const reAuthAcc = ref<Account | null>(null)
const testingAcc = ref<Account | null>(null)
const statsAcc = ref<Account | null>(null)
const showSchedulePanel = ref(false)
const scheduleAcc = ref<Account | null>(null)
const scheduleModelOptions = ref<SelectOption[]>([])
const togglingSchedulable = ref<number | null>(null)
const filtersOpen = ref(false)
const accountStatusTooltip = reactive({
  show: false,
  content: '',
  top: 0,
  left: 0,
  arrowLeft: 0,
  placement: 'top' as 'top' | 'bottom'
})
const menu = reactive<{show:boolean, acc:Account|null, pos:{top:number, left:number}|null}>({ show: false, acc: null, pos: null })
const exportingData = ref(false)
const accountExportStepUp = useStepUp()
const upstreamBillingProbeSettings = reactive<UpstreamBillingProbeSettings>({
  enabled: true,
  interval_minutes: 30
})
const upstreamBillingSettingsLoading = ref(false)
const upstreamBillingSettingsSaving = ref(false)
const probingUpstreamBilling = reactive(new Set<number>())
const upstreamBillingNow = ref(Date.now())
useIntervalFn(() => { upstreamBillingNow.value = Date.now() }, 60_000)

// Account tools dropdown
const showAccountToolsDropdown = ref(false)
const accountToolsDropdownRef = ref<HTMLElement | null>(null)
const showColumnDropdown = ref(false)
const columnDropdownRef = ref<HTMLElement | null>(null)

// Sorting settings
type AccountSortOrder = 'asc' | 'desc'
type AccountSortState = {
  sort_by: string
  sort_order: AccountSortOrder
}
const sortState = reactive<AccountSortState>({ sort_by: 'priority', sort_order: 'asc' })

// Auto refresh settings
const AUTO_REFRESH_STORAGE_KEY = 'account-auto-refresh'
const autoRefreshIntervals = [5, 10, 15, 30] as const
const autoRefreshEnabled = ref(false)
const autoRefreshIntervalSeconds = ref<(typeof autoRefreshIntervals)[number]>(30)
const autoRefreshCountdown = ref(0)
const autoRefreshETag = ref<string | null>(null)
const autoRefreshFetching = ref(false)
const AUTO_REFRESH_SILENT_WINDOW_MS = 15000
const autoRefreshSilentUntil = ref(0)
const hasPendingListSync = ref(false)
const todayStatsByAccountId = ref<Record<string, WindowStats>>({})
const todayStatsLoading = ref(false)
const todayStatsError = ref<string | null>(null)
const todayStatsReqSeq = ref(0)
const pendingTodayStatsRefresh = ref(false)
const usageManualRefreshToken = ref(0)
let usageRefreshedSyncTimer: ReturnType<typeof setTimeout> | null = null

const ACCOUNT_CARD_VISIBILITY_KEY = 'account-card-visible-fields'
const CARD_VISIBLE_FIELD_KEYS = [
  'platform_type',
  'account_plan',
  'groups',
  'priority',
  'scheduler_score',
  'capacity',
  'billing_rate',
  'upstream_billing_rate',
  'today_stats',
  'usage_windows'
] as const
type AccountCardFieldKey = typeof CARD_VISIBLE_FIELD_KEYS[number]
const DEFAULT_VISIBLE_CARD_FIELDS: AccountCardFieldKey[] = [
  'priority',
  'capacity',
  'billing_rate',
  'upstream_billing_rate',
  'today_stats',
  'usage_windows'
]
const cardVisibleFields = ref<Set<AccountCardFieldKey>>(new Set(DEFAULT_VISIBLE_CARD_FIELDS))

const cardVisibilityOptions = computed<Array<{ key: AccountCardFieldKey; label: string }>>(() => [
  { key: 'platform_type', label: t('admin.accounts.columns.platformType') },
  { key: 'account_plan', label: t('admin.accounts.columns.accountPlan') },
  { key: 'groups', label: t('admin.accounts.columns.groups') },
  { key: 'priority', label: t('admin.accounts.columns.priority') },
  { key: 'scheduler_score', label: t('admin.accounts.columns.schedulerScore') },
  { key: 'capacity', label: t('admin.accounts.columns.capacity') },
  { key: 'billing_rate', label: t('admin.accounts.columns.billingRateMultiplier') },
  { key: 'upstream_billing_rate', label: t('admin.accounts.columns.upstreamBillingRate') },
  { key: 'today_stats', label: t('admin.accounts.columns.todayStats') },
  { key: 'usage_windows', label: t('admin.accounts.columns.usageWindows') }
])

type AccountTableColumnKey =
  | 'account'
  | 'usage_windows'
  | 'priority'
  | 'scheduler_score'
  | 'capacity'
  | 'billing_rate'
  | 'upstream_billing_rate'
  | 'today_stats'
  | 'schedulable'
  | 'actions'

type AccountTableColumn = {
  key: AccountTableColumnKey
  label: string
}

const accountTableColumns = computed<AccountTableColumn[]>(() => [
  { key: 'account', label: t('admin.accounts.columns.name') },
  ...(isCardFieldVisible('usage_windows')
    ? [{ key: 'usage_windows' as const, label: t('admin.accounts.columns.usageWindows') }]
    : []),
  ...(isCardFieldVisible('priority')
    ? [{ key: 'priority' as const, label: t('admin.accounts.columns.priority') }]
    : []),
  ...(isCardFieldVisible('scheduler_score')
    ? [{ key: 'scheduler_score' as const, label: t('admin.accounts.columns.schedulerScore') }]
    : []),
  ...(isCardFieldVisible('capacity')
    ? [{ key: 'capacity' as const, label: t('admin.accounts.columns.capacity') }]
    : []),
  ...(isCardFieldVisible('billing_rate')
    ? [{ key: 'billing_rate' as const, label: t('admin.accounts.columns.billingRateMultiplier') }]
    : []),
  ...(isCardFieldVisible('upstream_billing_rate')
    ? [{ key: 'upstream_billing_rate' as const, label: t('admin.accounts.columns.upstreamBillingRate') }]
    : []),
  ...(isCardFieldVisible('today_stats')
    ? [{ key: 'today_stats' as const, label: t('admin.accounts.columns.todayStats') }]
    : []),
  { key: 'schedulable', label: t('admin.accounts.columns.schedulable') },
  { key: 'actions', label: t('admin.accounts.columns.actions') }
])

const accountTableColumnCount = computed(() => accountTableColumns.value.length)
const visibleColumnSettingCount = computed(() =>
  cardVisibilityOptions.value.filter(field => isCardFieldVisible(field.key)).length
)
const accountTableSummary = computed(() =>
  t('admin.requestLogs.loadedRecordsSummary', {
    count: formatNumber(accounts.value.length),
    total: formatNumber(pagination.total)
  })
)
const accountTableSummaryItems = computed(() => [
  `${formatNumber(pagination.total)} ${t('nav.accounts')}`,
  `${formatNumber(accounts.value.length)} ${t('common.visible')}`
])

const isAccountCardFieldKey = (value: unknown): value is AccountCardFieldKey => {
  return typeof value === 'string' && CARD_VISIBLE_FIELD_KEYS.includes(value as AccountCardFieldKey)
}

const loadSavedCardVisibility = () => {
  try {
    const saved = localStorage.getItem(ACCOUNT_CARD_VISIBILITY_KEY)
    if (!saved) return
    const parsed = JSON.parse(saved)
    if (!Array.isArray(parsed)) return
    const fields = parsed.filter(isAccountCardFieldKey)
    if (fields.includes('platform_type') && !fields.includes('account_plan')) {
      fields.push('account_plan')
    }
    cardVisibleFields.value = new Set(fields)
  } catch (e) {
    console.error('Failed to load saved card visibility:', e)
  }
}

const saveCardVisibility = () => {
  try {
    localStorage.setItem(ACCOUNT_CARD_VISIBILITY_KEY, JSON.stringify([...cardVisibleFields.value]))
  } catch (e) {
    console.error('Failed to save card visibility:', e)
  }
}

const isCardFieldVisible = (key: AccountCardFieldKey) => cardVisibleFields.value.has(key)

const toggleCardField = (key: AccountCardFieldKey) => {
  const next = new Set(cardVisibleFields.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  cardVisibleFields.value = next
  saveCardVisibility()
  if (key === 'scheduler_score') {
    load().catch((error) => {
      console.error('Failed to reload scheduler scores after changing column visibility:', error)
    })
  }
}

const resetCardVisibility = () => {
  const schedulerScoreWasVisible = isCardFieldVisible('scheduler_score')
  cardVisibleFields.value = new Set(DEFAULT_VISIBLE_CARD_FIELDS)
  saveCardVisibility()
  if (schedulerScoreWasVisible) {
    load().catch((error) => {
      console.error('Failed to reload accounts after resetting scheduler score visibility:', error)
    })
  }
}

const buildDefaultTodayStats = (): WindowStats => ({
  requests: 0,
  tokens: 0,
  cost: 0,
  standard_cost: 0,
  user_cost: 0
})

const refreshTodayStatsBatch = async () => {
  const accountIDs = accounts.value.map(account => account.id)
  const reqSeq = ++todayStatsReqSeq.value
  if (accountIDs.length === 0) {
    todayStatsByAccountId.value = {}
    todayStatsError.value = null
    todayStatsLoading.value = false
    return
  }

  todayStatsLoading.value = true
  todayStatsError.value = null

  try {
    const result = await adminAPI.accounts.getBatchTodayStats(accountIDs)
    if (reqSeq !== todayStatsReqSeq.value) return
    const serverStats = result.stats ?? {}
    const nextStats: Record<string, WindowStats> = {}
    for (const accountID of accountIDs) {
      const key = String(accountID)
      nextStats[key] = serverStats[key] ?? buildDefaultTodayStats()
    }
    todayStatsByAccountId.value = nextStats
  } catch (error) {
    if (reqSeq !== todayStatsReqSeq.value) return
    todayStatsError.value = 'Failed'
    console.error('Failed to load account today stats:', error)
  } finally {
    if (reqSeq === todayStatsReqSeq.value) {
      todayStatsLoading.value = false
    }
  }
}

const loadSavedAutoRefresh = () => {
  try {
    const saved = localStorage.getItem(AUTO_REFRESH_STORAGE_KEY)
    if (!saved) return
    const parsed = JSON.parse(saved) as { enabled?: boolean; interval_seconds?: number }
    autoRefreshEnabled.value = parsed.enabled === true
    const interval = Number(parsed.interval_seconds)
    if (autoRefreshIntervals.includes(interval as any)) {
      autoRefreshIntervalSeconds.value = interval as any
    }
  } catch (e) {
    console.error('Failed to load saved auto refresh settings:', e)
  }
}

const saveAutoRefreshToStorage = () => {
  try {
    localStorage.setItem(
      AUTO_REFRESH_STORAGE_KEY,
      JSON.stringify({
        enabled: autoRefreshEnabled.value,
        interval_seconds: autoRefreshIntervalSeconds.value
      })
    )
  } catch (e) {
    console.error('Failed to save auto refresh settings:', e)
  }
}

if (typeof window !== 'undefined') {
  loadSavedCardVisibility()
  loadSavedAutoRefresh()
}

const setAutoRefreshEnabled = (enabled: boolean) => {
  autoRefreshEnabled.value = enabled
  saveAutoRefreshToStorage()
  if (enabled) {
    autoRefreshCountdown.value = autoRefreshIntervalSeconds.value
    resumeAutoRefresh()
  } else {
    pauseAutoRefresh()
    autoRefreshCountdown.value = 0
  }
}

const {
  items: accounts,
  loading,
  params,
  pagination,
  load: baseLoad,
  reload: baseReload,
  debouncedReload: baseDebouncedReload
} = useTableLoader<Account, any>({
  fetchFn: adminAPI.accounts.list,
  pageSize: 1000,
  initialParams: {
    platform: '',
    type: '',
    status: '',
    privacy_mode: '',
    group: '',
    search: '',
    include_scheduler_score: isCardFieldVisible('scheduler_score') ? '1' : '0',
    sort_by: sortState.sort_by,
    sort_order: sortState.sort_order
  }
})

const resetAutoRefreshCache = () => {
  autoRefreshETag.value = null
}

const isFirstLoad = ref(true)

const syncAccountListDerivedParams = () => {
  const requestParams = params as any
  requestParams.include_scheduler_score = isCardFieldVisible('scheduler_score') ? '1' : '0'
}

const load = async () => {
  const requestParams = params as any
  syncAccountListDerivedParams()
  hasPendingListSync.value = false
  resetAutoRefreshCache()
  pendingTodayStatsRefresh.value = false
  if (isFirstLoad.value) {
    requestParams.lite = '1'
  }
  await baseLoad()
  if (isFirstLoad.value) {
    isFirstLoad.value = false
    delete requestParams.lite
  }
  await refreshTodayStatsBatch()
}

const reload = async () => {
  syncAccountListDerivedParams()
  hasPendingListSync.value = false
  resetAutoRefreshCache()
  pendingTodayStatsRefresh.value = false
  await baseReload()
  await refreshTodayStatsBatch()
}

const debouncedReload = () => {
  syncAccountListDerivedParams()
  hasPendingListSync.value = false
  resetAutoRefreshCache()
  pendingTodayStatsRefresh.value = true
  baseDebouncedReload()
}

const resetFilters = () => {
  Object.assign(params, {
    platform: '',
    type: '',
    status: '',
    privacy_mode: '',
    group: '',
    search: '',
  })
  debouncedReload()
}

const activeFilterCount = computed(() => {
  let count = 0
  if (params.search) count += 1
  if (params.platform) count += 1
  if (params.status) count += 1
  if (params.group) count += 1
  return count
})

const accountListRefreshing = computed(() => loading.value || autoRefreshFetching.value)
const autoRefreshTitle = computed(() =>
  autoRefreshEnabled.value
    ? t('admin.accounts.autoRefreshCountdown', { seconds: autoRefreshCountdown.value })
    : t('admin.accounts.autoRefresh')
)

const toggleAutoRefresh = () => {
  setAutoRefreshEnabled(!autoRefreshEnabled.value)
}

watch(loading, (isLoading, wasLoading) => {
  if (wasLoading && !isLoading && pendingTodayStatsRefresh.value) {
    pendingTodayStatsRefresh.value = false
    refreshTodayStatsBatch().catch((error) => {
      console.error('Failed to refresh account today stats after table load:', error)
    })
  }
})

const isAnyModalOpen = computed(() => {
  return (
    showCreate.value ||
    showEdit.value ||
    showSync.value ||
    showImportData.value ||
    showExportDataDialog.value ||
    showTempUnsched.value ||
    showDeleteDialog.value ||
    showReAuth.value ||
    showTest.value ||
    showStats.value ||
    showSchedulePanel.value ||
    showErrorPassthrough.value ||
    showTLSFingerprintProfiles.value
  )
})

const enterAutoRefreshSilentWindow = () => {
  autoRefreshSilentUntil.value = Date.now() + AUTO_REFRESH_SILENT_WINDOW_MS
  autoRefreshCountdown.value = autoRefreshIntervalSeconds.value
}

const inAutoRefreshSilentWindow = () => {
  return Date.now() < autoRefreshSilentUntil.value
}

const shouldReplaceAutoRefreshRow = (current: Account, next: Account) => {
  return (
    current.updated_at !== next.updated_at ||
    current.current_concurrency !== next.current_concurrency ||
    current.current_window_cost !== next.current_window_cost ||
    current.active_sessions !== next.active_sessions ||
    current.schedulable !== next.schedulable ||
    current.status !== next.status ||
    current.rate_limit_reset_at !== next.rate_limit_reset_at ||
    current.overload_until !== next.overload_until ||
    current.temp_unschedulable_until !== next.temp_unschedulable_until ||
    (
      isCardFieldVisible('scheduler_score') &&
      JSON.stringify(current.scheduler_scores ?? current.scheduler_score ?? null) !==
        JSON.stringify(next.scheduler_scores ?? next.scheduler_score ?? null)
    ) ||
    buildOpenAIUsageRefreshKey(current) !== buildOpenAIUsageRefreshKey(next) ||
    JSON.stringify(current.extra?.upstream_billing_probe ?? null) !==
      JSON.stringify(next.extra?.upstream_billing_probe ?? null)
  )
}

const syncAccountRefs = (nextAccount: Account) => {
  if (edAcc.value?.id === nextAccount.id) edAcc.value = nextAccount
  if (reAuthAcc.value?.id === nextAccount.id) reAuthAcc.value = nextAccount
  if (tempUnschedAcc.value?.id === nextAccount.id) tempUnschedAcc.value = nextAccount
  if (deletingAcc.value?.id === nextAccount.id) deletingAcc.value = nextAccount
  if (menu.acc?.id === nextAccount.id) menu.acc = nextAccount
}

const mergeAccountsIncrementally = (nextRows: Account[]) => {
  const currentRows = accounts.value
  const currentByID = new Map(currentRows.map(row => [row.id, row]))
  let changed = nextRows.length !== currentRows.length
  const mergedRows = nextRows.map((nextRow) => {
    const currentRow = currentByID.get(nextRow.id)
    if (!currentRow) {
      changed = true
      return nextRow
    }
    if (shouldReplaceAutoRefreshRow(currentRow, nextRow)) {
      changed = true
      syncAccountRefs(nextRow)
      return nextRow
    }
    return currentRow
  })
  if (!changed) {
    for (let i = 0; i < mergedRows.length; i += 1) {
      if (mergedRows[i].id !== currentRows[i]?.id) {
        changed = true
        break
      }
    }
  }
  if (changed) {
    accounts.value = mergedRows
  }
}

const refreshAccountsIncrementally = async () => {
  if (autoRefreshFetching.value) return
  autoRefreshFetching.value = true
  try {
    syncAccountListDerivedParams()
    const result = await adminAPI.accounts.listWithEtag(
      pagination.page,
      pagination.page_size,
      { ...params } as {
        platform?: string
        type?: string
        status?: string
        privacy_mode?: string
        group?: string
        search?: string
        include_scheduler_score?: string
        sort_by?: string
        sort_order?: AccountSortOrder

      },
      { etag: autoRefreshETag.value }
    )

    if (result.etag) {
      autoRefreshETag.value = result.etag
    }
    if (!result.notModified && result.data) {
      pagination.total = result.data.total || 0
      pagination.pages = result.data.pages || 0
      mergeAccountsIncrementally(result.data.items || [])
      hasPendingListSync.value = false
    }

    await refreshTodayStatsBatch()
  } catch (error) {
    console.error('Auto refresh failed:', error)
  } finally {
    autoRefreshFetching.value = false
  }
}

const handleManualRefresh = async () => {
  await load()
  usageManualRefreshToken.value += 1
}

const handleUsageRefreshed = () => {
  if (usageRefreshedSyncTimer) {
    clearTimeout(usageRefreshedSyncTimer)
  }
  usageRefreshedSyncTimer = setTimeout(() => {
    usageRefreshedSyncTimer = null
    refreshAccountsIncrementally().catch((error) => {
      console.error('Failed to sync accounts after usage refresh:', error)
    })
  }, 150)
}

const closeAccountToolsDropdown = () => {
  showAccountToolsDropdown.value = false
}

const openSyncFromCrs = () => {
  closeAccountToolsDropdown()
  showSync.value = true
}

const openImportData = () => {
  closeAccountToolsDropdown()
  showImportData.value = true
}

const openExportDataDialogFromMenu = () => {
  closeAccountToolsDropdown()
  openExportDataDialog()
}

const openErrorPassthrough = () => {
  closeAccountToolsDropdown()
  showErrorPassthrough.value = true
}

const openTLSFingerprintProfiles = () => {
  closeAccountToolsDropdown()
  showTLSFingerprintProfiles.value = true
}

const loadUpstreamBillingProbeSettings = async () => {
  upstreamBillingSettingsLoading.value = true
  try {
    Object.assign(upstreamBillingProbeSettings, await adminAPI.accounts.getUpstreamBillingProbeSettings())
  } catch (error) {
    console.error('Failed to load upstream billing probe settings:', error)
  } finally {
    upstreamBillingSettingsLoading.value = false
  }
}

const saveUpstreamBillingProbeSettings = async () => {
  upstreamBillingSettingsSaving.value = true
  try {
    const saved = await adminAPI.accounts.updateUpstreamBillingProbeSettings({
      enabled: upstreamBillingProbeSettings.enabled,
      interval_minutes: Math.min(
        1440,
        Math.max(5, Number(upstreamBillingProbeSettings.interval_minutes) || 30)
      )
    })
    Object.assign(upstreamBillingProbeSettings, saved)
    appStore.showSuccess(t('admin.accounts.upstreamBilling.settingsSaved'))
  } catch (error) {
    console.error('Failed to save upstream billing probe settings:', error)
    appStore.showError(extractApiErrorMessage(error, t('admin.accounts.upstreamBilling.settingsFailed')))
  } finally {
    upstreamBillingSettingsSaving.value = false
  }
}

const syncPendingListChanges = async () => {
  hasPendingListSync.value = false
  await load()
  usageManualRefreshToken.value += 1
}

const { pause: pauseAutoRefresh, resume: resumeAutoRefresh } = useIntervalFn(
  async () => {
    if (!autoRefreshEnabled.value) return
    if (document.hidden) return
    if (loading.value || autoRefreshFetching.value) return
    if (isAnyModalOpen.value) return
    if (menu.show || showAccountToolsDropdown.value) return
    if (inAutoRefreshSilentWindow()) {
      autoRefreshCountdown.value = Math.max(
        0,
        Math.ceil((autoRefreshSilentUntil.value - Date.now()) / 1000)
      )
      return
    }

    if (autoRefreshCountdown.value <= 0) {
      autoRefreshCountdown.value = autoRefreshIntervalSeconds.value
      await refreshAccountsIncrementally()
      return
    }

    autoRefreshCountdown.value -= 1
  },
  1000,
  { immediate: false }
)

// Antigravity 订阅等级辅助函数
function getAntigravityTierFromRow(row: any): string | null {
  if (row.platform !== 'antigravity') return null
  const extra = row.extra as Record<string, unknown> | undefined
  if (!extra) return null
  const lca = extra.load_code_assist as Record<string, unknown> | undefined
  if (!lca) return null
  const paid = lca.paidTier as Record<string, unknown> | undefined
  if (paid && typeof paid.id === 'string') return paid.id
  const current = lca.currentTier as Record<string, unknown> | undefined
  if (current && typeof current.id === 'string') return current.id
  return null
}

function getAntigravityTierLabel(row: any): string | null {
  const tier = getAntigravityTierFromRow(row)
  switch (tier) {
    case 'free-tier': return t('admin.accounts.tier.free')
    case 'g1-pro-tier': return t('admin.accounts.tier.pro')
    case 'g1-ultra-tier': return t('admin.accounts.tier.ultra')
    default: return null
  }
}

type OpenAICompactBadgeState = 'active' | 'blocked' | 'auto'

function getOpenAICompactState(row: any): OpenAICompactBadgeState | null {
  if (row.platform !== 'openai' || (row.type !== 'oauth' && row.type !== 'apikey')) return null
  const extra = row.extra as Record<string, unknown> | undefined
  const mode = typeof extra?.openai_compact_mode === 'string' ? extra.openai_compact_mode : 'auto'
  if (mode === 'force_on') return 'active'
  if (mode === 'force_off') return 'blocked'
  if (typeof extra?.openai_compact_supported === 'boolean') {
    return extra.openai_compact_supported ? 'active' : 'blocked'
  }
  return 'auto'
}

function getOpenAICompactMeta(row: any): { label: string; className: string; dotClass: string } | null {
  const state = getOpenAICompactState(row)
  if (!state) return null
  switch (state) {
    case 'active':
      return {
        label: t('admin.accounts.openai.compactSupported'),
        className: 'text-emerald-600 dark:text-emerald-300',
        dotClass: 'bg-emerald-500 shadow-[0_0_0_2px_rgba(16,185,129,0.14)]'
      }
    case 'blocked':
      return {
        label: t('admin.accounts.openai.compactUnsupported'),
        className: 'text-rose-600 dark:text-rose-300',
        dotClass: 'bg-rose-500 shadow-[0_0_0_2px_rgba(244,63,94,0.14)]'
      }
    case 'auto':
      return {
        label: t('admin.accounts.openai.compactAuto'),
        className: 'text-slate-500 dark:text-slate-400',
        dotClass: 'bg-slate-300 dark:bg-slate-500'
      }
  }
}

function getOpenAICompactTitle(row: any): string {
  const extra = row.extra as Record<string, unknown> | undefined
  const checkedAt = typeof extra?.openai_compact_checked_at === 'string' ? extra.openai_compact_checked_at : ''
  const label = getOpenAICompactMeta(row)?.label || ''
  if (!checkedAt) return label
  return `${label} | ${t('admin.accounts.openai.compactLastChecked')}: ${formatDateTime(new Date(checkedAt))}`
}

function getAntigravityTierClass(row: any): string {
  const tier = getAntigravityTierFromRow(row)
  switch (tier) {
    case 'free-tier': return 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
    case 'g1-pro-tier': return 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300'
    case 'g1-ultra-tier': return 'bg-purple-100 text-purple-600 dark:bg-purple-900/40 dark:text-purple-300'
    default: return ''
  }
}

const getRecordString = (record: Record<string, unknown> | undefined, key: string) => {
  const value = record?.[key]
  return typeof value === 'string' ? value : ''
}

const getCredentialString = (account: Account, key: string) => {
  return getRecordString(account.credentials, key)
}

const getOpenAIAuthMode = (account: Account): string | undefined => {
  if (account.platform !== 'openai' || account.type !== 'oauth') return undefined
  const value = account.credentials?.auth_mode
  return typeof value === 'string' && value.trim() ? value : undefined
}

const getExtraString = (account: Account, key: string) => {
  return getRecordString(account.extra, key)
}

const getAccountParentString = (account: Account, key: string) => {
  const value = (account as unknown as Record<string, unknown>)[key]
  return typeof value === 'string' ? value : ''
}

const isAccountPrivacyModeDisplayable = (account: Account) => {
  if (account.type !== 'oauth') return false
  if (account.platform !== 'openai' && account.platform !== 'antigravity') return false
  return [
    'training_off',
    'training_set_cf_blocked',
    'training_set_failed',
    'privacy_set',
    'privacy_set_failed'
  ].includes(getExtraString(account, 'privacy_mode'))
}

const hasAccountPlanDisplay = (account: Account) => {
  return Boolean(
    getCredentialString(account, 'plan_type') ||
    getAccountParentString(account, 'parent_plan_type') ||
    isAccountPrivacyModeDisplayable(account) ||
    getAccountParentString(account, 'parent_privacy_mode') ||
    getCredentialString(account, 'subscription_expires_at') ||
    getAccountParentString(account, 'parent_subscription_expires_at')
  )
}

const getAccountEmail = (account: Account) => {
  return (
    getExtraString(account, 'email_address') ||
    getExtraString(account, 'email') ||
    getCredentialString(account, 'email') ||
    getAccountParentString(account, 'parent_email')
  )
}

const getTodayStats = (account: Account) => todayStatsByAccountId.value[String(account.id)] ?? buildDefaultTodayStats()

const formatMultiplier = (value: number | null | undefined) => {
  return `${(Number(value ?? 1) || 0).toFixed(2)}x`
}

const formatSchedulerScore = (value: number | null | undefined) => {
  const score = Number(value)
  if (!Number.isFinite(score)) return '-'
  return score.toFixed(6).replace(/\.?0+$/, '')
}

const formatStickySchedulerScore = (score: AccountSchedulerGroupScore) => {
  if (score.sticky_score_infinity) return '+∞'
  return formatSchedulerScore(score.sticky_score)
}

const getSchedulerScoreRows = (account: Account): AccountSchedulerGroupScore[] => {
  const groupRows = Array.isArray(account.scheduler_scores)
    ? account.scheduler_scores.filter(score => score.group_id != null)
    : []
  if (groupRows.length) return groupRows
  if (account.scheduler_score) {
    return [{ group_id: null, ...account.scheduler_score }]
  }
  return []
}

const formatSchedulerScoreGroup = (score: AccountSchedulerGroupScore) => {
  if (score.group_name) return score.group_name
  if (score.group_id != null) return `#${score.group_id}`
  return t('admin.accounts.schedulerScore.ungrouped')
}

const formatTodayRequests = (account: Account) => {
  const stats = getTodayStats(account)
  if (todayStatsLoading.value && !todayStatsByAccountId.value[String(account.id)]) return '...'
  if (todayStatsError.value && !todayStatsByAccountId.value[String(account.id)]) return '-'
  return `${stats.requests.toLocaleString()} ${t('admin.dashboard.requestsShort')}`
}

const formatTodayTokens = (account: Account) => {
  const stats = getTodayStats(account)
  return `${formatTokensK(stats.tokens)} ${t('admin.dashboard.tokensShort')}`
}

const formatTodayCost = (account: Account) => {
  const stats = getTodayStats(account)
  return formatCurrency(stats.cost)
}

const hasUsageWindowDisplay = (account: Account) => {
  if (account.platform === 'gemini') return true
  if (account.type === 'oauth' || account.type === 'setup-token') return true
  if (account.type === 'apikey' || account.type === 'bedrock') {
    return (
      (account.quota_daily_limit ?? 0) > 0 ||
      (account.quota_weekly_limit ?? 0) > 0 ||
      (account.quota_limit ?? 0) > 0
    )
  }
  return false
}

const isFutureTime = (value: string | null | undefined) => {
  if (!value) return false
  const time = new Date(value).getTime()
  return Number.isFinite(time) && time > Date.now()
}

const isAccountRateLimited = (account: Account) => isFutureTime(account.rate_limit_reset_at)

const isAccountOverloaded = (account: Account) => isFutureTime(account.overload_until)

const isAccountTempUnschedulable = (account: Account) => isFutureTime(account.temp_unschedulable_until)

const isAccountQuotaExceeded = (account: Account) => {
  const exceeded = (used?: number | null, limit?: number | null) =>
    typeof limit === 'number' && limit > 0 && typeof used === 'number' && used >= limit
  return (
    exceeded(account.quota_used, account.quota_limit) ||
    exceeded(account.quota_daily_used, account.quota_daily_limit) ||
    exceeded(account.quota_weekly_used, account.quota_weekly_limit)
  )
}

const getActiveModelLimitCount = (account: Account) => {
  const modelLimits = account.extra?.model_rate_limits
  if (!modelLimits) return 0
  const now = Date.now()
  return Object.values(modelLimits).filter(item => new Date(item.rate_limit_reset_at).getTime() > now).length
}

const hasAccountWarning = (account: Account) => {
  return (
    account.status === 'error' ||
    isAccountRateLimited(account) ||
    isAccountOverloaded(account) ||
    isAccountTempUnschedulable(account) ||
    isAccountQuotaExceeded(account) ||
    getActiveModelLimitCount(account) > 0
  )
}

const accountAccentClass = (account: Account) => {
  if (!account.schedulable) return 'account-rich-row-muted'
  if (account.status === 'error') return 'account-rich-row-error'
  if (hasAccountWarning(account)) return 'account-rich-row-warning'
  if (account.status !== 'active') return 'account-rich-row-muted'
  return 'account-rich-row-active'
}

const getAccountStatusText = (account: Account) => {
  if (!account.schedulable) return t('admin.accounts.status.paused')
  if (isAccountRateLimited(account)) return t('admin.accounts.status.rateLimited')
  if (isAccountOverloaded(account)) return t('admin.accounts.status.overloaded')
  if (account.status === 'error') return t('admin.accounts.status.error')
  if (isAccountTempUnschedulable(account)) return t('admin.accounts.status.tempUnschedulable')
  if (account.status !== 'active') return t(`admin.accounts.status.${account.status}`)
  if (isAccountQuotaExceeded(account)) return t('admin.accounts.status.quotaExceeded')
  const modelLimitCount = getActiveModelLimitCount(account)
  if (modelLimitCount > 0) return t('admin.accounts.status.modelLimitedCount', { count: modelLimitCount })
  return t('admin.accounts.status.active')
}

const getAccountStatusClass = (account: Account) => {
  if (!account.schedulable) return 'account-status-chip-muted'
  if (account.status === 'error') return 'account-status-chip-error'
  if (hasAccountWarning(account)) return 'account-status-chip-warning'
  if (account.status !== 'active') return 'account-status-chip-muted'
  return 'account-status-chip-success'
}

const parseFutureTimestamp = (value: unknown): number | null => {
  if (!value) return null
  const timestamp = new Date(String(value)).getTime()
  if (!Number.isFinite(timestamp) || timestamp <= Date.now()) return null
  return timestamp
}

const getOpenAIRateLimitResetAt = (account: Account): string | null => {
  if (account.platform !== 'openai' || account.type !== 'oauth') {
    return account.rate_limit_reset_at
  }

  const extra = account.extra ?? {}
  const windows = [
    {
      usedPercent: typeof extra.codex_5h_used_percent === 'number' ? extra.codex_5h_used_percent : null,
      resetAt: typeof extra.codex_5h_reset_at === 'string' ? extra.codex_5h_reset_at : null,
    },
    {
      usedPercent: typeof extra.codex_7d_used_percent === 'number' ? extra.codex_7d_used_percent : null,
      resetAt: typeof extra.codex_7d_reset_at === 'string' ? extra.codex_7d_reset_at : null,
    },
  ]

  const activeWindowResetAt = windows
    .filter((window) => window.usedPercent != null && window.usedPercent >= 100)
    .map((window) => ({ raw: window.resetAt, ts: parseFutureTimestamp(window.resetAt) }))
    .filter((window): window is { raw: string; ts: number } => !!window.raw && window.ts != null)
    .sort((a, b) => a.ts - b.ts)[0]

  return activeWindowResetAt?.raw || account.rate_limit_reset_at
}

const getAccountStatusDetail = (account: Account) => {
  if (!account.schedulable) return ''
  if (isAccountRateLimited(account)) {
    const rateLimitResetAt = getOpenAIRateLimitResetAt(account)
    const countdown = formatCountdown(rateLimitResetAt)
    const resumeText = countdown ? t('admin.accounts.status.rateLimitedAutoResume', { time: countdown }) : ''
    return ['429', resumeText, formatDateTime(rateLimitResetAt)].filter(Boolean).join(' · ')
  }
  if (isAccountOverloaded(account)) {
    const countdown = formatCountdown(account.overload_until)
    return ['529', countdown, formatDateTime(account.overload_until)].filter(Boolean).join(' · ')
  }
  if (isAccountTempUnschedulable(account)) {
    const countdown = formatCountdown(account.temp_unschedulable_until)
    return [account.temp_unschedulable_reason, countdown, formatDateTime(account.temp_unschedulable_until)].filter(Boolean).join(' · ')
  }
  if (account.status === 'error') return account.error_message || ''
  const modelLimitCount = getActiveModelLimitCount(account)
  if (modelLimitCount > 0) {
    return t('admin.accounts.status.modelLimitedDetail', { count: modelLimitCount })
  }
  return ''
}

const hideAccountStatusTooltip = () => {
  accountStatusTooltip.show = false
}

const showAccountStatusTooltip = (account: Account, event: MouseEvent | FocusEvent) => {
  const detail = getAccountStatusDetail(account)
  if (!detail) {
    hideAccountStatusTooltip()
    return
  }

  const target = event.currentTarget as HTMLElement | null
  if (!target) return

  const rect = target.getBoundingClientRect()
  const padding = 12
  const gap = 8
  const estimatedTextWidth = Math.max(48, detail.length * 7)
  const maxViewportWidth = Math.max(48, window.innerWidth - padding * 2)
  const tooltipWidth = Math.min(360, maxViewportWidth, Math.max(48, estimatedTextWidth + 24))
  const tooltipHeightEstimate = Math.min(160, Math.max(56, Math.ceil(detail.length / 40) * 18 + 28))
  const preferredLeft = rect.left + rect.width / 2 - tooltipWidth / 2
  const maxLeft = Math.max(padding, window.innerWidth - tooltipWidth - padding)
  const left = Math.max(padding, Math.min(preferredLeft, maxLeft))
  const hasTopSpace = rect.top >= tooltipHeightEstimate + gap + padding
  const placement = hasTopSpace ? 'top' : 'bottom'
  const top = placement === 'top'
    ? Math.max(padding, rect.top - gap)
    : Math.min(window.innerHeight - padding, rect.bottom + gap)
  const arrowLeft = Math.max(12, Math.min(rect.left + rect.width / 2 - left, tooltipWidth - 12))

  accountStatusTooltip.show = true
  accountStatusTooltip.content = detail
  accountStatusTooltip.top = top
  accountStatusTooltip.left = left
  accountStatusTooltip.arrowLeft = arrowLeft
  accountStatusTooltip.placement = placement
}

const handleEdit = (a: Account) => { edAcc.value = a; showEdit.value = true }
const openMenu = (a: Account, e: MouseEvent) => {
  menu.acc = a

  const target = e.currentTarget as HTMLElement
  if (target) {
    const rect = target.getBoundingClientRect()
    const menuWidth = 200
    const menuHeight = 240
    const padding = 8
    const viewportWidth = window.innerWidth
    const viewportHeight = window.innerHeight

    let left: number
    let top: number

    if (viewportWidth < 768) {
      // 居中显示,水平位置
      left = Math.max(padding, Math.min(
        rect.left + rect.width / 2 - menuWidth / 2,
        viewportWidth - menuWidth - padding
      ))

      // 优先显示在按钮下方
      top = rect.bottom + 4

      // 如果下方空间不够,显示在上方
      if (top + menuHeight > viewportHeight - padding) {
        top = rect.top - menuHeight - 4
        // 如果上方也不够,就贴在视口顶部
        if (top < padding) {
          top = padding
        }
      }
    } else {
      left = Math.max(padding, Math.min(
        e.clientX - menuWidth,
        viewportWidth - menuWidth - padding
      ))
      top = e.clientY
      if (top + menuHeight > viewportHeight - padding) {
        top = viewportHeight - menuHeight - padding
      }
    }

    menu.pos = { top, left }
  } else {
    menu.pos = { top: e.clientY, left: e.clientX - 200 }
  }

  menu.show = true
}

const updateSchedulableInList = (accountIds: number[], schedulable: boolean) => {
  if (accountIds.length === 0) return
  const idSet = new Set(accountIds)
  accounts.value = accounts.value.map((account) => (idSet.has(account.id) ? { ...account, schedulable } : account))
}
const handleDataImported = () => { showImportData.value = false; reload() }
const ACCOUNT_UNGROUPED_GROUP_QUERY_VALUE = 'ungrouped'
const ACCOUNT_PRIVACY_MODE_UNSET_QUERY_VALUE = '__unset__'
const buildAccountQueryFilters = () => ({
  platform: params.platform || '',
  type: params.type || '',
  status: params.status || '',
  group: params.group || '',
  privacy_mode: params.privacy_mode || '',
  search: params.search || '',
  sort_by: sortState.sort_by,
  sort_order: sortState.sort_order
})
const accountMatchesCurrentFilters = (account: Account) => {
  const filters = buildAccountQueryFilters()
  if (filters.platform && account.platform !== filters.platform) return false
  if (filters.type && account.type !== filters.type) return false
  if (filters.status) {
    const now = Date.now()
    const rateLimitResetAt = account.rate_limit_reset_at ? new Date(account.rate_limit_reset_at).getTime() : Number.NaN
    const isRateLimited = Number.isFinite(rateLimitResetAt) && rateLimitResetAt > now
    const tempUnschedUntil = account.temp_unschedulable_until ? new Date(account.temp_unschedulable_until).getTime() : Number.NaN
    const isTempUnschedulable = Number.isFinite(tempUnschedUntil) && tempUnschedUntil > now

    if (filters.status === 'active') {
      if (account.status !== 'active' || isRateLimited || isTempUnschedulable || !account.schedulable) return false
    } else if (filters.status === 'rate_limited') {
      if (account.status !== 'active' || !isRateLimited || isTempUnschedulable) return false
    } else if (filters.status === 'temp_unschedulable') {
      if (account.status !== 'active' || !isTempUnschedulable) return false
    } else if (filters.status === 'unschedulable') {
      if (account.status !== 'active' || account.schedulable || isRateLimited || isTempUnschedulable) return false
    } else if (account.status !== filters.status) {
      return false
    }
  }
  if (filters.group) {
    const groupIds = account.group_ids ?? account.groups?.map((group) => group.id) ?? []
    if (filters.group === ACCOUNT_UNGROUPED_GROUP_QUERY_VALUE) {
      if (groupIds.length > 0) return false
    } else if (!groupIds.includes(Number(filters.group))) {
      return false
    }
  }
  const privacyMode = typeof account.extra?.privacy_mode === 'string' ? account.extra.privacy_mode : ''
  if (filters.privacy_mode) {
    if (filters.privacy_mode === ACCOUNT_PRIVACY_MODE_UNSET_QUERY_VALUE) {
      if (privacyMode.trim() !== '') return false
    } else if (privacyMode !== filters.privacy_mode) {
      return false
    }
  }
  const search = String(filters.search || '').trim().toLowerCase()
  if (search && !account.name.toLowerCase().includes(search)) return false
  return true
}
const mergeRuntimeFields = (oldAccount: Account, updatedAccount: Account): Account => ({
  ...updatedAccount,
  current_concurrency: updatedAccount.current_concurrency ?? oldAccount.current_concurrency,
  current_window_cost: updatedAccount.current_window_cost ?? oldAccount.current_window_cost,
  active_sessions: updatedAccount.active_sessions ?? oldAccount.active_sessions
})

const syncPaginationAfterLocalRemoval = () => {
  const nextTotal = Math.max(0, pagination.total - 1)
  pagination.total = nextTotal
  pagination.pages = nextTotal > 0 ? Math.ceil(nextTotal / pagination.page_size) : 0

  const maxPage = Math.max(1, pagination.pages || 1)

  if (pagination.page > maxPage) {
    pagination.page = maxPage
  }
  // 行被本地移除后不立刻全量补页，改为提示用户手动同步。
  hasPendingListSync.value = nextTotal > 0
}

const patchAccountInList = (updatedAccount: Account) => {
  const index = accounts.value.findIndex(account => account.id === updatedAccount.id)
  if (index === -1) return
  const mergedAccount = mergeRuntimeFields(accounts.value[index], updatedAccount)
  if (!accountMatchesCurrentFilters(mergedAccount)) {
    accounts.value = accounts.value.filter(account => account.id !== mergedAccount.id)
    syncPaginationAfterLocalRemoval()
    if (menu.acc?.id === mergedAccount.id) {
      menu.show = false
      menu.acc = null
    }
    return
  }
  const nextAccounts = [...accounts.value]
  nextAccounts[index] = mergedAccount
  accounts.value = nextAccounts
  syncAccountRefs(mergedAccount)
}

const patchUpstreamBillingSnapshot = (accountID: number, snapshot: UpstreamBillingProbeSnapshot) => {
  const account = accounts.value.find(item => item.id === accountID)
  if (!account) return
  patchAccountInList({
    ...account,
    extra: { ...account.extra, upstream_billing_probe: snapshot }
  })
}

const handleProbeUpstreamBilling = async (account: Account) => {
  if (probingUpstreamBilling.has(account.id)) return
  probingUpstreamBilling.add(account.id)
  try {
    const result = await adminAPI.accounts.probeUpstreamBilling(account.id)
    if (result.snapshot) patchUpstreamBillingSnapshot(account.id, result.snapshot)
  } catch (error) {
    console.error('Failed to probe upstream billing:', error)
    appStore.showError(extractApiErrorMessage(error, t('admin.accounts.upstreamBilling.probeFailed')))
  } finally {
    probingUpstreamBilling.delete(account.id)
  }
}

const handleBulkProbeUpstreamBilling = async () => {
  const accountIDs = accounts.value
    .filter(account => account.platform === 'openai' && account.type === 'apikey')
    .map(account => account.id)

  if (accountIDs.length === 0) {
    appStore.showError(t('admin.accounts.upstreamBilling.noEligibleAccounts'))
    return
  }
  if (accountIDs.length > 20) {
    appStore.showError(t('admin.accounts.upstreamBilling.batchLimit'))
    return
  }

  accountIDs.forEach(id => probingUpstreamBilling.add(id))
  try {
    const results = await adminAPI.accounts.probeUpstreamBillingBatch(accountIDs)
    results.forEach(result => {
      if (result.snapshot) patchUpstreamBillingSnapshot(result.account_id, result.snapshot)
    })
    const failed = results.filter(result => result.error).length
    if (failed > 0) {
      appStore.showError(t('admin.accounts.upstreamBilling.batchPartial', {
        success: results.length - failed,
        failed
      }))
    } else {
      appStore.showSuccess(t('admin.accounts.upstreamBilling.batchCompleted', { count: results.length }))
    }
  } catch (error) {
    console.error('Failed to probe upstream billing in batch:', error)
    appStore.showError(extractApiErrorMessage(error, t('admin.accounts.upstreamBilling.probeFailed')))
  } finally {
    accountIDs.forEach(id => probingUpstreamBilling.delete(id))
  }
}

const handleAccountUpdated = (updatedAccount: Account) => {
  patchAccountInList(updatedAccount)
  enterAutoRefreshSilentWindow()
}
const formatExportTimestamp = () => {
  const now = new Date()
  const pad2 = (value: number) => String(value).padStart(2, '0')
  return `${now.getFullYear()}${pad2(now.getMonth() + 1)}${pad2(now.getDate())}${pad2(now.getHours())}${pad2(now.getMinutes())}${pad2(now.getSeconds())}`
}
const openExportDataDialog = () => {
  includeProxyOnExport.value = true
  showExportDataDialog.value = true
}
const handleExportData = async () => {
  if (exportingData.value) return
  exportingData.value = true
  try {
    const dataPayload = await accountExportStepUp.run(() => adminAPI.accounts.exportData({
      includeProxies: includeProxyOnExport.value,
      filters: buildAccountQueryFilters()
    }))
    const timestamp = formatExportTimestamp()
    const filename = `sub2api-account-${timestamp}.json`
    const blob = new Blob([JSON.stringify(dataPayload, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
    if (dataPayload.skipped_shadows && dataPayload.skipped_shadows > 0) {
      appStore.showWarning(t('admin.accounts.dataExportedSkippedShadows', { count: dataPayload.skipped_shadows }))
    } else {
      appStore.showSuccess(t('admin.accounts.dataExported'))
    }
  } catch (error: any) {
    if (isStepUpCancelled(error)) {
      return
    }
    if (isStepUpBlocked(error)) {
      appStore.showError(
        stepUpBlockReason(error) === 'STEP_UP_ADMIN_API_KEY_FORBIDDEN'
          ? t('stepUp.adminApiKeyForbidden')
          : t('stepUp.notEnabled')
      )
      return
    }
    appStore.showError(error?.message || t('admin.accounts.dataExportFailed'))
  } finally {
    exportingData.value = false
    showExportDataDialog.value = false
  }
}
const closeTestModal = () => { showTest.value = false; testingAcc.value = null }
const closeStatsModal = () => { showStats.value = false; statsAcc.value = null }
const closeReAuthModal = () => { showReAuth.value = false; reAuthAcc.value = null }
const handleTest = (a: Account) => { testingAcc.value = a; showTest.value = true }
const handleViewStats = (a: Account) => { statsAcc.value = a; showStats.value = true }
const handleSchedule = async (a: Account) => {
  scheduleAcc.value = a
  scheduleModelOptions.value = []
  showSchedulePanel.value = true
  try {
    const models = await adminAPI.accounts.getAvailableModels(a.id)
    scheduleModelOptions.value = models.map((m: ClaudeModel) => ({ value: m.id, label: m.display_name || m.id }))
  } catch {
    scheduleModelOptions.value = []
  }
}
const closeSchedulePanel = () => { showSchedulePanel.value = false; scheduleAcc.value = null; scheduleModelOptions.value = [] }
const handleReAuth = (a: Account) => { reAuthAcc.value = a; showReAuth.value = true }
const duplicatingAccountIDs = new Set<number>()
const handleDuplicateAccount = async (account: Account) => {
  if (duplicatingAccountIDs.has(account.id)) return
  duplicatingAccountIDs.add(account.id)
  try {
    const duplicateAccount = await adminAPI.accounts.duplicate(account.id)
    appStore.showSuccess(t('admin.accounts.duplicateSuccess', { name: duplicateAccount.name }))
    await reload()
  } catch (error: any) {
    console.error('Failed to duplicate account:', error)
    appStore.showError(error?.message || t('admin.accounts.duplicateFailed'))
  } finally {
    duplicatingAccountIDs.delete(account.id)
  }
}
const handleRefresh = async (a: Account) => {
  try {
    const updated = await adminAPI.accounts.refreshCredentials(a.id)
    patchAccountInList(updated)
    enterAutoRefreshSilentWindow()
  } catch (error) {
    console.error('Failed to refresh credentials:', error)
  }
}
const handleRecoverState = async (a: Account) => {
  try {
    const updated = await adminAPI.accounts.recoverState(a.id)
    patchAccountInList(updated)
    enterAutoRefreshSilentWindow()
    appStore.showSuccess(t('admin.accounts.recoverStateSuccess'))
  } catch (error: any) {
    console.error('Failed to recover account state:', error)
    appStore.showError(error?.message || t('admin.accounts.recoverStateFailed'))
  }
}
const handleResetQuota = async (a: Account) => {
  try {
    const updated = await adminAPI.accounts.resetAccountQuota(a.id)
    patchAccountInList(updated)
    enterAutoRefreshSilentWindow()
    appStore.showSuccess(t('common.success'))
  } catch (error) {
    console.error('Failed to reset quota:', error)
  }
}
const handleSetPrivacy = async (a: Account) => {
  try {
    const updated = await adminAPI.accounts.setPrivacy(a.id)
    patchAccountInList(updated)
    enterAutoRefreshSilentWindow()
    appStore.showSuccess(t('common.success'))
  } catch (error: any) {
    console.error('Failed to set privacy:', error)
    appStore.showError(error?.response?.data?.message || t('admin.accounts.privacyFailed'))
  }
}
const onRevertFallback = async (a: Account) => {
  try {
    await adminAPI.accounts.revertProxyFallback(a.id)
    appStore.showSuccess(t('admin.accounts.revertProxySuccess'))
    reload()
  } catch (error: any) {
    console.error('Failed to revert proxy fallback:', error)
    appStore.showError(error?.response?.data?.message || t('admin.accounts.revertProxyFailed'))
  }
}
const handleCreateSparkShadow = (a: Account) => {
  creatingShadowAcc.value = a
  showCreateShadowDialog.value = true
}
const confirmCreateSparkShadow = async () => {
  const account = creatingShadowAcc.value
  if (!account) return
  try {
    await adminAPI.accounts.createSparkShadow(account.id, { name: `${account.name} (Spark)` })
    showCreateShadowDialog.value = false
    creatingShadowAcc.value = null
    appStore.showSuccess(t('admin.accounts.createSparkShadowSuccess'))
    reload()
  } catch (error: any) {
    console.error('Failed to create spark shadow:', error)
    appStore.showError(error?.response?.data?.message || t('admin.accounts.createSparkShadowFailed'))
  }
}
const handleDelete = (a: Account) => { deletingAcc.value = a; showDeleteDialog.value = true }
const confirmDelete = async () => { if(!deletingAcc.value) return; try { await adminAPI.accounts.delete(deletingAcc.value.id); showDeleteDialog.value = false; deletingAcc.value = null; reload() } catch (error) { console.error('Failed to delete account:', error) } }
const handleToggleSchedulable = async (a: Account) => {
  const nextSchedulable = !a.schedulable
  togglingSchedulable.value = a.id
  try {
    const updated = await adminAPI.accounts.setSchedulable(a.id, nextSchedulable)
    updateSchedulableInList([a.id], updated?.schedulable ?? nextSchedulable)
    enterAutoRefreshSilentWindow()
  } catch (error) {
    console.error('Failed to toggle schedulable:', error)
    appStore.showError(t('admin.accounts.failedToToggleSchedulable'))
  } finally {
    togglingSchedulable.value = null
  }
}
const handleShowTempUnsched = (a: Account) => { tempUnschedAcc.value = a; showTempUnsched.value = true }
const handleTempUnschedReset = async (updated: Account) => {
  showTempUnsched.value = false
  tempUnschedAcc.value = null
  patchAccountInList(updated)
  enterAutoRefreshSilentWindow()
}
const formatExpiresAt = (value: number | null) => {
  if (!value) return '-'
  return formatDateTime(
    new Date(value * 1000),
    {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    },
    'sv-SE'
  )
}
const isExpired = (value: number | null) => {
  if (!value) return false
  return value * 1000 <= Date.now()
}
const proxyExpiryBadge = (p: AccountProxy): string => proxyExpiryBadgeClass(p.expires_at, p.status)
const proxyExpiryText = (p: AccountProxy): string => {
  const { key, params } = proxyExpiryLabelKey(p.expires_at, p.status)
  return params ? t(key, params) : t(key)
}

const handleScroll = () => {
  menu.show = false
  hideAccountStatusTooltip()
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (accountToolsDropdownRef.value && !accountToolsDropdownRef.value.contains(target)) {
    showAccountToolsDropdown.value = false
  }
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(target)) {
    showColumnDropdown.value = false
  }
}

onMounted(async () => {
  load()
  loadUpstreamBillingProbeSettings()
  try {
    const [p, g] = await Promise.all([adminAPI.proxies.getAll(), adminAPI.groups.getAll()])
    proxies.value = p
    groups.value = g
  } catch (error) {
    console.error('Failed to load proxies/groups:', error)
  }
  window.addEventListener('scroll', handleScroll, true)
  document.addEventListener('click', handleClickOutside)

  if (autoRefreshEnabled.value) {
    autoRefreshCountdown.value = autoRefreshIntervalSeconds.value
    resumeAutoRefresh()
  } else {
    pauseAutoRefresh()
  }
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll, true)
  document.removeEventListener('click', handleClickOutside)
  if (usageRefreshedSyncTimer) {
    clearTimeout(usageRefreshedSyncTimer)
    usageRefreshedSyncTimer = null
  }
  hideAccountStatusTooltip()
})
</script>

<style scoped>
.account-tools-menu-item {
  @apply flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-left text-sm text-gray-600 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white;
}

.account-filter-card :deep(.account-filter-row) {
  @apply grid gap-3 md:grid-cols-[14rem_8.5rem_8.5rem_11rem_auto];
}

.account-filter-card :deep(.account-filter-search),
.account-filter-card :deep(.account-filter-select),
.account-filter-card :deep(.account-filter-group) {
  @apply min-w-0 w-full;
}

.account-filter-card :deep(.account-filter-reset) {
  @apply justify-self-start md:justify-self-end;
}

.account-table-scroll {
  @apply overflow-x-auto;
}

.account-table {
  @apply min-w-[980px] w-full divide-y divide-gray-200/70 dark:divide-white/[0.06];
}

.account-table-header {
  @apply bg-gray-50/80 dark:bg-white/[0.03];
}

.account-table-head-cell {
  @apply px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-400;
}

.account-table-row {
  @apply bg-white transition-colors hover:bg-gray-50/80 dark:bg-dark-900 dark:hover:bg-white/[0.035];
}

.account-table-row + .account-table-row .account-table-cell {
  border-top: 1px solid rgb(243 244 246);
}

.dark .account-table-row + .account-table-row .account-table-cell {
  border-top-color: rgb(30 41 59);
}

.account-table-cell {
  @apply px-4 py-4 text-left align-middle text-sm text-gray-700 dark:text-dark-200;
}

.account-table-cell:first-child {
  @apply border-l-[3px] border-l-transparent;
}

.account-table-account-cell {
  @apply min-w-[340px] max-w-[460px];
}

.account-table-usage-cell {
  @apply min-w-[220px];
}

.account-table-muted {
  @apply text-xs text-gray-400 dark:text-dark-500;
}

.account-table-metric-value {
  @apply whitespace-nowrap text-sm font-semibold tabular-nums text-gray-950 dark:text-white;
}

.account-table-today-stat {
  @apply relative inline-block min-w-[128px] text-left;
}

.account-table-today-stat:hover .account-metric-detail-popover,
.account-table-today-stat:focus-within .account-metric-detail-popover {
  @apply block;
}

.account-table-actions {
  @apply flex items-center justify-start gap-2;
}

.account-table-empty-cell {
  @apply px-4 py-0;
}

.account-table-row.account-rich-row-active .account-table-cell:first-child {
  @apply border-l-emerald-500;
}

.account-table-row.account-rich-row-warning .account-table-cell:first-child {
  @apply border-l-amber-400 dark:border-l-amber-500;
}

.account-table-row.account-rich-row-error .account-table-cell:first-child {
  @apply border-l-rose-400 dark:border-l-rose-500;
}

.account-table-row.account-rich-row-muted .account-table-cell:first-child {
  @apply border-l-gray-300 dark:border-l-gray-600;
}

.account-rich-row-skeleton {
  @apply animate-pulse;
}

.account-empty-state {
  @apply flex flex-col items-center justify-center gap-2 px-4 py-12 text-center;
}

.account-empty-icon {
  @apply mb-1 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 text-gray-400 dark:bg-white/[0.06] dark:text-gray-500;
}

.account-card-main {
  @apply min-w-0 flex-1 space-y-1.5;
}

.account-card-title-row {
  @apply flex min-w-0 flex-wrap items-center gap-2;
}

.account-card-title {
  @apply min-w-0 max-w-full truncate text-sm font-semibold text-gray-950 dark:text-white md:max-w-[260px];
}

.account-tiny-badge {
  @apply inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium;
}

.account-compact-badge {
  @apply inline-flex items-center gap-1.5 pl-0.5 text-[11px] font-medium leading-4;
}

.account-card-subline {
  @apply flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500 dark:text-gray-400;
}

.account-card-badges {
  @apply flex min-w-0 flex-wrap items-center gap-1.5 text-xs;
}

.account-status-chip {
  @apply relative inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium outline-none;
  @apply focus:ring-2 focus:ring-primary-500 focus:ring-offset-1 dark:focus:ring-offset-dark-900;
}

.account-status-chip-success {
  @apply bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300;
}

.account-status-chip-warning {
  @apply bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300;
}

.account-status-chip-error {
  @apply bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-300;
}

.account-status-chip-muted {
  @apply bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300;
}

.account-status-tooltip {
  @apply pointer-events-none fixed z-[99999] max-w-[min(22.5rem,calc(100vw-1.5rem))] rounded-md bg-gray-900 px-3 py-2 text-left text-xs font-normal leading-relaxed text-white shadow-xl ring-1 ring-white/10 dark:bg-gray-700;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
  width: max-content;
}

.account-status-tooltip-top {
  transform: translateY(-100%);
}

.account-status-tooltip-bottom {
  transform: translateY(0);
}

.account-status-tooltip-arrow {
  @apply absolute h-2 w-2 -translate-x-1/2 rotate-45 bg-gray-900 dark:bg-gray-700;
}

.account-status-tooltip-arrow-top {
  @apply -bottom-1;
}

.account-status-tooltip-arrow-bottom {
  @apply -top-1;
}

.account-inline-meta {
  @apply inline-flex min-w-0 items-center gap-1 text-xs text-gray-500 dark:text-gray-400;
}

.account-state-badge {
  @apply inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium;
}

.account-state-badge-warning {
  @apply bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300;
}

.account-state-badge-success {
  @apply bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300;
}

.account-metric-subvalue {
  @apply mt-0.5 truncate text-xs font-mono text-gray-500 dark:text-gray-400;
}

.account-metric-subvalue-muted {
  @apply font-sans text-[11px] text-gray-400 dark:text-gray-500;
}

.account-metric-detail-popover {
  @apply pointer-events-none absolute bottom-full left-1/2 z-50 mb-2 hidden w-max max-w-[220px] -translate-x-1/2 whitespace-normal rounded-md bg-gray-900 px-3 py-2 text-center text-xs font-normal leading-relaxed text-white shadow-xl dark:bg-gray-700;
}

.account-list-switch {
  @apply relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out;
  @apply focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:focus:ring-offset-dark-800;
}

.account-list-switch-on {
  @apply bg-emerald-500 hover:bg-emerald-600;
}

.account-list-switch-off {
  @apply bg-gray-200 hover:bg-gray-300 dark:bg-dark-600 dark:hover:bg-dark-500;
}

.account-list-switch-thumb {
  @apply pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out;
}

.account-row-action {
  @apply inline-flex min-w-[52px] flex-col items-center justify-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium leading-none text-gray-500 transition-colors;
  @apply hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white;
}

.account-row-action-label {
  @apply whitespace-nowrap text-[11px] leading-none;
}

.account-row-action-danger {
  @apply hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400;
}

</style>
