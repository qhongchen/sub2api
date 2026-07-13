export interface PeakRateFields {
  peak_rate_enabled?: boolean
  peak_start?: string
  peak_end?: string
  peak_rate_multiplier?: number
}

export function hasPeakRate(fields?: PeakRateFields | null): boolean {
  return Boolean(fields?.peak_rate_enabled && fields.peak_start && fields.peak_end)
}

export function serverTimezoneLabel(utcOffset?: string | null): string {
  return utcOffset ? `UTC${utcOffset}` : ''
}

export function formatPeakRateWindow(
  fields: PeakRateFields | null | undefined,
  timezoneLabel?: string
): string {
  if (!hasPeakRate(fields) || !fields) return ''
  const window = `${fields.peak_start}-${fields.peak_end} ×${fields.peak_rate_multiplier ?? 1}`
  return timezoneLabel ? `${window} (${timezoneLabel})` : window
}
