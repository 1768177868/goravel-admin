/**
 * @vitest-environment happy-dom
 */
import { describe, it, expect } from 'vitest'
import { normalizeEntity, normalizeTreeList, normalizeListResponse } from '../normalize'

describe('normalize', () => {
  describe('normalizeEntity', () => {
    it('should unify PascalCase fields to snake_case', () => {
      const entity = { ID: 1, Name: 'test', Status: 1 }
      const result = normalizeEntity(entity)
      expect(result.id).toBe(1)
      expect(result.name).toBe('test')
      expect(result.status).toBe(1)
    })

    it('should normalize nested children', () => {
      const entity = {
        ID: 1,
        Name: 'root',
        children: [{ ID: 2, Name: 'child' }]
      }
      const result = normalizeEntity(entity)
      expect(result.children[0].id).toBe(2)
      expect(result.children[0].name).toBe('child')
    })
  })

  describe('normalizeTreeList', () => {
    it('should normalize array items', () => {
      const list = [{ ID: 1, Name: 'a' }]
      const result = normalizeTreeList(list)
      expect(result[0].id).toBe(1)
      expect(result[0].name).toBe('a')
    })
  })

  describe('normalizeListResponse', () => {
    it('should normalize list in response data', () => {
      const res = {
        code: 200,
        data: {
          list: [{ ID: 1, Name: 'item' }],
          total: 1
        }
      }
      const result = normalizeListResponse(res)
      expect(result.data.list[0].id).toBe(1)
      expect(result.data.list[0].name).toBe('item')
    })
  })
})
