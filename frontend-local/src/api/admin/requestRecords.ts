import { apiClient } from '../client'
import type { ImageSizeBreakdown, ImageSizeSource, PaginatedResponse, UsageRequestType } from '@/types'

export type AdminRequestLogKind = 'success' | 'pending' | 'error'
export type AdminRequestLogKindFilter = AdminRequestLogKind | 'all'
export type AdminRequestRecordOutcome = 'pending' | 'success' | 'non_billable' | 'error' | 'cancelled' | 'unknown'
export type AdminRequestLogSort = 'created_at_desc' | 'duration_desc'

export interface AdminRequestLogSummary {
  total_requests: number
  success_requests: number
  error_requests: number
  error_rate: number
}

export interface AdminRequestLogIdentity {
  id: number
  name?: string
  email?: string
  username?: string
}

export interface AdminRequestLogUpstreamAttempt {
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

export interface AdminRequestLog {
  kind: AdminRequestLogKind
  outcome?: AdminRequestRecordOutcome
  billable?: boolean
  created_at: string
  updated_at?: string
  completed_at?: string | null
  request_id: string
  client_request_id?: string
  session_id?: string
  session_source?: string
  client_session_id?: string
  status_code: number
  platform: string
  model: string
  requested_model?: string | null
  upstream_model?: string | null
  billing_model?: string | null
  model_mapping_chain?: string | null
  billing_tier?: string | null
  service_tier?: string | null
  reasoning_effort?: string | null
  prompt_cache_key?: string | null
  request_type?: UsageRequestType
  stream: boolean
  duration_ms?: number | null
  first_token_ms?: number | null

  input_tokens?: number | null
  output_tokens?: number | null
  cache_creation_tokens?: number | null
  cache_read_tokens?: number | null
  cache_creation_5m_tokens?: number | null
  cache_creation_1h_tokens?: number | null
  image_output_tokens?: number | null
  image_output_cost?: number | null
  image_count?: number | null
  image_size?: string | null
  image_input_size?: string | null
  image_output_size?: string | null
  image_size_source?: ImageSizeSource | null
  image_size_breakdown?: ImageSizeBreakdown | null
  input_cost?: number | null
  output_cost?: number | null
  cache_creation_cost?: number | null
  cache_read_cost?: number | null
  total_cost?: number | null
  actual_cost?: number | null
  rate_multiplier?: number | null
  billing_type?: number | null
  billing_mode?: string | null
  account_rate_multiplier?: number | null
  account_stats_cost?: number | null
  cache_ttl_overridden?: boolean | null

  user_id?: number | null
  user?: AdminRequestLogIdentity | null
  api_key_id?: number | null
  api_key?: AdminRequestLogIdentity | null
  account_id?: number | null
  account?: AdminRequestLogIdentity | null
  group_id?: number | null
  group?: AdminRequestLogIdentity | null

  inbound_endpoint?: string
  upstream_endpoint?: string
  ip_address?: string
  user_agent?: string

  error_id?: number | null
  error_message?: string
  upstream_attempts?: AdminRequestLogUpstreamAttempt[]
  phase?: string
  severity?: string
  message?: string
}

export interface AdminRequestLogQueryParams {
  page?: number
  page_size?: number
  time_range?: 'today' | 'this_week' | 'last_7_days' | 'this_month' | 'last_30_days'
  start_time?: string
  end_time?: string
  kind?: AdminRequestLogKindFilter
  status_codes?: string
  status_codes_other?: boolean
  platform?: string
  model?: string
  request_type?: UsageRequestType
  stream?: boolean
  user_id?: number
  api_key_id?: number
  account_id?: number
  group_id?: number
  request_id?: string
  client_request_id?: string
  session_id?: string
  billable?: boolean
  q?: string
  sort?: AdminRequestLogSort
}

export interface AdminRequestLogResponse extends PaginatedResponse<AdminRequestLog> {
  summary: AdminRequestLogSummary
}

type BackendRequestType = UsageRequestType | 0 | 1 | 2 | 3 | null | undefined

interface AdminRequestRecord extends Omit<AdminRequestLog, 'kind' | 'request_type'> {
  outcome?: AdminRequestRecordOutcome
  request_type?: BackendRequestType
}

interface AdminRequestRecordResponse extends PaginatedResponse<AdminRequestRecord> {
  summary?: AdminRequestLogResponse['summary']
}

const REQUEST_TYPE_BY_CODE: Record<number, UsageRequestType> = {
  0: 'unknown',
  1: 'sync',
  2: 'stream',
  3: 'ws_v2'
}

function requestTypeFromBackend(value: BackendRequestType, stream?: boolean): UsageRequestType {
  if (typeof value === 'string') return value
  if (typeof value === 'number') return REQUEST_TYPE_BY_CODE[value] || 'unknown'
  return stream ? 'stream' : 'sync'
}

function kindFromOutcome(outcome?: AdminRequestRecordOutcome): AdminRequestLog['kind'] {
  if (outcome === 'success' || outcome === 'non_billable') return 'success'
  if (outcome === 'pending') return 'pending'
  return 'error'
}

function mapRecord(record: AdminRequestRecord): AdminRequestLog {
  const outcome = record.outcome || (record.status_code && record.status_code >= 400 ? 'error' : 'success')
  const billable = Boolean(record.billable)
  return {
    ...record,
    kind: kindFromOutcome(outcome),
    outcome,
    billable,
    request_type: requestTypeFromBackend(record.request_type, record.stream),
    actual_cost: billable ? record.actual_cost : null,
    total_cost: billable ? record.total_cost : null,
    input_cost: billable ? record.input_cost : null,
    output_cost: billable ? record.output_cost : null,
    cache_creation_cost: billable ? record.cache_creation_cost : null,
    cache_read_cost: billable ? record.cache_read_cost : null,
    image_output_cost: billable ? record.image_output_cost : null,
    account_stats_cost: billable ? record.account_stats_cost : null
  }
}

function buildSummary(result: AdminRequestRecordResponse): AdminRequestLogResponse['summary'] {
  if (result.summary) return result.summary
  const items = result.items || []
  const errorRequests = items.filter((item) => kindFromOutcome(item.outcome) === 'error').length
  const successRequests = items.filter((item) => kindFromOutcome(item.outcome) === 'success').length
  const totalRequests = result.total || items.length
  return {
    total_requests: totalRequests,
    success_requests: successRequests,
    error_requests: errorRequests,
    error_rate: totalRequests > 0 ? errorRequests / totalRequests : 0
  }
}

function toRecordParams(params: AdminRequestLogQueryParams): Record<string, unknown> {
  const out: Record<string, unknown> = { ...params }
  delete out.kind
  delete out.status_codes
  delete out.status_codes_other
  delete out.time_range

  if (params.kind && params.kind !== 'all') {
    out.outcome = params.kind
  }
  if (params.status_codes && /^\d+$/.test(params.status_codes.trim())) {
    out.status_code = Number(params.status_codes.trim())
  }
  if (params.status_codes_other) {
    out.status_codes_other = true
  }
  if (params.time_range) {
    out.time_range = params.time_range
  }
  return out
}

export async function list(
  params: AdminRequestLogQueryParams,
  options?: { signal?: AbortSignal }
): Promise<AdminRequestLogResponse> {
  const { data } = await apiClient.get<AdminRequestRecordResponse>('/admin/request-records', {
    params: toRecordParams(params),
    signal: options?.signal
  })
  return {
    items: (data.items || []).map(mapRecord),
    total: data.total || 0,
    page: data.page || params.page || 1,
    page_size: data.page_size || params.page_size || 20,
    pages: data.pages || 1,
    summary: buildSummary(data)
  }
}

export const requestRecordsAPI = {
  list
}

export default requestRecordsAPI
