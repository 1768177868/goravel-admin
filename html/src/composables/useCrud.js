import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import logger from '../utils/logger'

/**
 * @typedef {import('../types/composables').UseCrudOptions} UseCrudOptions
 * @typedef {import('../types/composables').UseCrudReturn} UseCrudReturn
 */

/**
 * CRUD 操作通用 composable
 * 
 * @description 提供通用的增删改查操作，支持按需使用
 * 
 * @example
 * // 1. 只需要添加/编辑功能（不需要删除）
 * const { dialogVisible, editId, handleAdd, handleEdit, handleClose, handleFormSuccess } = useCrud()
 * 
 * // 2. 需要删除功能
 * const { handleDelete } = useCrud({ deleteApi: deleteXxx })
 * 
 * // 3. 需要批量删除功能（使用专门的批量删除 API）
 * const { handleBatchDelete } = useCrud({ batchDeleteApi: batchDeleteXxx })
 * 
 * // 4. 完整功能
 * const { handleDelete, handleBatchDelete } = useCrud({ deleteApi, batchDeleteApi })
 * 
 * @param {UseCrudOptions} [options={}] 配置选项（所有参数都是可选的）
 * @returns {UseCrudReturn} 返回 CRUD 相关的状态和方法
 */
export function useCrud(options = {}) {
  const {
    deleteApi = null,
    batchDeleteApi = null,
    deleteConfirmKey = '',
    deleteSuccessKey = '',
    batchDeleteConfirmKey = '',
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
      ElMessage.warning(t('common.please_select_items'))
      return
    }

    try {
      const confirmKey = batchDeleteConfirmKey || 'common.batch_delete_confirm'
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

      const ids = rows.map(row => row.id || row.ID).filter(Boolean)
      if (ids.length === 0) {
        return
      }

      // 优先使用批量删除 API，否则循环调用单个删除 API
      if (batchDeleteApi) {
        await batchDeleteApi(ids)
      } else if (deleteApi) {
        for (const id of ids) {
          await deleteApi(id)
        }
      } else {
        logger.error('useCrud: deleteApi or batchDeleteApi is required for batch delete')
        return
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

