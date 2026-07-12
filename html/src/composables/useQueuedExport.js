import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import logger from '@/utils/logger'
import ErrorHandler from '@/utils/errorHandler'

/**
 * Export that queues a background job and redirects to the exports page.
 */
export function useQueuedExport(options = {}) {
  const {
    exportApi,
    getParams = () => ({}),
    redirectPath = '/exports',
    queuedMessageKey = 'common.queued',
    failedMessageKey = 'common.output_failed',
    requireExportId = true
  } = options

  const router = useRouter()
  const { t } = useI18n()
  const isExporting = ref(false)

  const handleExport = async () => {
    if (isExporting.value) {
      return
    }

    isExporting.value = true
    try {
      const params = typeof getParams === 'function' ? getParams() : getParams
      const response = await exportApi(params)
      const exportId = response.data?.export_id || response.data?.data?.export_id

      if (requireExportId && !exportId) {
        ElMessage.error(t(failedMessageKey))
        return
      }

      ElMessage.success(t(queuedMessageKey) || response.data?.message)
      router.push(redirectPath)
    } catch (error) {
      logger.error('Queued export error:', error)
      if (error.response?.status === 429) {
        ElMessage.warning(t('common.already_queued'))
      } else if (!error.__handled) {
        ErrorHandler.handle(error, { silent: true })
      }
    } finally {
      isExporting.value = false
    }
  }

  return {
    isExporting,
    handleExport
  }
}
