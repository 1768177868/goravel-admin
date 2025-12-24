import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import logger from '../utils/logger'

/**
 * CRUD 操作通用 composable
 * @param {Object} options 配置选项
 * @param {Function} options.deleteApi - 删除 API 函数
 * @param {String} options.deleteConfirmKey - 删除确认提示的 i18n key
 * @param {String} options.deleteSuccessKey - 删除成功提示的 i18n key
 * @param {String} options.tipKey - 提示标题的 i18n key（默认 'form.tip'）
 * @param {Function} options.onDeleteSuccess - 删除成功后的回调（可选）
 * @param {Function} options.onDeleteError - 删除失败后的回调（可选）
 * @param {Function} options.beforeDelete - 删除前的钩子函数（可选，返回 false 可取消删除）
 * @param {Function} options.afterDelete - 删除后的钩子函数（可选）
 * @returns {Object} 返回 CRUD 相关的状态和方法
 */
export function useCrud(options = {}) {
  const {
    deleteApi = null,
    deleteConfirmKey = '',
    deleteSuccessKey = '',
    tipKey = 'form.tip',
    onDeleteSuccess = null,
    onDeleteError = null,
    beforeDelete = null,
    afterDelete = null
  } = options

  const { t } = useI18n()

  // 对话框显示状态
  const dialogVisible = ref(false)
  // 编辑ID
  const editId = ref(null)

  /**
   * 打开添加对话框
   */
  const handleAdd = () => {
    editId.value = null
    dialogVisible.value = true
  }

  /**
   * 打开编辑对话框
   * @param {Object|Number|String} rowOrId - 行数据或ID
   */
  const handleEdit = (rowOrId) => {
    if (typeof rowOrId === 'object' && rowOrId !== null) {
      editId.value = rowOrId.id || rowOrId.ID
    } else {
      editId.value = rowOrId
    }
    dialogVisible.value = true
  }

  /**
   * 关闭对话框
   */
  const handleClose = () => {
    dialogVisible.value = false
    editId.value = null
  }

  /**
   * 表单提交成功后的处理
   * @param {Function} reloadData - 重新加载数据的函数
   */
  const handleFormSuccess = (reloadData = null) => {
    handleClose()
    if (reloadData && typeof reloadData === 'function') {
      reloadData()
    }
  }

  /**
   * 删除操作
   * @param {Object|Number|String} rowOrId - 行数据或ID
   * @param {Function} reloadData - 重新加载数据的函数
   */
  const handleDelete = async (rowOrId, reloadData = null) => {
    try {
      // 获取ID
      let id
      if (typeof rowOrId === 'object' && rowOrId !== null) {
        id = rowOrId.id || rowOrId.ID
      } else {
        id = rowOrId
      }

      if (!id) {
        logger.error('useCrud: delete id is required')
        return
      }

      // 删除前的钩子
      if (beforeDelete) {
        const result = await beforeDelete(rowOrId, id)
        if (result === false) {
          return // 取消删除
        }
      }

      // 确认删除
      const confirmKey = deleteConfirmKey || 'common.delete_confirm'
      await ElMessageBox.confirm(
        t(confirmKey),
        t(tipKey),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )

      // 执行删除
      if (!deleteApi) {
        logger.error('useCrud: deleteApi is required')
        return
      }

      await deleteApi(id)

      // 删除成功提示
      const successKey = deleteSuccessKey || 'common.delete_success'
      ElMessage.success(t(successKey))

      // 删除成功回调
      if (onDeleteSuccess) {
        onDeleteSuccess(rowOrId, id)
      }

      // 删除后钩子
      if (afterDelete) {
        afterDelete(rowOrId, id)
      }

      // 重新加载数据
      if (reloadData && typeof reloadData === 'function') {
        reloadData()
      }
    } catch (error) {
      if (error === 'cancel') {
        return // 用户取消
      }

      logger.error('Delete error:', error)

      // 删除失败回调
      if (onDeleteError) {
        onDeleteError(error, rowOrId)
      }
    }
  }

  /**
   * 批量删除
   * @param {Array} rows - 要删除的行数据数组
   * @param {Function} reloadData - 重新加载数据的函数
   */
  const handleBatchDelete = async (rows, reloadData = null) => {
    if (!rows || rows.length === 0) {
      return
    }

    try {
      const confirmKey = deleteConfirmKey || 'common.batch_delete_confirm'
      const count = rows.length
      
      await ElMessageBox.confirm(
        t(confirmKey, { count }),
        t(tipKey),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )

      if (!deleteApi) {
        logger.error('useCrud: deleteApi is required for batch delete')
        return
      }

      // 批量删除（假设 API 支持批量删除）
      const ids = rows.map(row => row.id || row.ID).filter(Boolean)
      if (ids.length === 0) {
        return
      }

      // 这里需要根据实际的批量删除 API 调整
      // 假设 deleteApi 可以接受数组或需要循环调用
      if (Array.isArray(ids) && ids.length > 0) {
        // 如果 API 支持批量删除，直接传递数组
        // 否则循环调用
        for (const id of ids) {
          await deleteApi(id)
        }
      }

      const successKey = deleteSuccessKey || 'common.delete_success'
      ElMessage.success(t(successKey))

      if (onDeleteSuccess) {
        onDeleteSuccess(rows, ids)
      }

      if (reloadData && typeof reloadData === 'function') {
        reloadData()
      }
    } catch (error) {
      if (error === 'cancel') {
        return
      }

      logger.error('Batch delete error:', error)

      if (onDeleteError) {
        onDeleteError(error, rows)
      }
    }
  }

  return {
    // 状态
    dialogVisible,
    editId,

    // 方法
    handleAdd,
    handleEdit,
    handleClose,
    handleFormSuccess,
    handleDelete,
    handleBatchDelete
  }
}

