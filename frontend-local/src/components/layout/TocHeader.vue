<template>
  <header class="consumer-header">
    <div class="consumer-header-inner">
      <div class="flex min-w-0 items-center gap-3">
        <button
          type="button"
          class="consumer-icon-btn md:hidden"
          aria-label="Toggle navigation menu"
          @click="mobileOpen = true"
        >
          <Icon name="menu" size="md" />
        </button>

        <router-link :to="homePath" class="consumer-brand consumer-brand-top">
          <span class="consumer-brand-mark">
            <img
              v-if="settingsLoaded"
              :src="siteLogo || '/logo.png'"
              alt="Logo"
              class="h-full w-full object-contain"
            />
          </span>
          <span class="hidden min-w-0 truncate text-sm font-semibold text-gray-950 dark:text-white sm:block">
            {{ siteName }}
          </span>
        </router-link>

        <nav class="consumer-nav hidden md:flex" aria-label="Primary navigation">
          <router-link
            v-for="item in primaryNavItems"
            :key="item.path"
            :to="item.path"
            class="consumer-nav-link"
            :class="{ 'consumer-nav-link-active': isActive(item.path) }"
            :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
          >
            {{ item.label }}
          </router-link>

          <div
            v-if="isAdmin"
            ref="adminSettingsRef"
            class="relative overflow-visible"
            @pointerenter="openAdminSettingsDelayed"
            @pointerleave="closeAdminSettingsDelayed"
          >
            <button
              type="button"
              class="consumer-nav-link"
              :class="{ 'consumer-nav-link-active': isAdminSettingsActive }"
              @click="adminSettingsOpen = !adminSettingsOpen"
            >
              {{ t('common.more') }}
              <Icon
                name="chevronDown"
                size="xs"
                class="transition-transform"
                :class="{ 'rotate-180': adminSettingsOpen }"
              />
            </button>

            <transition name="dropdown">
              <div v-if="adminSettingsOpen" class="consumer-dropdown consumer-system-dropdown left-0 mt-2">
                <div class="space-y-3 p-1">
                  <div
                    v-for="group in adminSettingsGroups"
                    :key="group.label"
                    class="consumer-system-group"
                  >
                    <div class="consumer-system-group-title">{{ group.label }}</div>
                    <router-link
                      v-for="item in group.items"
                      :key="item.path"
                      :to="item.path"
                      class="consumer-system-item"
                      :class="{
                        'consumer-system-item-active': isActive(item.path),
                        'consumer-settings-separator': item.separatorBefore
                      }"
                      :id="tourIdForPath(item.path)"
                      @click="closeAdminSettings"
                    >
                      <Icon :name="item.icon" size="sm" />
                      <span class="min-w-0 truncate">{{ item.label }}</span>
                    </router-link>
                  </div>
                </div>
              </div>
            </transition>
          </div>

          <div
            v-if="!isAdmin && secondaryNavItems.length"
            ref="moreRef"
            class="relative overflow-visible"
            @pointerenter="openMoreDelayed"
            @pointerleave="closeMoreDelayed"
          >
            <button
              type="button"
              class="consumer-nav-link"
              :class="{ 'consumer-nav-link-active': isSecondaryActive }"
              @click="toggleMoreMenu"
            >
              {{ t('common.more') }}
              <Icon
                name="chevronDown"
                size="xs"
                class="transition-transform"
                :class="{ 'rotate-180': moreOpen }"
              />
            </button>

            <transition name="dropdown">
              <div v-if="moreOpen" class="consumer-dropdown left-0 mt-2 w-56">
                <router-link
                  v-for="item in secondaryNavItems"
                  :key="item.path"
                  :to="item.path"
                  class="consumer-dropdown-item"
                  :class="{ 'consumer-dropdown-item-active': isActive(item.path) }"
                  @click="closeMoreMenu"
                >
                  <Icon :name="item.icon" size="sm" />
                  <span>{{ item.label }}</span>
                </router-link>
              </div>
            </transition>
          </div>
        </nav>
      </div>

      <div class="flex items-center gap-2">
        <AnnouncementBell v-if="user" />

        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="consumer-icon-btn hidden sm:inline-flex"
          :title="t('nav.docs')"
        >
          <Icon name="book" size="sm" />
        </a>

        <LocaleSwitcher class="hidden sm:block" />
        <SubscriptionProgressMini v-if="user" class="hidden lg:flex" />

        <button
          type="button"
          class="consumer-icon-btn"
          :title="isDark ? t('nav.lightMode') : t('nav.darkMode')"
          @click="toggleTheme"
        >
          <Icon v-if="isDark" name="sun" size="sm" class="text-amber-500" />
          <Icon v-else name="moon" size="sm" />
        </button>

        <div v-if="user" ref="userMenuRef" class="relative overflow-visible">
          <button
            type="button"
            class="consumer-user-trigger"
            aria-label="User Menu"
            @click="userMenuOpen = !userMenuOpen"
          >
            <span class="consumer-avatar">
              <img
                v-if="avatarUrl"
                :src="avatarUrl"
                :alt="displayName"
                class="h-full w-full object-cover"
              />
              <span v-else>{{ userInitials }}</span>
            </span>
            <span class="hidden max-w-28 truncate text-sm font-medium text-gray-900 dark:text-white md:block">
              {{ displayName }}
            </span>
            <Icon name="chevronDown" size="xs" class="hidden text-gray-400 md:block" />
          </button>

          <transition name="dropdown">
            <div v-if="userMenuOpen" class="consumer-dropdown right-0 mt-2 w-56">
              <div class="border-b border-gray-100 px-4 py-3 dark:border-dark-700">
                <div class="truncate text-sm font-medium text-gray-900 dark:text-white">
                  {{ displayName }}
                </div>
                <div class="truncate text-xs text-gray-500 dark:text-dark-400">
                  {{ user.email }}
                </div>
              </div>

              <div class="border-b border-gray-100 px-4 py-2 dark:border-dark-700 sm:hidden">
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ t('common.balance') }}</div>
                <div class="text-sm font-semibold text-orange-600 dark:text-orange-300">
                  ${{ user.balance?.toFixed(2) || '0.00' }}
                </div>
              </div>

              <router-link to="/profile" class="consumer-dropdown-item" @click="closeUserMenu">
                <Icon name="user" size="sm" />
                {{ t('nav.profile') }}
              </router-link>
              <router-link :to="isAdmin ? '/admin/users' : '/keys'" class="consumer-dropdown-item" @click="closeUserMenu">
                <Icon :name="isAdmin ? 'users' : 'key'" size="sm" />
                {{ isAdmin ? t('nav.users') : t('nav.apiKeys') }}
              </router-link>

              <button
                type="button"
                class="consumer-dropdown-item w-full text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                @click="handleLogout"
              >
                <Icon name="login" size="sm" />
                {{ t('nav.logout') }}
              </button>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </header>

  <transition name="fade">
    <div v-if="mobileOpen" class="fixed inset-0 z-40 bg-black/40 md:hidden" @click="mobileOpen = false">
      <aside class="consumer-mobile-panel" @click.stop>
        <div class="mb-6 flex items-center justify-between">
          <router-link :to="homePath" class="consumer-brand" @click="mobileOpen = false">
            <span class="consumer-brand-mark">
              <img
                v-if="settingsLoaded"
                :src="siteLogo || '/logo.png'"
                alt="Logo"
                class="h-full w-full object-contain"
              />
            </span>
            <span class="min-w-0 truncate text-sm font-semibold text-gray-950 dark:text-white">
              {{ siteName }}
            </span>
          </router-link>
          <button type="button" class="consumer-icon-btn" aria-label="Close navigation menu" @click="mobileOpen = false">
            <Icon name="x" size="md" />
          </button>
        </div>

        <nav class="flex flex-col gap-2" aria-label="Mobile navigation">
          <router-link
            v-for="item in mobileNavItems"
            :key="item.path"
            :to="item.path"
            class="consumer-mobile-link"
            :class="{ 'consumer-mobile-link-active': isActive(item.path) }"
            :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
            @click="mobileOpen = false"
          >
            <Icon :name="item.icon" size="sm" />
            <span>{{ item.label }}</span>
          </router-link>

          <template v-if="isAdmin && adminSettingsGroups.length">
            <div class="consumer-mobile-section-title">{{ t('common.more') }}</div>
            <template
              v-for="group in adminSettingsGroups"
              :key="group.label"
            >
              <div class="consumer-mobile-subsection-title">{{ group.label }}</div>
              <router-link
                v-for="item in group.items"
                :key="item.path"
                :to="item.path"
                class="consumer-mobile-link"
                :class="{ 'consumer-mobile-link-active': isActive(item.path) }"
                :id="tourIdForPath(item.path)"
                @click="mobileOpen = false"
              >
                <Icon :name="item.icon" size="sm" />
                <span>{{ item.label }}</span>
              </router-link>
            </template>
          </template>
        </nav>

        <div class="mt-auto border-t border-gray-100 pt-4 dark:border-dark-700">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="consumer-mobile-link"
          >
            <Icon name="book" size="sm" />
            <span>{{ t('nav.docs') }}</span>
          </a>
          <LocaleSwitcher class="mt-3" />
        </div>
      </aside>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAdminSettingsStore, useAppStore, useAuthStore } from '@/stores'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import Icon from '@/components/icons/Icon.vue'
import { FeatureFlags, makeSidebarFlag } from '@/utils/featureFlags'

type TocIconName =
  | 'grid'
  | 'key'
  | 'chart'
  | 'server'
  | 'creditCard'
  | 'gift'
  | 'users'
  | 'user'
  | 'document'
  | 'cog'
  | 'shield'
  | 'globe'
  | 'bell'

interface TocNavItem {
  path: string
  label: string
  icon: TocIconName
  hideInSimpleMode?: boolean
  featureFlag?: () => boolean
  separatorBefore?: boolean
}

interface TocNavGroup {
  label: string
  items: TocNavItem[]
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()

const mobileOpen = ref(false)
const moreOpen = ref(false)
const adminSettingsOpen = ref(false)
const userMenuOpen = ref(false)
const userMenuRef = ref<HTMLElement | null>(null)
const moreRef = ref<HTMLElement | null>(null)
const adminSettingsRef = ref<HTMLElement | null>(null)
const isDark = ref(document.documentElement.classList.contains('dark'))
let adminSettingsOpenTimer: ReturnType<typeof setTimeout> | null = null
let adminSettingsCloseTimer: ReturnType<typeof setTimeout> | null = null
let moreOpenTimer: ReturnType<typeof setTimeout> | null = null
let moreCloseTimer: ReturnType<typeof setTimeout> | null = null

const flagChannelMonitor = makeSidebarFlag(FeatureFlags.channelMonitor)
const flagPayment = makeSidebarFlag(FeatureFlags.payment)
const flagAvailableChannels = makeSidebarFlag(FeatureFlags.availableChannels)
const flagAffiliate = makeSidebarFlag(FeatureFlags.affiliate)
const flagRiskControl = makeSidebarFlag(FeatureFlags.riskControl)
const flagOpsMonitoring = () => adminSettingsStore.opsMonitoringEnabled
const flagAdminPayment = () => adminSettingsStore.paymentEnabled

const user = computed(() => authStore.user)
const isAdmin = computed(() => authStore.isAdmin)
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => appStore.siteLogo)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const docUrl = computed(() => appStore.docUrl)
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const homePath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))

const displayName = computed(() => {
  if (!user.value) return ''
  return user.value.username || user.value.email?.split('@')[0] || ''
})

const userInitials = computed(() => {
  if (!user.value) return ''
  const source = user.value.username || user.value.email?.split('@')[0] || ''
  return source.substring(0, 2).toUpperCase()
})

const customMenuItems = computed(() => {
  const items = appStore.cachedPublicSettings?.custom_menu_items ?? []
  return items
    .filter((item) => item.visibility === 'user')
    .sort((a, b) => a.sort_order - b.sort_order)
    .map((item): TocNavItem => ({
      path: `/custom/${item.id}`,
      label: item.label,
      icon: 'document',
      hideInSimpleMode: true
    }))
})

const customAdminMenuItems = computed(() => {
  return adminSettingsStore.customMenuItems
    .filter((item) => item.visibility === 'admin')
    .sort((a, b) => a.sort_order - b.sort_order)
    .map((item): TocNavItem => ({
      path: `/custom/${item.id}`,
      label: item.label,
      icon: 'document',
      hideInSimpleMode: true
    }))
})

const userPrimaryNavItems = computed(() => finalizeNav([
  { path: '/dashboard', label: t('nav.dashboard'), icon: 'grid' },
  { path: '/usage', label: t('nav.usage'), icon: 'chart', hideInSimpleMode: true },
  { path: '/keys', label: t('nav.apiKeys'), icon: 'key' },
  {
    path: '/available-channels',
    label: t('nav.availableChannels'),
    icon: 'server',
    hideInSimpleMode: true,
    featureFlag: flagAvailableChannels
  },
  { path: '/monitor', label: t('nav.channelStatus'), icon: 'server', featureFlag: flagChannelMonitor }
]))

const adminPrimaryNavItems = computed(() => finalizeNav([
  { path: '/admin/dashboard', label: t('nav.dashboard'), icon: 'grid' },
  { path: '/admin/usage', label: t('nav.usage'), icon: 'chart', hideInSimpleMode: true },
  { path: '/admin/accounts', label: t('nav.accounts'), icon: 'globe' },
  { path: '/admin/users', label: t('nav.users'), icon: 'users' }
]))

const primaryNavItems = computed(() => (isAdmin.value ? adminPrimaryNavItems.value : userPrimaryNavItems.value))

const secondaryNavItems = computed(() => finalizeNav([
  { path: '/subscriptions', label: t('nav.mySubscriptions'), icon: 'creditCard', hideInSimpleMode: true },
  { path: '/purchase', label: t('nav.buySubscription'), icon: 'creditCard', hideInSimpleMode: true, featureFlag: flagPayment },
  { path: '/orders', label: t('nav.myOrders'), icon: 'document', hideInSimpleMode: true, featureFlag: flagPayment },
  { path: '/redeem', label: t('nav.redeem'), icon: 'gift', hideInSimpleMode: true },
  { path: '/affiliate', label: t('nav.affiliate'), icon: 'users', hideInSimpleMode: true, featureFlag: flagAffiliate },
  { path: '/profile', label: t('nav.profile'), icon: 'user' },
  ...customMenuItems.value
]))

const adminSettingsGroups = computed((): TocNavGroup[] => {
  if (!isAdmin.value) return []

  const groups: TocNavGroup[] = [
    {
      label: t('nav.usersAndGroups'),
      items: [
        { path: '/admin/groups', label: t('nav.groups'), icon: 'users', hideInSimpleMode: true }
      ]
    },
    {
      label: t('nav.channelsAndAccounts'),
      items: [
        { path: '/admin/channels/pricing', label: t('nav.channelPricing'), icon: 'server', hideInSimpleMode: true },
        {
          path: '/admin/channels/monitor',
          label: t('nav.channelMonitor'),
          icon: 'server',
          hideInSimpleMode: true,
          featureFlag: flagChannelMonitor
        },
        { path: '/admin/proxies', label: t('nav.proxies'), icon: 'server' }
      ]
    },
    {
      label: t('nav.operationsAndBilling'),
      items: [
        { path: '/admin/subscriptions', label: t('nav.subscriptions'), icon: 'creditCard', hideInSimpleMode: true },
        { path: '/admin/announcements', label: t('nav.announcements'), icon: 'bell' },
        { path: '/admin/redeem', label: t('nav.redeemCodes'), icon: 'gift', hideInSimpleMode: true },
        { path: '/admin/promo-codes', label: t('nav.promoCodes'), icon: 'gift', hideInSimpleMode: true },
        {
          path: '/admin/risk-control',
          label: t('nav.riskControl'),
          icon: 'shield',
          hideInSimpleMode: true,
          featureFlag: flagRiskControl
        },
        {
          path: '/admin/affiliates/invites',
          label: t('nav.affiliateInviteRecords'),
          icon: 'users',
          hideInSimpleMode: true,
          featureFlag: flagAffiliate
        },
        {
          path: '/admin/affiliates/rebates',
          label: t('nav.affiliateRebateRecords'),
          icon: 'document',
          hideInSimpleMode: true,
          featureFlag: flagAffiliate
        },
        {
          path: '/admin/affiliates/transfers',
          label: t('nav.affiliateTransferRecords'),
          icon: 'creditCard',
          hideInSimpleMode: true,
          featureFlag: flagAffiliate
        },
        {
          path: '/admin/orders/dashboard',
          label: t('nav.paymentDashboard'),
          icon: 'chart',
          hideInSimpleMode: true,
          featureFlag: flagAdminPayment
        },
        {
          path: '/admin/orders',
          label: t('nav.orderManagement'),
          icon: 'document',
          hideInSimpleMode: true,
          featureFlag: flagAdminPayment
        },
        {
          path: '/admin/orders/plans',
          label: t('nav.paymentPlans'),
          icon: 'creditCard',
          hideInSimpleMode: true,
          featureFlag: flagAdminPayment
        }
      ]
    },
    {
      label: t('nav.systemConfiguration'),
      items: [
        { path: '/admin/ops', label: t('nav.ops'), icon: 'chart', featureFlag: flagOpsMonitoring },
        { path: '/admin/settings', label: t('nav.settings'), icon: 'cog' },
        ...customAdminMenuItems.value
      ]
    }
  ]

  return groups
    .map((group) => ({ ...group, items: finalizeNav(group.items) }))
    .filter((group) => group.items.length > 0)
})

const mobileNavItems = computed(() => [
  ...primaryNavItems.value,
  ...secondaryNavItems.value
])

const isSecondaryActive = computed(() => secondaryNavItems.value.some((item) => isActive(item.path)))
const isAdminSettingsActive = computed(() => adminSettingsGroups.value.some((group) => (
  group.items.some((item) => isActive(item.path))
)))

function finalizeNav(items: TocNavItem[]): TocNavItem[] {
  const visible = items.filter((item) => item.featureFlag?.() !== false)
  return authStore.isSimpleMode ? visible.filter((item) => !item.hideInSimpleMode) : visible
}

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(`${path}/`)
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function closeUserMenu() {
  userMenuOpen.value = false
}

function closeAdminSettings() {
  clearAdminSettingsTimers()
  adminSettingsOpen.value = false
}

function toggleMoreMenu() {
  clearMoreTimers()
  moreOpen.value = !moreOpen.value
}

function closeMoreMenu() {
  clearMoreTimers()
  moreOpen.value = false
}

function clearAdminSettingsTimers() {
  if (adminSettingsOpenTimer) {
    clearTimeout(adminSettingsOpenTimer)
    adminSettingsOpenTimer = null
  }
  if (adminSettingsCloseTimer) {
    clearTimeout(adminSettingsCloseTimer)
    adminSettingsCloseTimer = null
  }
}

function clearMoreTimers() {
  if (moreOpenTimer) {
    clearTimeout(moreOpenTimer)
    moreOpenTimer = null
  }
  if (moreCloseTimer) {
    clearTimeout(moreCloseTimer)
    moreCloseTimer = null
  }
}

function openAdminSettingsDelayed() {
  if (adminSettingsCloseTimer) {
    clearTimeout(adminSettingsCloseTimer)
    adminSettingsCloseTimer = null
  }
  if (adminSettingsOpen.value || adminSettingsOpenTimer) return
  adminSettingsOpenTimer = setTimeout(() => {
    adminSettingsOpen.value = true
    adminSettingsOpenTimer = null
  }, 150)
}

function closeAdminSettingsDelayed() {
  if (adminSettingsOpenTimer) {
    clearTimeout(adminSettingsOpenTimer)
    adminSettingsOpenTimer = null
  }
  if (adminSettingsCloseTimer) return
  adminSettingsCloseTimer = setTimeout(() => {
    adminSettingsOpen.value = false
    adminSettingsCloseTimer = null
  }, 200)
}

function openMoreDelayed() {
  if (moreCloseTimer) {
    clearTimeout(moreCloseTimer)
    moreCloseTimer = null
  }
  if (moreOpen.value || moreOpenTimer) return
  moreOpenTimer = setTimeout(() => {
    moreOpen.value = true
    moreOpenTimer = null
  }, 150)
}

function closeMoreDelayed() {
  if (moreOpenTimer) {
    clearTimeout(moreOpenTimer)
    moreOpenTimer = null
  }
  if (moreCloseTimer) return
  moreCloseTimer = setTimeout(() => {
    moreOpen.value = false
    moreCloseTimer = null
  }, 200)
}

function tourIdForPath(path: string): string | undefined {
  if (path === '/admin/groups') return 'sidebar-group-manage'
  if (path === '/admin/accounts') return 'sidebar-channel-manage'
  if (path === '/admin/subscriptions') return 'sidebar-wallet'
  return undefined
}

async function handleLogout() {
  closeUserMenu()
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as Node
  if (userMenuRef.value && !userMenuRef.value.contains(target)) {
    userMenuOpen.value = false
  }
  if (moreRef.value && !moreRef.value.contains(target)) {
    closeMoreMenu()
  }
  if (adminSettingsRef.value && !adminSettingsRef.value.contains(target)) {
    adminSettingsOpen.value = false
  }
}

const savedTheme = localStorage.getItem('theme')
if (
  savedTheme === 'dark' ||
  (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
) {
  isDark.value = true
  document.documentElement.classList.add('dark')
}

watch(
  () => route.fullPath,
  () => {
    mobileOpen.value = false
    closeMoreMenu()
    closeAdminSettings()
    userMenuOpen.value = false
  }
)

watch(
  isAdmin,
  (value) => {
    if (value) {
      void adminSettingsStore.fetch()
    }
  },
  { immediate: true }
)

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (isAdmin.value) {
    void adminSettingsStore.fetch()
  }
})

onBeforeUnmount(() => {
  clearAdminSettingsTimers()
  clearMoreTimers()
  document.removeEventListener('click', handleClickOutside)
})
</script>
