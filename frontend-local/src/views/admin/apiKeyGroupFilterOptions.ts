import type { AdminGroup } from '@/types'

export interface ApiKeyGroupFilterOption {
  value: number | null
  label: string
  kind?: 'group'
  disabled?: boolean
}

export interface ApiKeyGroupFilterLabels {
  all: string
  exclusive: string
  public: string
  subscription: string
  disabled: string
}

const HEADER_EXCLUSIVE = -1
const HEADER_PUBLIC = -2
const HEADER_SUBSCRIPTION = -3
const HEADER_DISABLED = -4

export function buildApiKeyGroupFilterOptions(
  groups: AdminGroup[],
  labels: ApiKeyGroupFilterLabels
): ApiKeyGroupFilterOption[] {
  const exclusive: ApiKeyGroupFilterOption[] = []
  const publicGroups: ApiKeyGroupFilterOption[] = []
  const subscription: ApiKeyGroupFilterOption[] = []
  const disabledGroups: ApiKeyGroupFilterOption[] = []

  for (const group of groups) {
    const item: ApiKeyGroupFilterOption = { value: group.id, label: group.name }
    if (group.status !== 'active') {
      disabledGroups.push(item)
    } else if (group.subscription_type === 'subscription') {
      subscription.push(item)
    } else if (group.is_exclusive) {
      exclusive.push(item)
    } else {
      publicGroups.push(item)
    }
  }

  const options: ApiKeyGroupFilterOption[] = [{ value: null, label: labels.all }]
  const sections: Array<[string, number, ApiKeyGroupFilterOption[]]> = [
    [labels.exclusive, HEADER_EXCLUSIVE, exclusive],
    [labels.public, HEADER_PUBLIC, publicGroups],
    [labels.subscription, HEADER_SUBSCRIPTION, subscription],
    [labels.disabled, HEADER_DISABLED, disabledGroups],
  ]

  for (const [label, headerValue, items] of sections) {
    if (items.length === 0) continue
    options.push({ value: headerValue, label, kind: 'group', disabled: true })
    options.push(...items)
  }

  return options
}
