import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppHeader.vue')
const componentSource = readFileSync(componentPath, 'utf8')

function between(source: string, start: string, end: string): string {
  const startIndex = source.indexOf(start)
  const endIndex = source.indexOf(end, startIndex)

  expect(startIndex).toBeGreaterThanOrEqual(0)
  expect(endIndex).toBeGreaterThan(startIndex)

  return source.slice(startIndex, endIndex)
}

describe('AppHeader account dropdown', () => {
  it('uses the account dropdown as the single entry for personal menu links', () => {
    const dropdownTemplate = between(
      componentSource,
      '<div class="py-1">',
      '<a'
    )
    const accountDropdownBlock = between(
      componentSource,
      'const accountDropdownItems = computed(() => finalizeNav([',
      'function finalizeNav'
    )

    expect(dropdownTemplate).toContain('v-for="item in accountDropdownItems"')
    expect(dropdownTemplate).not.toContain("authStore.isAdmin ? '/admin/users' : '/keys'")
    expect(accountDropdownBlock).toContain("{ path: '/profile', label: t('nav.profile')")
    expect(accountDropdownBlock).toContain("{ path: '/keys', label: t('nav.apiKeys')")
    expect(accountDropdownBlock).toContain("{ path: '/usage', label: t('nav.usage')")
    expect(accountDropdownBlock).toContain("{ path: '/subscriptions', label: t('nav.mySubscriptions')")
    expect(accountDropdownBlock).toContain("{ path: '/orders', label: t('nav.myOrders')")
    expect(accountDropdownBlock).not.toContain("'/admin/users'")
  })
})
