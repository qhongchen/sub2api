<template>
  <AppLayout>
    <div class="users-page-layout space-y-5">
      <section class="flex flex-wrap items-start justify-between gap-4">
        <div class="min-w-0">
          <h1 class="text-2xl font-semibold text-gray-950 dark:text-white">
            {{ t('admin.users.title') }}
          </h1>
          <p class="mt-1 max-w-2xl text-sm text-gray-500 dark:text-dark-300">
            {{ t('admin.users.description') }}
          </p>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <button
            v-if="selectedCount > 0"
            type="button"
            class="btn btn-secondary"
            @click="showBulkEditModal = true"
          >
            <Icon name="users" size="md" class="mr-2" />
            {{ text(`批量设置限制（${selectedCount}）`, `Set limits (${selectedCount})`) }}
          </button>
          <button @click="showCreateModal = true" class="btn btn-primary">
            <Icon name="plus" size="md" class="mr-2" />
            {{ t('admin.users.createUser') }}
          </button>
          <button
            @click="showAttributesModal = true"
            class="btn btn-secondary px-2 md:px-3"
            :aria-label="t('admin.users.attributes.configButton')"
            :title="t('admin.users.attributes.configButton')"
          >
            <Icon name="cog" size="md" />
          </button>
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
            <div class="relative" ref="columnDropdownRef">
              <button
                type="button"
                class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-800 shadow-sm transition-colors hover:bg-gray-50 dark:border-white/[0.08] dark:bg-dark-900 dark:text-dark-100 dark:hover:bg-white/[0.04]"
                :aria-label="t('admin.requestLogs.columnSettings')"
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
                  :disabled="isForcedVisibleColumn(col.key)"
                  :class="[
                    'flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-left text-sm transition-colors',
                    isForcedVisibleColumn(col.key)
                      ? 'cursor-not-allowed text-gray-400 dark:text-gray-500'
                      : 'text-gray-600 hover:bg-gray-50 hover:text-gray-950 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white'
                  ]"
                  :title="isForcedVisibleColumn(col.key) ? t('admin.users.columnAlwaysVisible') : ''"
                  @click="toggleColumn(col.key)"
                >
                  <span class="truncate">{{ col.label }}</span>
                  <Icon
                    v-if="isColumnVisible(col.key)"
                    name="check"
                    size="sm"
                    :class="isForcedVisibleColumn(col.key) ? 'text-gray-400 dark:text-gray-500' : 'shrink-0 text-orange-500'"
                    :stroke-width="2"
                  />
                </button>
                <div class="my-1 border-t border-gray-100 dark:border-white/[0.06]" />
                <button
                  type="button"
                  class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-950 dark:text-dark-400 dark:hover:bg-dark-700 dark:hover:text-white"
                  @click="resetColumnVisibility"
                >
                  <Icon name="refresh" size="xs" />
                  {{ t('admin.requestLogs.resetColumns') }}
                </button>
              </div>
            </div>

            <button
              @click="loadUsers"
              :disabled="loading"
              class="consumer-icon-btn"
              :aria-label="t('common.refresh')"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
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
          <div v-if="filtersOpen" class="users-filter-card cch-panel-card p-4 md:p-5">
            <div class="users-filter-row">
              <div class="relative users-filter-search">
                <Icon
                  name="search"
                  size="md"
                  class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                />
                <input
                  v-model="searchQuery"
                  type="text"
                  :placeholder="t('admin.users.searchUsers')"
                  class="input pl-10"
                  @keyup.enter="applyFilter"
                />
              </div>

              <Select
                v-model="filters.role"
                class="users-filter-select"
                :options="[
                  { value: '', label: t('admin.users.allRoles') },
                  { value: 'admin', label: t('admin.users.admin') },
                  { value: 'user', label: t('admin.users.user') }
                ]"
              />

              <Select
                v-model="filters.status"
                class="users-filter-select"
                :options="[
                  { value: '', label: t('admin.users.allStatus') },
                  { value: 'active', label: t('common.active') },
                  { value: 'disabled', label: t('admin.users.disabled') }
                ]"
              />

              <Select
                v-model="filters.group"
                class="users-filter-group"
                :options="groupFilterOptions"
                searchable
                creatable
                :creatable-prefix="t('admin.users.fuzzySearch')"
                :search-placeholder="t('admin.users.searchAuthorizedGroups')"
              />

              <Select
                v-model="filters.apiKeyGroup"
                class="users-filter-api-key-group"
                :options="apiKeyGroupFilterOptions"
                searchable
                :search-placeholder="t('admin.users.searchApiKeyGroups')"
              />

              <template v-for="attr in filterableAttributes" :key="attr.id">
                <div class="relative users-filter-attribute">
                  <input
                    v-if="['text', 'textarea', 'email', 'url', 'date'].includes(attr.type || 'text')"
                    :value="activeAttributeFilters[attr.id] ?? ''"
                    :placeholder="attr.name"
                    class="input w-full"
                    @input="(e) => updateAttributeFilter(attr.id, (e.target as HTMLInputElement).value)"
                    @keyup.enter="applyFilter"
                  />
                  <input
                    v-else-if="attr.type === 'number'"
                    :value="activeAttributeFilters[attr.id] ?? ''"
                    type="number"
                    :placeholder="attr.name"
                    class="input w-full"
                    @input="(e) => updateAttributeFilter(attr.id, (e.target as HTMLInputElement).value)"
                    @keyup.enter="applyFilter"
                  />
                  <Select
                    v-else-if="['select', 'multi_select'].includes(attr.type || '')"
                    :model-value="activeAttributeFilters[attr.id] ?? ''"
                    :options="[
                      { value: '', label: attr.name },
                      ...(attr.options || [])
                    ]"
                    @update:model-value="(val) => updateAttributeFilter(attr.id, String(val ?? ''))"
                  />
                  <input
                    v-else
                    :value="activeAttributeFilters[attr.id] ?? ''"
                    :placeholder="attr.name"
                    class="input w-full"
                    @input="(e) => updateAttributeFilter(attr.id, (e.target as HTMLInputElement).value)"
                    @keyup.enter="applyFilter"
                  />
                </div>
              </template>

              <div class="users-filter-actions">
                <button type="button" class="btn btn-primary" @click="applyFilter">
                  {{ t('usage.applyFilters') }}
                </button>
                <button
                  type="button"
                  class="btn btn-secondary"
                  :disabled="activeFilterCount === 0"
                  @click="resetFilters"
                >
                  {{ t('common.reset') }}
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </section>

      <TablePanel
        :summary="userTableSummary"
        :summary-items="userTableSummaryItems"
        :show-footer="users.length > 0 || loadingMore || hasMoreUsers"
      >
        <div class="users-table-card">
          <DataTable
            :columns="columns"
            :data="users"
            :loading="loading"
            :actions-count="7"
            :virtualized="false"
            :sticky-first-column="false"
            :sticky-actions-column="false"
            wrapper-class="hidden overflow-x-auto md:block"
            table-class="w-full min-w-[980px] divide-y divide-gray-200/70 dark:divide-white/[0.06]"
            header-class="bg-gray-50/80 dark:bg-white/[0.03]"
            header-cell-base-class="py-3 text-left text-xs font-medium text-gray-500 dark:text-dark-400"
            cell-base-class="py-4 text-left"
            padding-class="px-4"
            row-class="transition-colors hover:bg-orange-50/40 dark:hover:bg-white/[0.035]"
            mobile-list-class="divide-y divide-gray-100 md:hidden dark:divide-white/[0.06]"
            mobile-card-class="space-y-3 p-4"
            mobile-empty-class="flex min-h-[220px] flex-col items-center justify-center gap-3 p-6 text-center"
          >
            <template #header-select>
              <input
                type="checkbox"
                class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                :checked="allVisibleSelected"
                :disabled="users.length === 0"
                :aria-label="text('选择当前列表中的全部用户', 'Select all visible users')"
                @change="toggleVisible(($event.target as HTMLInputElement).checked)"
              />
            </template>
            <template #cell-select="{ row }">
              <input
                type="checkbox"
                class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                :checked="isSelected(row.id)"
                :aria-label="text(`选择 ${row.email}`, `Select ${row.email}`)"
                @click.stop
                @change="toggle(row.id)"
              />
            </template>
            <template #cell-email="{ value, row }">
            <div class="flex items-center gap-2">
              <div
                class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30"
              >
                <span class="text-sm font-medium text-primary-700 dark:text-primary-300">
                  {{ value.charAt(0).toUpperCase() }}
                </span>
              </div>
              <div class="min-w-0">
                <div class="truncate font-medium text-gray-900 dark:text-white">{{ value }}</div>
                <div class="mt-0.5 truncate text-xs text-gray-500 dark:text-dark-400">
                  {{ row.username || '-' }}
                </div>
              </div>
            </div>
          </template>

          <template #cell-notes="{ value }">
            <div class="max-w-xs">
              <span
                v-if="value"
                :title="value.length > 30 ? value : undefined"
                class="block truncate text-sm text-gray-600 dark:text-gray-400"
              >
                {{ value.length > 30 ? value.substring(0, 25) + '...' : value }}
              </span>
              <span v-else class="text-sm text-gray-400">-</span>
            </div>
          </template>

          <!-- Dynamic attribute columns -->
          <template
            v-for="def in attributeDefinitions.filter(d => d.enabled)"
            :key="def.id"
            #[`cell-attr_${def.id}`]="{ row }"
          >
            <div class="max-w-xs">
              <span
                class="block truncate text-sm text-gray-700 dark:text-gray-300"
                :title="getAttributeValue(row.id, def.id)"
              >
                {{ getAttributeValue(row.id, def.id) }}
              </span>
            </div>
          </template>

          <template #cell-role="{ value }">
            <span :class="['badge', value === 'admin' ? 'badge-purple' : 'badge-gray']">
              {{ t('admin.users.roles.' + value) }}
            </span>
          </template>

          <template #cell-groups="{ row }">
            <div v-if="allGroups.length > 0" class="flex flex-col gap-1">
              <!-- 专属分组行 -->
              <span
                v-if="getUserGroups(row).exclusive.length > 0"
                class="group/ex relative inline-flex cursor-pointer items-center gap-1 whitespace-nowrap text-xs"
                @click.stop="toggleExpandedGroup(row.id)"
              >
                <Icon name="shield" size="xs" class="h-3.5 w-3.5 text-purple-500 dark:text-purple-400" />
                <span class="font-medium text-purple-600 dark:text-purple-400">{{ getUserGroups(row).exclusive.length }}</span>
                <span class="text-gray-500 dark:text-dark-400">{{ t('admin.users.exclusiveLabel') }}</span>
                <!-- Hover tooltip（操作菜单未打开时显示） -->
                <div
                  v-if="expandedGroupUserId !== row.id"
                  class="pointer-events-none absolute left-0 top-full z-50 mt-1.5 rounded bg-gray-900 px-2.5 py-1.5 text-xs text-white opacity-0 shadow-lg transition-opacity duration-75 group-hover/ex:opacity-100 dark:bg-dark-600"
                >
                  <div class="absolute left-4 bottom-full border-4 border-transparent border-b-gray-900 dark:border-b-dark-600"></div>
                  <div class="flex flex-col gap-0.5 whitespace-nowrap">
                    <span v-for="g in getUserGroups(row).exclusive" :key="g.id">{{ g.name }}</span>
                  </div>
                </div>
                <!-- 点击展开分组操作菜单 -->
                <div
                  v-if="expandedGroupUserId === row.id"
                  class="absolute left-0 top-full z-50 mt-1.5 min-w-[160px] overflow-hidden rounded-lg border border-gray-200 bg-white py-1 text-xs shadow-xl dark:border-dark-600 dark:bg-dark-700"
                >
                  <div class="border-b border-gray-100 px-3 py-1.5 text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:border-dark-600 dark:text-dark-400">
                    {{ t('admin.users.clickToReplace') }}
                  </div>
                  <div
                    v-for="g in getUserGroups(row).exclusive"
                    :key="g.id"
                    class="flex cursor-pointer items-center gap-2 px-3 py-2 text-gray-700 transition-colors hover:bg-primary-50 hover:text-primary-600 dark:text-dark-200 dark:hover:bg-primary-900/30 dark:hover:text-primary-400"
                    @click.stop="openGroupReplace(row, g)"
                  >
                    <Icon name="swap" size="xs" class="h-3.5 w-3.5 flex-shrink-0 opacity-50" />
                    <span class="flex-1">{{ g.name }}</span>
                  </div>
                </div>
              </span>
              <!-- 公开分组行 -->
              <span
                v-if="getUserGroups(row).publicGroups.length > 0"
                class="group/pub relative inline-flex cursor-default items-center gap-1 whitespace-nowrap text-xs"
              >
                <Icon name="globe" size="xs" class="h-3.5 w-3.5 text-gray-400 dark:text-dark-500" />
                <span class="font-medium text-gray-600 dark:text-dark-300">{{ getUserGroups(row).publicGroups.length }}</span>
                <span class="text-gray-400 dark:text-dark-500">{{ t('admin.users.publicLabel') }}</span>
                <!-- Tooltip: 向下弹出 -->
                <div class="pointer-events-none absolute left-0 top-full z-50 mt-1.5 rounded bg-gray-900 px-2.5 py-1.5 text-xs text-white opacity-0 shadow-lg transition-opacity duration-75 group-hover/pub:opacity-100 dark:bg-dark-600">
                  <div class="absolute left-4 bottom-full border-4 border-transparent border-b-gray-900 dark:border-b-dark-600"></div>
                  <div class="flex flex-col gap-0.5 whitespace-nowrap">
                    <span v-for="g in getUserGroups(row).publicGroups" :key="g.id">{{ g.name }}</span>
                  </div>
                </div>
              </span>
              <!-- 都没有 -->
              <span
                v-if="getUserGroups(row).exclusive.length === 0 && getUserGroups(row).publicGroups.length === 0"
                class="text-xs text-gray-400 dark:text-dark-500"
              >-</span>
            </div>
            <span v-else class="text-xs text-gray-400 dark:text-dark-500">-</span>
          </template>

          <template #cell-subscriptions="{ row }">
            <div
              v-if="row.subscriptions && row.subscriptions.length > 0"
              class="flex flex-wrap gap-1.5"
            >
              <GroupBadge
                v-for="sub in row.subscriptions"
                :key="sub.id"
                :name="sub.group?.name || ''"
                :platform="sub.group?.platform"
                :subscription-type="sub.group?.subscription_type"
                :rate-multiplier="sub.group?.rate_multiplier"
                :days-remaining="sub.expires_at ? getDaysRemaining(sub.expires_at) : null"
                :title="sub.expires_at ? formatDateTime(sub.expires_at) : ''"
              />
            </div>
            <span
              v-else
              class="inline-flex items-center gap-1.5 rounded-md bg-gray-50 px-2 py-1 text-xs text-gray-400 dark:bg-dark-700/50 dark:text-dark-500"
            >
              <Icon name="ban" size="xs" class="h-3.5 w-3.5" />
              <span>{{ t('admin.users.noSubscription') }}</span>
            </span>
          </template>

          <template #cell-balance="{ value, row }">
            <div class="flex items-center gap-2">
              <div class="group relative">
                <button
                  class="font-medium text-gray-900 underline decoration-dashed decoration-gray-300 underline-offset-4 transition-colors hover:text-primary-600 dark:text-white dark:decoration-dark-500 dark:hover:text-primary-400"
                  @click="handleBalanceHistory(row)"
                >
                  ${{ value.toFixed(2) }}
                </button>
                <!-- Instant tooltip -->
                <div class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-1.5 -translate-x-1/2 whitespace-nowrap rounded bg-gray-900 px-2 py-1 text-xs text-white opacity-0 shadow-lg transition-opacity duration-75 group-hover:opacity-100 dark:bg-dark-600">
                  {{ t('admin.users.balanceHistoryTip') }}
                  <div class="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-gray-900 dark:border-t-dark-600"></div>
                </div>
              </div>
              <button
                @click.stop="handleDeposit(row)"
                class="rounded px-2 py-0.5 text-xs font-medium text-emerald-600 transition-colors hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-900/20"
                :title="t('admin.users.deposit')"
              >
                {{ t('admin.users.deposit') }}
              </button>
            </div>
          </template>

          <template #cell-balance_platform_quota="{ row }">
            <button
              type="button"
              class="block text-left underline decoration-dashed decoration-gray-300 underline-offset-4 transition-colors hover:decoration-primary-400 dark:decoration-dark-500"
              :title="t('admin.users.platformQuota.cellColumnTooltip')"
              @click="handlePlatformQuota(row)"
            >
              <UserPlatformQuotaCell :quotas="platformQuotaStats[row.id]" />
            </button>
          </template>

          <template #cell-usage="{ row }">
            <PlatformUsageBreakdown
              :today="usageStats[row.id]?.today_actual_cost ?? 0"
              :total="usageStats[row.id]?.total_actual_cost ?? 0"
              :by-platform="usageStats[row.id]?.by_platform"
            />
          </template>

          <template #cell-usage_anthropic="{ row }">
            <PlatformCostCell :usage="getPlatformUsage(row.id, 'anthropic')" />
          </template>

          <template #cell-usage_openai="{ row }">
            <PlatformCostCell :usage="getPlatformUsage(row.id, 'openai')" />
          </template>

          <template #cell-usage_gemini="{ row }">
            <PlatformCostCell :usage="getPlatformUsage(row.id, 'gemini')" />
          </template>

          <template #cell-usage_antigravity="{ row }">
            <PlatformCostCell :usage="getPlatformUsage(row.id, 'antigravity')" />
          </template>

          <template #cell-concurrency="{ row }">
            <UserConcurrencyCell
              :current="row.current_concurrency ?? 0"
              :max="row.concurrency"
            />
          </template>

          <template #cell-status="{ value }">
            <div class="flex items-center gap-1.5">
              <span
                :class="[
                  'inline-block h-2 w-2 rounded-full',
                  value === 'active' ? 'bg-green-500' : 'bg-red-500'
                ]"
              ></span>
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ value === 'active' ? t('common.active') : t('admin.users.disabled') }}
              </span>
            </div>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-last_used_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatDateTime(value) : '-' }}
            </span>
          </template>

          <template #cell-last_active_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatDateTime(value) : '-' }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <button
                type="button"
                @click="handleEdit(row)"
                class="admin-row-action"
                :aria-label="t('common.edit')"
                :title="t('common.edit')"
              >
                <Icon name="edit" size="sm" />
                <span class="admin-row-action-label">{{ t('common.edit') }}</span>
              </button>

              <button
                v-if="row.role !== 'admin'"
                type="button"
                @click="handleToggleStatus(row)"
                class="admin-row-action"
                :aria-label="row.status === 'active' ? t('admin.users.disable') : t('admin.users.enable')"
                :title="row.status === 'active' ? t('admin.users.disable') : t('admin.users.enable')"
              >
                <Icon v-if="row.status === 'active'" name="ban" size="sm" />
                <Icon v-else name="checkCircle" size="sm" />
                <span class="admin-row-action-label">
                  {{ row.status === 'active' ? t('admin.users.disable') : t('admin.users.enable') }}
                </span>
              </button>

              <button
                type="button"
                @click="openActionMenu(row, $event)"
                class="action-menu-trigger admin-row-action"
                :class="{ 'bg-gray-100 text-gray-900 dark:bg-dark-700 dark:text-white': activeMenuId === row.id }"
                :aria-label="t('common.more')"
                :title="t('common.more')"
              >
                <Icon name="more" size="sm" />
                <span class="admin-row-action-label">{{ t('common.more') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.users.noUsersYet')"
              :description="t('admin.users.createFirstUser')"
              :action-text="t('admin.users.createUser')"
              @action="showCreateModal = true"
            />
          </template>
          </DataTable>
        </div>

        <template #footer>
          <div class="flex min-h-10 items-center justify-center text-sm text-gray-500 dark:text-dark-400">
            <span v-if="loadingMore" class="inline-flex items-center gap-2">
              <Icon name="refresh" size="sm" class="animate-spin" />
              {{ t('admin.requestLogs.loadingMore') }}
            </span>
            <span v-else-if="hasMoreUsers">
              {{ t('admin.requestLogs.scrollLoadMore') }}
            </span>
            <span v-else-if="users.length > 0">
              {{ t('admin.requestLogs.allRecordsLoaded') }}
            </span>
          </div>
          <div ref="loadMoreSentinelRef" class="h-px" aria-hidden="true" />
        </template>
      </TablePanel>
    </div>

    <!-- Action Menu (Teleported) -->
    <Teleport to="body">
      <div
        v-if="activeMenuId !== null && menuPosition"
        class="action-menu-content fixed z-[9999] w-48 overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5 dark:bg-dark-800 dark:ring-white/10"
        :style="{ top: menuPosition.top + 'px', left: menuPosition.left + 'px' }"
      >
        <div class="py-1">
          <template v-for="user in users" :key="user.id">
            <template v-if="user.id === activeMenuId">
              <button
                @click="handleViewApiKeys(user); closeActionMenu()"
                class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                {{ t('admin.users.apiKeys') }}
              </button>

              <button
                @click="handleAllowedGroups(user); closeActionMenu()"
                class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                {{ t('admin.users.groups') }}
              </button>

              <button
                @click="handleDeposit(user); closeActionMenu()"
                class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                {{ t('admin.users.deposit') }}
              </button>

              <button
                @click="handleWithdraw(user); closeActionMenu()"
                class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                {{ t('admin.users.withdraw') }}
              </button>

              <button
                @click="handlePlatformQuota(user); closeActionMenu()"
                class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                {{ t('admin.users.platformQuota.menuItem') }}
              </button>

              <button
                @click="handleBalanceHistory(user); closeActionMenu()"
                class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                {{ t('admin.users.balanceHistory') }}
              </button>

              <button
                v-if="user.role !== 'admin'"
                @click="handleDelete(user); closeActionMenu()"
                class="flex w-full items-center px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
              >
                {{ t('common.delete') }}
              </button>
            </template>
          </template>
        </div>
      </div>
    </Teleport>

    <ConfirmDialog :show="showDeleteDialog" :title="t('admin.users.deleteUser')" :message="t('admin.users.deleteConfirm', { email: deletingUser?.email })" :danger="true" @confirm="confirmDelete" @cancel="showDeleteDialog = false" />
    <UserCreateModal :show="showCreateModal" @close="showCreateModal = false" @success="loadUsers" />
    <UserEditModal :show="showEditModal" :user="editingUser" @close="closeEditModal" @success="loadUsers" />
    <BulkEditUserModal
      :show="showBulkEditModal"
      :selected-ids="selectedIds"
      @close="showBulkEditModal = false"
      @success="handleBulkLimitsSuccess"
    />
    <UserPlatformQuotaModal
      :show="showPlatformQuotaModal"
      :user="platformQuotaUser"
      @close="closePlatformQuotaModal"
      @success="loadUsers"
    />
    <UserApiKeysModal :show="showApiKeysModal" :user="viewingUser" @close="closeApiKeysModal" />
    <UserAllowedGroupsModal :show="showAllowedGroupsModal" :user="allowedGroupsUser" @close="closeAllowedGroupsModal" @success="loadUsers" />
    <UserBalanceModal :show="showBalanceModal" :user="balanceUser" :operation="balanceOperation" @close="closeBalanceModal" @success="loadUsers" />
    <UserBalanceHistoryModal :show="showBalanceHistoryModal" :user="balanceHistoryUser" @close="closeBalanceHistoryModal" @deposit="handleDepositFromHistory" @withdraw="handleWithdrawFromHistory" />
    <GroupReplaceModal :show="showGroupReplaceModal" :user="groupReplaceUser" :old-group="groupReplaceOldGroup" :all-groups="allGroups" @close="closeGroupReplaceModal" @success="loadUsers" />
    <UserAttributesConfigModal :show="showAttributesModal" @close="handleAttributesModalClose" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { useInfinitePagedList } from '@/composables/useInfinitePagedList'
import { useTableSelection } from '@/composables/useTableSelection'
import { formatDateTime, formatNumber } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'

const { t, locale } = useI18n()
import { adminAPI } from '@/api/admin'
import type { AdminUser, AdminGroup, PaginatedResponse, UserAttributeDefinition } from '@/types'
import type { BatchUserUsageStats } from '@/api/admin/dashboard'
import type { PlatformQuotaItem } from '@/api/admin/users'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import TablePanel from '@/components/common/TablePanel.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import { buildApiKeyGroupFilterOptions } from './apiKeyGroupFilterOptions'
import UserAttributesConfigModal from '@/components/user/UserAttributesConfigModal.vue'
import UserConcurrencyCell from '@/components/user/UserConcurrencyCell.vue'
import PlatformUsageBreakdown from '@/components/user/PlatformUsageBreakdown.vue'
import PlatformCostCell from '@/components/user/PlatformCostCell.vue'
import UserPlatformQuotaCell from '@/components/user/UserPlatformQuotaCell.vue'
import UserCreateModal from '@/components/admin/user/UserCreateModal.vue'
import UserEditModal from '@/components/admin/user/UserEditModal.vue'
import BulkEditUserModal from '@/components/admin/user/BulkEditUserModal.vue'
import UserPlatformQuotaModal from '@/components/admin/user/UserPlatformQuotaModal.vue'
import UserApiKeysModal from '@/components/admin/user/UserApiKeysModal.vue'
import UserAllowedGroupsModal from '@/components/admin/user/UserAllowedGroupsModal.vue'
import UserBalanceModal from '@/components/admin/user/UserBalanceModal.vue'
import UserBalanceHistoryModal from '@/components/admin/user/UserBalanceHistoryModal.vue'
import GroupReplaceModal from '@/components/admin/user/GroupReplaceModal.vue'

const appStore = useAppStore()
const text = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en

// Generate dynamic attribute columns from enabled definitions
const attributeColumns = computed<Column[]>(() =>
  attributeDefinitions.value
    .filter(def => def.enabled)
    .map(def => ({
      key: `attr_${def.id}`,
      label: def.name
    }))
)

// Get formatted attribute value for display in table
const getAttributeValue = (userId: number, attrId: number): string => {
  const userAttrs = userAttributeValues.value[userId]
  if (!userAttrs) return '-'
  const value = userAttrs[attrId]
  if (!value) return '-'

  // Find definition for this attribute
  const def = attributeDefinitions.value.find(d => d.id === attrId)
  if (!def) return value

  // Format based on type
  if (def.type === 'multi_select' && value) {
    try {
      const arr = JSON.parse(value)
      if (Array.isArray(arr)) {
        // Map values to labels
        return arr.map(v => {
          const opt = def.options?.find(o => o.value === v)
          return opt?.label || v
        }).join(', ')
      }
    } catch {
      return value
    }
  }

  if (def.type === 'select' && value && def.options) {
    const opt = def.options.find(o => o.value === value)
    return opt?.label || value
  }

  return value
}

// All possible columns (for column settings)
const allColumns = computed<Column[]>(() => [
  { key: 'select', label: '' },
  { key: 'email', label: t('admin.users.columns.user') },
  { key: 'id', label: t('admin.users.columns.id') },
  { key: 'notes', label: t('admin.users.columns.notes') },
  // Dynamic attribute columns
  ...attributeColumns.value,
  { key: 'role', label: t('admin.users.columns.role') },
  { key: 'groups', label: t('admin.users.columns.groups') },
  { key: 'subscriptions', label: t('admin.users.columns.subscriptions') },
  { key: 'balance', label: t('admin.users.columns.balance') },
  { key: 'balance_platform_quota', label: t('admin.users.columns.balancePlatformQuota') },
  { key: 'usage', label: t('admin.users.columns.usage') },
  { key: 'usage_anthropic', label: t('admin.users.columns.usageAnthropic') },
  { key: 'usage_openai', label: t('admin.users.columns.usageOpenAI') },
  { key: 'usage_gemini', label: t('admin.users.columns.usageGemini') },
  { key: 'usage_antigravity', label: t('admin.users.columns.usageAntigravity') },
  { key: 'concurrency', label: t('admin.users.columns.concurrency') },
  { key: 'status', label: t('admin.users.columns.status') },
  { key: 'last_active_at', label: t('admin.users.columns.lastActive') },
  { key: 'last_used_at', label: t('admin.users.columns.lastUsed') },
  { key: 'created_at', label: t('admin.users.columns.created') },
  { key: 'actions', label: t('admin.users.columns.actions') }
])

// Columns that can be toggled (exclude email and actions which are always visible)
const toggleableColumns = computed(() =>
  allColumns.value.filter(col => !['select', 'email', 'actions'].includes(col.key))
)

// Hidden columns (stored in Set - columns NOT in this set are visible)
// This way, new columns are visible by default
const hiddenColumns = reactive<Set<string>>(new Set())

// 用户管理页默认只展示高频判断字段，其余列仍可在"列设置"里手动打开。
const DEFAULT_VISIBLE_COLUMNS = new Set([
  'select',
  'email',
  'role',
  'usage',
  'concurrency',
  'status',
  'actions'
])
const getDefaultHiddenColumns = () =>
  allColumns.value
    .map((col) => col.key)
    .filter((key) => !DEFAULT_VISIBLE_COLUMNS.has(key))
const REMOVED_COLUMNS = new Set(['last_login_at'])
// 强制可见列：加载时会被强制移出 hiddenColumns，并在列设置 UI 上 disabled。
// 当前没有列需要强制可见 —— last_active_at 已改为可被用户隐藏。
const FORCED_VISIBLE_COLUMNS = new Set<string>()

// localStorage keys for column settings
const HIDDEN_COLUMNS_KEY = 'user-hidden-columns'
// 列设置 schema 版本号。每次给 DEFAULT_HIDDEN_COLUMNS 新增列时 bump 一次，
// 并在 VERSION_NEW_HIDDEN_COLUMNS 中登记该版本新增的 key。
// 这样老用户升级后这些新列会被自动隐藏一次，而不会影响他们对其它老列的偏好。
const COLUMN_SETTINGS_VERSION_KEY = 'user-column-settings-version'
const COLUMN_SETTINGS_VERSION = 4
const getVersionNewHiddenColumns = (version: number): string[] => {
  if (version === 2) return ['usage_anthropic', 'usage_openai', 'usage_gemini', 'usage_antigravity']
  if (version === 3) return getDefaultHiddenColumns()
  if (version === 4) return ['balance_platform_quota']
  return []
}

// Load saved column settings
const loadSavedColumns = () => {
  try {
    const saved = localStorage.getItem(HIDDEN_COLUMNS_KEY)
    if (saved) {
      const parsed = JSON.parse(saved) as string[]
      parsed
        .filter(key => !REMOVED_COLUMNS.has(key) && !FORCED_VISIBLE_COLUMNS.has(key))
        .forEach(key => hiddenColumns.add(key))

      // 老用户升级：把每个未应用过的版本里新增的默认隐藏列自动追加到 hiddenColumns。
      const storedVersion = Number(localStorage.getItem(COLUMN_SETTINGS_VERSION_KEY) ?? '1')
      if (storedVersion < COLUMN_SETTINGS_VERSION) {
        let mutated = false
        for (let v = storedVersion + 1; v <= COLUMN_SETTINGS_VERSION; v++) {
          for (const key of getVersionNewHiddenColumns(v)) {
            if (REMOVED_COLUMNS.has(key) || FORCED_VISIBLE_COLUMNS.has(key)) continue
            if (!hiddenColumns.has(key)) {
              hiddenColumns.add(key)
              mutated = true
            }
          }
        }
        if (mutated) saveColumnsToStorage()
        else localStorage.setItem(COLUMN_SETTINGS_VERSION_KEY, String(COLUMN_SETTINGS_VERSION))
      }
    } else {
      // Use default hidden columns on first load
      getDefaultHiddenColumns().forEach(key => hiddenColumns.add(key))
      localStorage.setItem(COLUMN_SETTINGS_VERSION_KEY, String(COLUMN_SETTINGS_VERSION))
    }
  } catch (e) {
    console.error('Failed to load saved columns:', e)
    getDefaultHiddenColumns().forEach(key => hiddenColumns.add(key))
  }
}

// Save column settings to localStorage
const saveColumnsToStorage = () => {
  try {
    localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
    localStorage.setItem(COLUMN_SETTINGS_VERSION_KEY, String(COLUMN_SETTINGS_VERSION))
  } catch (e) {
    console.error('Failed to save columns:', e)
  }
}

// Toggle column visibility
const isForcedVisibleColumn = (key: string) => FORCED_VISIBLE_COLUMNS.has(key)
const toggleColumn = (key: string) => {
  // 强制可见列(如 last_active_at)在加载时会被恢复成可见，
  // 这里阻止用户在当前会话隐藏它，避免"取消勾选 → 刷新又恢复"的反直觉行为。
  if (FORCED_VISIBLE_COLUMNS.has(key)) return
  const wasHidden = hiddenColumns.has(key)
  if (hiddenColumns.has(key)) {
    hiddenColumns.delete(key)
  } else {
    hiddenColumns.add(key)
  }
  saveColumnsToStorage()
  if (wasHidden && (key === 'usage' || key.startsWith('usage_') || key.startsWith('attr_') || key === 'balance_platform_quota')) {
    refreshCurrentPageSecondaryData()
  }
  if (key === 'subscriptions') {
    void loadUsers()
  }
  if (wasHidden && key === 'groups') {
    loadAllGroups()
  }
}
const resetColumnVisibility = () => {
  hiddenColumns.clear()
  getDefaultHiddenColumns().forEach(key => hiddenColumns.add(key))
  saveColumnsToStorage()
  void loadUsers()
}

// Check if column is visible (not in hidden set)
const isColumnVisible = (key: string) => !hiddenColumns.has(key)
// usage 主列或任意 usage_<platform> 子列可见时都需要批量拉取用量数据
const PLATFORM_USAGE_COLUMNS = ['usage_anthropic', 'usage_openai', 'usage_gemini', 'usage_antigravity']
const hasVisibleUsageColumn = computed(
  () => !hiddenColumns.has('usage') || PLATFORM_USAGE_COLUMNS.some((k) => !hiddenColumns.has(k))
)
const hasVisiblePlatformQuotaColumn = computed(() => !hiddenColumns.has('balance_platform_quota'))
const hasVisibleAttributeColumns = computed(() =>
  attributeDefinitions.value.some((def) => def.enabled && !hiddenColumns.has(`attr_${def.id}`))
)

// Filtered columns based on visibility
const columns = computed<Column[]>(() =>
  allColumns.value.filter(col =>
    col.key === 'email' || col.key === 'actions' || !hiddenColumns.has(col.key)
  )
)
const visibleColumnCount = computed(() => columns.value.length)

const searchQuery = ref('')
const USER_DEFAULT_SORT = {
  sort_by: 'created_at',
  sort_order: 'desc' as const
}

// Groups data for the groups column
const allGroups = ref<AdminGroup[]>([])
const loadAllGroups = async () => {
  if (allGroups.value.length > 0) return
  try {
    allGroups.value = await adminAPI.groups.getAll()
  } catch (e) {
    console.error('Failed to load groups:', e)
  }
}
const allGroupsForApiKeyFilter = ref<AdminGroup[]>([])
const loadAllGroupsForApiKeyFilter = async () => {
  if (allGroupsForApiKeyFilter.value.length > 0) return
  try {
    allGroupsForApiKeyFilter.value = await adminAPI.groups.getAllIncludingInactive()
  } catch (e) {
    console.error('Failed to load groups for API key filter:', e)
  }
}
// Resolve user's accessible groups: exclusive groups first, then public groups
const getUserGroups = (user: AdminUser) => {
  const exclusive: AdminGroup[] = []
  const publicGroups: AdminGroup[] = []
  for (const g of allGroups.value) {
    if (g.status !== 'active' || g.subscription_type !== 'standard') continue
    if (g.is_exclusive) {
      if (user.allowed_groups?.includes(g.id)) {
        exclusive.push(g)
      }
    } else {
      publicGroups.push(g)
    }
  }
  return { exclusive, publicGroups }
}

// Group filter options: "All Groups" + active exclusive groups (value = group name for fuzzy match)
const groupFilterOptions = computed(() => {
  const options: { value: string; label: string }[] = [
    { value: '', label: t('admin.users.allAuthorizedGroups') }
  ]
  for (const g of allGroups.value) {
    if (g.status !== 'active' || !g.is_exclusive || g.subscription_type !== 'standard') continue
    options.push({ value: g.name, label: g.name })
  }
  return options
})

const apiKeyGroupFilterOptions = computed(() =>
  buildApiKeyGroupFilterOptions(allGroupsForApiKeyFilter.value, {
    all: t('admin.users.allApiKeyGroups'),
    exclusive: t('admin.users.apiKeyGroupExclusive'),
    public: t('admin.users.apiKeyGroupPublic'),
    subscription: t('admin.users.apiKeyGroupSubscription'),
    disabled: t('admin.users.apiKeyGroupDisabled'),
  }) as SelectOption[]
)

// Filter values (role, status, and custom attributes)
const filters = reactive({
  role: '',
  status: '',
  group: '',  // group name for fuzzy match, '' = all
  apiKeyGroup: null as number | null
})
const activeAttributeFilters = reactive<Record<number, string>>({})
const loadMoreSentinelRef = ref<HTMLElement | null>(null)
const filtersOpen = ref(false)
const activeFilterCount = computed(() => {
  let count = 0
  if (searchQuery.value.trim()) count += 1
  if (filters.role) count += 1
  if (filters.status) count += 1
  if (filters.group) count += 1
  if (filters.apiKeyGroup !== null) count += 1
  count += Object.values(activeAttributeFilters).filter(Boolean).length
  return count
})
const userTableSummary = computed(() =>
  t('admin.requestLogs.loadedRecordsSummary', {
    count: formatNumber(users.value.length),
    total: formatNumber(pagination.total)
  })
)
const userTableSummaryItems = computed(() => [
  `${formatNumber(pagination.total)} ${t('nav.users')}`,
  `${formatNumber(users.value.length)} ${t('common.visible')}`
])

// Dropdown states
const showColumnDropdown = ref(false)

// Dropdown refs for click outside detection
const columnDropdownRef = ref<HTMLElement | null>(null)

// localStorage keys
const FILTER_VALUES_KEY = 'user-filter-values'

// All filterable attribute definitions (enabled attributes)
const filterableAttributes = computed(() =>
  attributeDefinitions.value.filter(def => def.enabled)
)

// Load saved filters from localStorage
const loadSavedFilters = () => {
  try {
    // Load filter values
    const savedValues = localStorage.getItem(FILTER_VALUES_KEY)
    if (savedValues) {
      const parsed = JSON.parse(savedValues)
      if (parsed.role) filters.role = parsed.role
      if (parsed.status) filters.status = parsed.status
      if (parsed.group) filters.group = parsed.group
      if (typeof parsed.apiKeyGroup === 'number') filters.apiKeyGroup = parsed.apiKeyGroup
      if (parsed.attributes) {
        Object.assign(activeAttributeFilters, parsed.attributes)
      }
    }
  } catch (e) {
    console.error('Failed to load saved filters:', e)
  }
}

// Save filters to localStorage
const saveFiltersToStorage = () => {
  try {
    // Save filter values
    const values = {
      role: filters.role,
      status: filters.status,
      group: filters.group,
      apiKeyGroup: filters.apiKeyGroup,
      attributes: activeAttributeFilters
    }
    localStorage.setItem(FILTER_VALUES_KEY, JSON.stringify(values))
  } catch (e) {
    console.error('Failed to save filters:', e)
  }
}

const usageStats = ref<Record<string, BatchUserUsageStats>>({})
const platformQuotaStats = ref<Record<string, PlatformQuotaItem[]>>({})

const getPlatformUsage = (userId: number, platform: string) =>
  usageStats.value[userId]?.by_platform?.find((p) => p.platform === platform)

// User attribute definitions and values
const attributeDefinitions = ref<UserAttributeDefinition[]>([])
const userAttributeValues = ref<Record<number, Record<number, string>>>({})

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showBulkEditModal = ref(false)
const showDeleteDialog = ref(false)
const showApiKeysModal = ref(false)
const showAttributesModal = ref(false)
const showPlatformQuotaModal = ref(false)
const editingUser = ref<AdminUser | null>(null)
const deletingUser = ref<AdminUser | null>(null)
const viewingUser = ref<AdminUser | null>(null)
const platformQuotaUser = ref<AdminUser | null>(null)

const handlePlatformQuota = (user: AdminUser) => {
  platformQuotaUser.value = user
  showPlatformQuotaModal.value = true
}

const closePlatformQuotaModal = () => {
  showPlatformQuotaModal.value = false
  platformQuotaUser.value = null
}
let secondaryDataSeq = 0

const loadUsersSecondaryData = async (
  userIds: number[],
  signal?: AbortSignal,
  expectedSeq?: number
) => {
  if (userIds.length === 0) return

  const tasks: Promise<void>[] = []

  if (hasVisibleUsageColumn.value) {
    tasks.push(
      (async () => {
        try {
          const usageResponse = await adminAPI.dashboard.getBatchUsersUsage(userIds)
          if (signal?.aborted) return
          if (typeof expectedSeq === 'number' && expectedSeq !== secondaryDataSeq) return
          usageStats.value = { ...usageStats.value, ...usageResponse.stats }
        } catch (e) {
          if (signal?.aborted) return
          console.error('Failed to load usage stats:', e)
        }
      })()
    )
  }

  if (attributeDefinitions.value.length > 0 && hasVisibleAttributeColumns.value) {
    tasks.push(
      (async () => {
        try {
          const attrResponse = await adminAPI.userAttributes.getBatchUserAttributes(userIds)
          if (signal?.aborted) return
          if (typeof expectedSeq === 'number' && expectedSeq !== secondaryDataSeq) return
          userAttributeValues.value = { ...userAttributeValues.value, ...attrResponse.attributes }
        } catch (e) {
          if (signal?.aborted) return
          console.error('Failed to load user attribute values:', e)
        }
      })()
    )
  }

  if (hasVisiblePlatformQuotaColumn.value) {
    tasks.push(
      (async () => {
        try {
          const chunkSize = 6
          for (let index = 0; index < userIds.length; index += chunkSize) {
            if (signal?.aborted) return
            if (typeof expectedSeq === 'number' && expectedSeq !== secondaryDataSeq) return
            const chunk = userIds.slice(index, index + chunkSize)
            const results = await Promise.allSettled(
              chunk.map((userId) => adminAPI.users.getPlatformQuotas(userId))
            )
            if (signal?.aborted) return
            if (typeof expectedSeq === 'number' && expectedSeq !== secondaryDataSeq) return
            const merged = { ...platformQuotaStats.value }
            results.forEach((result, resultIndex) => {
              if (result.status === 'fulfilled') {
                merged[chunk[resultIndex]] = result.value.platform_quotas || []
              }
            })
            platformQuotaStats.value = merged
          }
        } catch (e) {
          if (signal?.aborted) return
          console.error('Failed to load platform quotas:', e)
        }
      })()
    )
  }

  if (tasks.length > 0) {
    await Promise.allSettled(tasks)
  }
}

const buildUserListFilters = () => {
  const attrFilters: Record<number, string> = {}
  for (const [attrId, value] of Object.entries(activeAttributeFilters)) {
    if (value) {
      attrFilters[Number(attrId)] = value
    }
  }

  return {
    role: filters.role as any,
    status: filters.status as any,
    search: searchQuery.value || undefined,
    group_name: filters.group || undefined,
    api_key_group_id: filters.apiKeyGroup ?? undefined,
    attributes: Object.keys(attrFilters).length > 0 ? attrFilters : undefined,
    include_subscriptions: true,
    sort_by: USER_DEFAULT_SORT.sort_by,
    sort_order: USER_DEFAULT_SORT.sort_order
  }
}

function isCanceled(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false
  const errorInfo = error as { name?: string; code?: string }
  return errorInfo.name === 'AbortError' || errorInfo.name === 'CanceledError' || errorInfo.code === 'ERR_CANCELED'
}

const {
  items: users,
  loading,
  loadingMore,
  pagination,
  hasMore: hasMoreUsers,
  reset: resetUsers,
  startObserver: startLoadMoreObserver,
  stopObserver: stopLoadMoreObserver
} = useInfinitePagedList<AdminUser, PaginatedResponse<AdminUser>>({
  pageSize: getPersistedPageSize(),
  sentinelRef: loadMoreSentinelRef,
  fetchPage: (pageState, options) =>
    adminAPI.users.list(
      pageState.page,
      pageState.page_size,
      buildUserListFilters(),
      options
    ),
  onSuccess: (response, mode) => {
    if (mode !== 'append') {
      secondaryDataSeq += 1
      usageStats.value = {}
      userAttributeValues.value = {}
      platformQuotaStats.value = {}
    }
    const userIds = (response.items || []).map((user) => user.id)
    if (userIds.length === 0) return
    const seq = secondaryDataSeq
    window.setTimeout(() => {
      if (seq !== secondaryDataSeq) return
      void loadUsersSecondaryData(userIds, undefined, seq)
    }, 50)
  },
  onError: (error) => {
    const message = error && typeof error === 'object' && 'message' in error
      ? String((error as { message?: unknown }).message || t('admin.users.failedToLoad'))
      : t('admin.users.failedToLoad')
    appStore.showError(message)
    console.error('Error loading users:', error)
  },
  isCanceled
})

const loadUsers = () => {
  secondaryDataSeq += 1
  usageStats.value = {}
  userAttributeValues.value = {}
  platformQuotaStats.value = {}
  return resetUsers()
}

const {
  selectedIds,
  selectedCount,
  allVisibleSelected,
  isSelected,
  toggle,
  toggleVisible,
  clear: clearSelection
} = useTableSelection<AdminUser>({
  rows: users,
  getId: (user) => user.id
})

const handleBulkLimitsSuccess = () => {
  clearSelection()
  void loadUsers()
}

const refreshCurrentPageSecondaryData = () => {
  const userIds = users.value.map((u) => u.id)
  if (userIds.length === 0) return
  const seq = secondaryDataSeq
  void loadUsersSecondaryData(userIds, undefined, seq)
}

// Action Menu State
const activeMenuId = ref<number | null>(null)
const menuPosition = ref<{ top: number; left: number } | null>(null)

const openActionMenu = (user: AdminUser, e: MouseEvent) => {
  if (activeMenuId.value === user.id) {
    closeActionMenu()
  } else {
    const target = e.currentTarget as HTMLElement
    if (!target) {
      closeActionMenu()
      return
    }

    const rect = target.getBoundingClientRect()
    const menuWidth = 200
    const menuHeight = 240
    const padding = 8
    const viewportWidth = window.innerWidth
    const viewportHeight = window.innerHeight

    let left, top

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

    menuPosition.value = { top, left }
    activeMenuId.value = user.id
  }
}

const closeActionMenu = () => {
  activeMenuId.value = null
  menuPosition.value = null
}

// Close menu when clicking outside
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.action-menu-trigger') && !target.closest('.action-menu-content')) {
    closeActionMenu()
  }
  // Close column dropdown when clicking outside
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(target)) {
    showColumnDropdown.value = false
  }
  // Close expanded group dropdown when clicking outside
  if (expandedGroupUserId.value !== null) {
    expandedGroupUserId.value = null
  }
}

// Allowed groups modal state
const showAllowedGroupsModal = ref(false)
const allowedGroupsUser = ref<AdminUser | null>(null)

// Expanded group dropdown state (click to show exclusive groups list)
const expandedGroupUserId = ref<number | null>(null)
const toggleExpandedGroup = (userId: number) => {
  expandedGroupUserId.value = expandedGroupUserId.value === userId ? null : userId
}

// Group replace modal state
const showGroupReplaceModal = ref(false)
const groupReplaceUser = ref<AdminUser | null>(null)
const groupReplaceOldGroup = ref<{ id: number; name: string } | null>(null)

// Balance (Deposit/Withdraw) modal state
const showBalanceModal = ref(false)
const balanceUser = ref<AdminUser | null>(null)
const balanceOperation = ref<'add' | 'subtract'>('add')

// Balance History modal state
const showBalanceHistoryModal = ref(false)
const balanceHistoryUser = ref<AdminUser | null>(null)

// 计算剩余天数
const getDaysRemaining = (expiresAt: string): number => {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diffMs = expires.getTime() - now.getTime()
  return Math.ceil(diffMs / (1000 * 60 * 60 * 24))
}

const loadAttributeDefinitions = async () => {
  try {
    attributeDefinitions.value = await adminAPI.userAttributes.listEnabledDefinitions()
  } catch (e) {
    console.error('Failed to load attribute definitions:', e)
  }
}

// Handle attributes modal close - reload definitions and users
const handleAttributesModalClose = async () => {
  showAttributesModal.value = false
  await loadAttributeDefinitions()
  void loadUsers()
}

const updateAttributeFilter = (attrId: number, value: string) => {
  activeAttributeFilters[attrId] = value
}

// Apply filter and save to localStorage
const applyFilter = () => {
  saveFiltersToStorage()
  void loadUsers()
}
const resetFilters = () => {
  searchQuery.value = ''
  filters.role = ''
  filters.status = ''
  filters.group = ''
  filters.apiKeyGroup = null
  for (const key of Object.keys(activeAttributeFilters)) {
    delete activeAttributeFilters[Number(key)]
  }
  applyFilter()
}

const handleEdit = (user: AdminUser) => {
  editingUser.value = user
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
  editingUser.value = null
}

const handleToggleStatus = async (user: AdminUser) => {
  const newStatus = user.status === 'active' ? 'disabled' : 'active'
  try {
    await adminAPI.users.toggleStatus(user.id, newStatus)
    appStore.showSuccess(
      newStatus === 'active' ? t('admin.users.userEnabled') : t('admin.users.userDisabled')
    )
    void loadUsers()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.users.failedToToggle'))
    console.error('Error toggling user status:', error)
  }
}

const handleViewApiKeys = (user: AdminUser) => {
  viewingUser.value = user
  showApiKeysModal.value = true
}

const closeApiKeysModal = () => {
  showApiKeysModal.value = false
  viewingUser.value = null
}

const handleAllowedGroups = (user: AdminUser) => {
  allowedGroupsUser.value = user
  showAllowedGroupsModal.value = true
}

const closeAllowedGroupsModal = () => {
  showAllowedGroupsModal.value = false
  allowedGroupsUser.value = null
}

const openGroupReplace = (user: AdminUser, group: { id: number; name: string }) => {
  expandedGroupUserId.value = null
  groupReplaceUser.value = user
  groupReplaceOldGroup.value = group
  showGroupReplaceModal.value = true
}

const closeGroupReplaceModal = () => {
  showGroupReplaceModal.value = false
  groupReplaceUser.value = null
  groupReplaceOldGroup.value = null
}

const handleDelete = (user: AdminUser) => {
  deletingUser.value = user
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingUser.value) return
  try {
    await adminAPI.users.delete(deletingUser.value.id)
    appStore.showSuccess(t('common.success'))
    showDeleteDialog.value = false
    deletingUser.value = null
    void loadUsers()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.users.failedToDelete'))
    console.error('Error deleting user:', error)
  }
}

const handleDeposit = (user: AdminUser) => {
  balanceUser.value = user
  balanceOperation.value = 'add'
  showBalanceModal.value = true
}

const handleWithdraw = (user: AdminUser) => {
  balanceUser.value = user
  balanceOperation.value = 'subtract'
  showBalanceModal.value = true
}

const closeBalanceModal = () => {
  showBalanceModal.value = false
  balanceUser.value = null
}

const handleBalanceHistory = (user: AdminUser) => {
  balanceHistoryUser.value = user
  showBalanceHistoryModal.value = true
}

const closeBalanceHistoryModal = () => {
  showBalanceHistoryModal.value = false
  balanceHistoryUser.value = null
}

// Handle deposit from balance history modal
const handleDepositFromHistory = () => {
  if (balanceHistoryUser.value) {
    handleDeposit(balanceHistoryUser.value)
  }
}

// Handle withdraw from balance history modal
const handleWithdrawFromHistory = () => {
  if (balanceHistoryUser.value) {
    handleWithdraw(balanceHistoryUser.value)
  }
}

// 滚动时关闭菜单
const handleScroll = () => {
  closeActionMenu()
}

onMounted(async () => {
  await loadAttributeDefinitions()
  loadSavedFilters()
  loadSavedColumns()
  await loadUsers()
  startLoadMoreObserver()
  loadAllGroups()
  loadAllGroupsForApiKeyFilter()
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('scroll', handleScroll, true)
})

onUnmounted(() => {
  stopLoadMoreObserver()
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('scroll', handleScroll, true)
})
</script>

<style scoped>
.users-filter-row {
  @apply grid gap-3 md:grid-cols-[14rem_8.5rem_8.5rem_11rem_11rem_auto];
}

.users-filter-search,
.users-filter-select,
.users-filter-group,
.users-filter-api-key-group,
.users-filter-attribute {
  @apply min-w-0 w-full;
}

.users-filter-actions {
  @apply flex shrink-0 items-center gap-2;
}

.users-filter-actions .btn {
  @apply whitespace-nowrap;
}

.users-table-card {
  @apply overflow-hidden;
}

.admin-row-action {
  @apply inline-flex min-w-[52px] flex-col items-center justify-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium leading-none text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-950 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white;
}

.admin-row-action-label {
  @apply whitespace-nowrap text-[11px] leading-none;
}

.users-page-layout :deep(.users-table-card .table-wrapper),
.users-page-layout :deep(.users-table-card .table-wrapper.is-natural-flow) {
  @apply overflow-x-auto overflow-y-visible;
}

.users-page-layout :deep(.users-table-card .cch-mobile-table-list) {
  @apply p-4;
}

.users-page-layout :deep(.users-table-card tbody tr + tr) {
  border-top: 1px solid rgb(226 232 240 / 0.8);
}

.users-page-layout :deep(.users-table-card tbody td) {
  border-top: 0;
}

.users-page-layout :deep(.users-table-card .sticky-col-left),
.users-page-layout :deep(.users-table-card .sticky-col-right) {
  position: static;
}

.users-page-layout :deep(.users-table-card .sticky-col-left::after),
.users-page-layout :deep(.users-table-card .sticky-col-right::before) {
  display: none;
}

.dark .users-page-layout :deep(.users-table-card tbody tr + tr) {
  border-top-color: rgb(51 65 85 / 0.75);
}

</style>
