import { describe, expect, it } from 'vitest'
import { buildRangeSearchParams } from '@/utils/listPageHelpers'
import { extractListFromResponse } from '@/utils/extractListFromResponse'

describe('buildRangeSearchParams', () => {
  it('maps range fields to _start/_end params', () => {
    const params = buildRangeSearchParams(
      {
        title: 'hello',
        created_at: ['2026-07-12 17:00:00', '2026-08-12 00:00:00']
      },
      { page: 1, page_size: 10 },
      ['created_at']
    )

    expect(params).toEqual({
      page: 1,
      page_size: 10,
      title: 'hello',
      created_at_start: '2026-07-12 17:00:00',
      created_at_end: '2026-08-12 00:00:00'
    })
  })

  it('removes empty range boundaries', () => {
    const params = buildRangeSearchParams(
      { updated_at: ['2026-07-01 00:00:00', ''] },
      {},
      ['updated_at']
    )

    expect(params.updated_at_start).toBe('2026-07-01 00:00:00')
    expect(params.updated_at_end).toBeUndefined()
    expect(params.updated_at).toBeUndefined()
  })
})

describe('extractListFromResponse', () => {
  it('reads list, data, and menus arrays', () => {
    expect(extractListFromResponse({ data: { list: [{ id: 1 }] } })).toEqual([{ id: 1 }])
    expect(extractListFromResponse({ data: { menus: [{ id: 2 }] } })).toEqual([{ id: 2 }])
    expect(extractListFromResponse({ data: { data: [{ id: 3 }] } })).toEqual([{ id: 3 }])
  })
})
