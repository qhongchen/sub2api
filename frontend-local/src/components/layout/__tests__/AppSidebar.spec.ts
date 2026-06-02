import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

function between(source: string, start: string, end: string): string {
  const startIndex = source.indexOf(start)
  const endIndex = source.indexOf(end, startIndex)

  expect(startIndex).toBeGreaterThanOrEqual(0)
  expect(endIndex).toBeGreaterThan(startIndex)

  return source.slice(startIndex, endIndex)
}

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar admin navigation', () => {
  it('keeps request logs in the admin main list and moves usage below billing entries', () => {
    const adminNavBlock = between(
      componentSource,
      'const adminNavItems = computed((): NavItem[] => {',
      'function toggleSidebar()'
    )

    expect(adminNavBlock).toContain("{ path: '/admin/request-logs', label: t('nav.requestLogs')")
    expect(adminNavBlock).toContain("{ path: '/admin/usage', label: t('nav.usage')")
    expect(adminNavBlock.indexOf("'/admin/request-logs'")).toBeLessThan(adminNavBlock.indexOf("'/admin/accounts'"))
    expect(adminNavBlock.indexOf("'/admin/subscriptions'")).toBeLessThan(adminNavBlock.indexOf("'/admin/usage'"))
    expect(adminNavBlock.indexOf("'/admin/usage'")).toBeLessThan(adminNavBlock.indexOf("'/admin/announcements'"))
  })

  it('removes the separate admin personal sidebar section', () => {
    expect(componentSource).not.toContain('personalNavItems')
    expect(componentSource).not.toContain("{{ t('nav.myAccount') }}")
  })
})
