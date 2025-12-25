import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock vue-i18n
vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key, params) => {
      const translations = {
        'common.delete_confirm': '确定要删除吗？',
        'common.batch_delete_confirm': `确定要删除选中的 ${params?.count || 0} 条数据吗?`,
        'common.delete_success': '删除成功',
        'common.update_success': '更新成功',
        'common.create_success': '创建成功',
        'common.confirm': '确定',
        'common.cancel': '取消',
        'form.tip': '提示',
        'common.please_select_items': '请选择要操作的项目'
      }
      return translations[key] || key
    }
  })
}))

// Mock element-plus
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn().mockResolvedValue(true)
  }
}))

// Mock logger
vi.mock('../../utils/logger', () => ({
  default: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn()
  }
}))

import { useCrud } from '../useCrud'
import { ElMessage, ElMessageBox } from 'element-plus'

describe('useCrud', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('对话框状态管理', () => {
    it('应该初始化 dialogVisible 为 false', () => {
      const { dialogVisible } = useCrud()
      expect(dialogVisible.value).toBe(false)
    })

    it('应该初始化 editId 为 null', () => {
      const { editId } = useCrud()
      expect(editId.value).toBe(null)
    })

    it('handleAdd 应该打开对话框并清空 editId', () => {
      const { dialogVisible, editId, handleAdd } = useCrud()
      
      editId.value = 123
      handleAdd()
      
      expect(dialogVisible.value).toBe(true)
      expect(editId.value).toBe(null)
    })

    it('handleEdit 应该打开对话框并设置 editId', () => {
      const { dialogVisible, editId, handleEdit } = useCrud()
      
      handleEdit({ id: 456 })
      
      expect(dialogVisible.value).toBe(true)
      expect(editId.value).toBe(456)
    })

    it('handleEdit 应该支持 ID 属性', () => {
      const { editId, handleEdit } = useCrud()
      
      handleEdit({ ID: 789 })
      
      expect(editId.value).toBe(789)
    })

    it('handleClose 应该关闭对话框并清空 editId', () => {
      const { dialogVisible, editId, handleClose, handleAdd } = useCrud()
      
      handleAdd()
      editId.value = 123
      handleClose()
      
      expect(dialogVisible.value).toBe(false)
      expect(editId.value).toBe(null)
    })
  })

  describe('handleFormSuccess', () => {
    it('应该关闭对话框', () => {
      const { dialogVisible, handleAdd, handleFormSuccess } = useCrud()
      
      handleAdd()
      handleFormSuccess()
      
      expect(dialogVisible.value).toBe(false)
    })

    it('应该调用 reloadData 回调', () => {
      const reloadData = vi.fn()
      const { handleAdd, handleFormSuccess } = useCrud()
      
      handleAdd()
      handleFormSuccess(reloadData)
      
      expect(reloadData).toHaveBeenCalled()
    })
  })

  describe('handleDelete', () => {
    it('应该显示确认对话框', async () => {
      const deleteApi = vi.fn().mockResolvedValue({})
      const { handleDelete } = useCrud({ deleteApi })
      
      await handleDelete({ id: 1 })
      
      expect(ElMessageBox.confirm).toHaveBeenCalled()
    })

    it('应该调用 deleteApi 并显示成功消息', async () => {
      const deleteApi = vi.fn().mockResolvedValue({})
      const { handleDelete } = useCrud({ deleteApi })
      
      await handleDelete({ id: 1 })
      
      expect(deleteApi).toHaveBeenCalledWith(1)
      expect(ElMessage.success).toHaveBeenCalled()
    })

    it('应该支持直接传递 ID', async () => {
      const deleteApi = vi.fn().mockResolvedValue({})
      const { handleDelete } = useCrud({ deleteApi })
      
      await handleDelete(123)
      
      expect(deleteApi).toHaveBeenCalledWith(123)
    })

    it('应该在删除后调用 reloadData', async () => {
      const deleteApi = vi.fn().mockResolvedValue({})
      const reloadData = vi.fn()
      const { handleDelete } = useCrud({ deleteApi })
      
      await handleDelete({ id: 1 }, reloadData)
      
      expect(reloadData).toHaveBeenCalled()
    })

    it('应该在用户取消时不执行删除', async () => {
      ElMessageBox.confirm.mockRejectedValueOnce('cancel')
      const deleteApi = vi.fn()
      const { handleDelete } = useCrud({ deleteApi })
      
      await handleDelete({ id: 1 })
      
      expect(deleteApi).not.toHaveBeenCalled()
    })
  })

  describe('handleBatchDelete', () => {
    it('应该拒绝空数组', async () => {
      const { handleBatchDelete } = useCrud()
      
      await handleBatchDelete([])
      
      expect(ElMessage.warning).toHaveBeenCalled()
      expect(ElMessageBox.confirm).not.toHaveBeenCalled()
    })

    it('应该使用 batchDeleteApi 当提供时', async () => {
      const batchDeleteApi = vi.fn().mockResolvedValue({})
      const { handleBatchDelete } = useCrud({ batchDeleteApi })
      
      await handleBatchDelete([{ id: 1 }, { id: 2 }])
      
      expect(batchDeleteApi).toHaveBeenCalledWith([1, 2])
    })

    it('应该回退到循环调用 deleteApi', async () => {
      const deleteApi = vi.fn().mockResolvedValue({})
      const { handleBatchDelete } = useCrud({ deleteApi })
      
      await handleBatchDelete([{ id: 1 }, { id: 2 }])
      
      expect(deleteApi).toHaveBeenCalledTimes(2)
      expect(deleteApi).toHaveBeenCalledWith(1)
      expect(deleteApi).toHaveBeenCalledWith(2)
    })

    it('应该在成功后调用 reloadData', async () => {
      const batchDeleteApi = vi.fn().mockResolvedValue({})
      const reloadData = vi.fn()
      const { handleBatchDelete } = useCrud({ batchDeleteApi })
      
      await handleBatchDelete([{ id: 1 }], reloadData)
      
      expect(reloadData).toHaveBeenCalled()
    })
  })

  describe('可选参数', () => {
    it('应该支持自定义 deleteConfirmKey', async () => {
      const deleteApi = vi.fn().mockResolvedValue({})
      const { handleDelete } = useCrud({
        deleteApi,
        deleteConfirmKey: 'custom.confirm'
      })
      
      await handleDelete({ id: 1 })
      
      // 验证确认框被调用
      expect(ElMessageBox.confirm).toHaveBeenCalled()
    })

    it('应该支持 onDeleteSuccess 回调', async () => {
      const deleteApi = vi.fn().mockResolvedValue({})
      const onDeleteSuccess = vi.fn()
      const { handleDelete } = useCrud({
        deleteApi,
        onDeleteSuccess
      })
      
      const row = { id: 1 }
      await handleDelete(row)
      
      expect(onDeleteSuccess).toHaveBeenCalledWith(row, 1)
    })
  })
})

