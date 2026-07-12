import { describe, it, expect, vi } from 'vitest'
import { getOperationTitle, createOperationTitleTranslator } from '../operationTitle'

describe('operationTitle', () => {
  const t = vi.fn((key) => `t:${key}`)
  const te = vi.fn((key) => key === 'permission.dictionary.update')
  const tm = vi.fn((namespace) => {
    if (namespace === 'permission') {
      return { 'role.store': 'Create Role' }
    }
    return null
  })

  it('returns dash for empty title', () => {
    expect(getOperationTitle(t, te, tm, '')).toBe('-')
  })

  it('translates permission slug via te()', () => {
    expect(getOperationTitle(t, te, tm, 'dictionaries.update')).toBe('t:permission.dictionary.update')
  })

  it('translates flat permission keys via tm()', () => {
    expect(getOperationTitle(t, vi.fn(() => false), tm, 'roles.store')).toBe('Create Role')
  })

  it('maps legacy title keys before translation', () => {
    expect(getOperationTitle(t, vi.fn(() => false), tm, 'online-admin.kick-out')).toBe(
      'online_admin.kick_out'
    )
  })

  it('createOperationTitleTranslator wraps i18n helpers', () => {
    const translate = createOperationTitleTranslator({ t, te, tm })
    expect(translate('dictionaries.update')).toBe('t:permission.dictionary.update')
  })
})
