import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get } = vi.hoisted(() => ({
  get: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
  },
}))

import requestRecordsAPI from '@/api/admin/requestRecords'

describe('admin request records api', () => {
  beforeEach(() => {
    get.mockReset()
  })

  it('maps request records into the existing request logs table contract', async () => {
    get.mockResolvedValueOnce({
      data: {
        items: [
          {
            outcome: 'success',
            billable: true,
            request_type: 2,
            stream: true,
            request_id: 'req-1',
            created_at: '2026-05-29T00:00:00Z',
            status_code: 200,
            model: 'gpt-5.4-mini',
            actual_cost: 0.01,
            upstream_attempts: [
              {
                account_id: 20,
                account_name: 'L站福利1',
                upstream_status_code: 401,
                message: 'Invalid API key',
              },
            ],
          },
          {
            outcome: 'error',
            billable: false,
            request_type: 1,
            stream: false,
            request_id: 'req-2',
            created_at: '2026-05-29T00:01:00Z',
            status_code: 503,
            model: 'gpt-5.4-mini',
            actual_cost: 0,
            total_cost: 0,
            input_cost: 0.01,
            output_cost: 0.02,
            cache_creation_cost: 0.03,
            cache_read_cost: 0.04,
            image_output_cost: 0.05,
            account_stats_cost: 0.06,
          },
          {
            outcome: 'pending',
            billable: false,
            request_type: 1,
            stream: false,
            request_id: 'req-3',
            created_at: '2026-05-29T00:02:00Z',
            status_code: 0,
            model: 'gpt-5.4-mini',
            actual_cost: null,
          },
        ],
        total: 3,
        page: 1,
        page_size: 20,
        pages: 1,
      },
    })

    const result = await requestRecordsAPI.list({ page: 1, page_size: 20, kind: 'error', status_codes: '503' })

    expect(get).toHaveBeenCalledWith('/admin/request-records', {
      params: expect.objectContaining({
        page: 1,
        page_size: 20,
        outcome: 'error',
        status_code: 503,
      }),
      signal: undefined,
    })
    expect(result.items[0]).toEqual(expect.objectContaining({
      kind: 'success',
      outcome: 'success',
      request_type: 'stream',
      billable: true,
      actual_cost: 0.01,
      upstream_attempts: [
        expect.objectContaining({
          account_id: 20,
          upstream_status_code: 401,
          message: 'Invalid API key',
        }),
      ],
    }))
    expect(result.items[1]).toEqual(expect.objectContaining({
      kind: 'error',
      outcome: 'error',
      request_type: 'sync',
      billable: false,
      actual_cost: null,
      total_cost: null,
      input_cost: null,
      output_cost: null,
      cache_creation_cost: null,
      cache_read_cost: null,
      image_output_cost: null,
      account_stats_cost: null,
    }))
    expect(result.items[2]).toEqual(expect.objectContaining({
      kind: 'pending',
      outcome: 'pending',
      request_type: 'sync',
      billable: false,
      actual_cost: null,
    }))
    expect(result.summary).toEqual({
      total_requests: 3,
      success_requests: 1,
      error_requests: 1,
      error_rate: 1 / 3,
    })
  })

  it('does not fall back to aggregate request logs when records endpoint fails', async () => {
    get.mockRejectedValueOnce({ status: 404, message: 'not found' })

    await expect(requestRecordsAPI.list({ page: 1, page_size: 20 })).rejects.toEqual({
      status: 404,
      message: 'not found',
    })

    expect(get).toHaveBeenCalledTimes(1)
    expect(get).toHaveBeenCalledWith('/admin/request-records', expect.any(Object))
  })
})
