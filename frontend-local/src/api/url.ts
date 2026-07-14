const DEFAULT_API_BASE_URL = '/api/v1'

const normalizePath = (path: string): string => (path.startsWith('/') ? path : `/${path}`)

export function getAPIBaseURL(): string {
  const raw = String(import.meta.env.VITE_API_BASE_URL || DEFAULT_API_BASE_URL).trim()
  const withoutTrailingSlash = (raw || DEFAULT_API_BASE_URL).replace(/\/+$/, '')
  if (/^[a-z][a-z\d+.-]*:\/\//i.test(withoutTrailingSlash) || withoutTrailingSlash.startsWith('//')) {
    return withoutTrailingSlash
  }
  return normalizePath(withoutTrailingSlash)
}

export function buildGatewayUrl(path: string): string {
  const suffix = normalizePath(path)
  try {
    const baseURL = getAPIBaseURL()
    const origin =
      typeof window === 'undefined'
        ? new URL(baseURL).origin
        : new URL(baseURL, window.location.origin).origin
    return `${origin}${suffix}`
  } catch {
    return suffix
  }
}
