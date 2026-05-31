import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../TocHeader.vue')
const componentSource = readFileSync(componentPath, 'utf8')

function between(source: string, start: string, end: string): string {
  const startIndex = source.indexOf(start)
  const endIndex = source.indexOf(end, startIndex)

  expect(startIndex).toBeGreaterThanOrEqual(0)
  expect(endIndex).toBeGreaterThan(startIndex)

  return source.slice(startIndex, endIndex)
}

describe('TocHeader admin top navigation', () => {
  it('keeps the admin top-level menu focused on dashboard, usage, accounts, channel status, and more', () => {
    const adminPrimaryBlock = between(
      componentSource,
      'const adminPrimaryNavItems = computed(() => finalizeNav([',
      'const primaryNavItems = computed'
    )

    expect(adminPrimaryBlock).toContain("{ path: '/admin/dashboard', label: t('nav.dashboard')")
    expect(adminPrimaryBlock).toContain("{ path: '/admin/usage', label: t('nav.usage')")
    expect(adminPrimaryBlock).toContain("{ path: '/admin/accounts', label: t('nav.accounts')")
    expect(adminPrimaryBlock).toContain("{ path: '/monitor', label: t('nav.channelStatus')")
    expect(adminPrimaryBlock.indexOf("'/admin/dashboard'")).toBeLessThan(adminPrimaryBlock.indexOf("'/admin/usage'"))
    expect(adminPrimaryBlock.indexOf("'/admin/usage'")).toBeLessThan(adminPrimaryBlock.indexOf("'/admin/accounts'"))
    expect(adminPrimaryBlock.indexOf("'/admin/accounts'")).toBeLessThan(adminPrimaryBlock.indexOf("'/monitor'"))
    expect(adminPrimaryBlock).not.toContain("{ path: '/admin/users'")
  })

  it('labels the admin management dropdown as more and hides the user more menu for admins', () => {
    const adminDropdownTemplate = between(
      componentSource,
      'v-if="isAdmin && adminSettingsGroups.length"',
      'v-if="!isAdmin && secondaryNavItems.length"'
    )

    expect(adminDropdownTemplate).toContain("{{ t('common.more') }}")
    expect(adminDropdownTemplate).toContain('v-if="adminSettingsOpen"')
    expect(adminDropdownTemplate).toContain('@pointerenter="openAdminSettingsDelayed"')
    expect(adminDropdownTemplate).not.toContain("{{ t('nav.settings') }}")
    expect(adminDropdownTemplate).not.toContain('activeAdminGroupId')
    expect(componentSource).toContain('v-if="!isAdmin && secondaryNavItems.length"')
  })

  it('renders admin more menu with visible category groups instead of a flat item list', () => {
    const adminDropdownTemplate = between(
      componentSource,
      '<div v-if="adminSettingsOpen"',
      '</transition>'
    )

    expect(adminDropdownTemplate).toContain('v-for="group in adminSettingsGroups"')
    expect(adminDropdownTemplate).toContain('class="consumer-system-group-title"')
    expect(adminDropdownTemplate).toContain('{{ group.label }}')
    expect(adminDropdownTemplate).not.toContain("{ path: '/admin/usage'")
    expect(adminDropdownTemplate).not.toContain('v-for="item in adminSettingsItems"')
  })

  it('adds a mine group for personal admin links and keeps user management under users and access', () => {
    const mineGroupBlock = between(
      componentSource,
      "label: t('nav.my')",
      "label: t('nav.usersAndGroups')"
    )
    const usersGroupBlock = between(
      componentSource,
      "label: t('nav.usersAndGroups')",
      "label: t('nav.channelsAndAccounts')"
    )
    const secondaryNavBlock = between(
      componentSource,
      'const secondaryNavItems = computed(() => finalizeNav([',
      'const adminSettingsGroups = computed'
    )

    expect(mineGroupBlock).toContain("{ path: '/usage', label: t('nav.usage')")
    expect(mineGroupBlock).toContain("{ path: '/keys', label: t('nav.apiKeys')")
    expect(mineGroupBlock).toContain('...secondaryNavItems.value')
    expect(secondaryNavBlock).toContain("{ path: '/subscriptions', label: t('nav.mySubscriptions')")
    expect(secondaryNavBlock).toContain("{ path: '/purchase', label: t('nav.buySubscription')")
    expect(secondaryNavBlock).toContain("{ path: '/orders', label: t('nav.myOrders')")
    expect(secondaryNavBlock).toContain("{ path: '/redeem', label: t('nav.redeem')")
    expect(secondaryNavBlock).toContain("{ path: '/affiliate', label: t('nav.affiliate')")
    expect(secondaryNavBlock).toContain("{ path: '/profile', label: t('nav.profile')")
    expect(usersGroupBlock).toContain("{ path: '/admin/users', label: t('nav.users')")
    expect(usersGroupBlock).toContain("{ path: '/admin/groups', label: t('nav.groups')")
  })

  it('assigns stable ids to every admin top-level category', () => {
    const adminSettingsBlock = between(
      componentSource,
      'const adminSettingsGroups = computed',
      'const mobileNavItems = computed'
    )

    expect(adminSettingsBlock).toContain("id: 'my'")
    expect(adminSettingsBlock).toContain("id: 'users-and-groups'")
    expect(adminSettingsBlock).toContain("id: 'channels-and-accounts'")
    expect(adminSettingsBlock).toContain("id: 'operations-and-billing'")
    expect(adminSettingsBlock).toContain("id: 'system-configuration'")
  })

  it('merges ops monitoring and system settings into one admin group', () => {
    const adminSettingsBlock = between(
      componentSource,
      "label: t('nav.systemConfiguration')",
      'return groups'
    )

    expect(adminSettingsBlock).toContain("{ path: '/admin/ops'")
    expect(adminSettingsBlock).toContain("{ path: '/admin/request-logs', label: t('nav.requestLogs')")
    expect(adminSettingsBlock).toContain("{ path: '/admin/ops', label: t('nav.ops'), icon: 'chart', featureFlag: flagOpsMonitoring }")
    expect(adminSettingsBlock).toContain("{ path: '/admin/request-logs', label: t('nav.requestLogs'), icon: 'document' }")
    expect(adminSettingsBlock).toContain("{ path: '/admin/settings'")
    expect(componentSource).not.toContain("label: t('nav.systemOverview')")
  })

  it('removes the top balance display from the header', () => {
    const topHeaderBlock = between(
      componentSource,
      '<LocaleSwitcher class="hidden sm:block" />',
      '<button'
    )

    expect(topHeaderBlock).not.toContain('Icon name="dollar" size="sm"')
    expect(topHeaderBlock).not.toContain('rounded-full border border-orange-200/80 bg-orange-50 px-3 py-1.5')
  })
})
