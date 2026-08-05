import { useState } from 'react'
import { App } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { ApiResponse } from '@/types'

interface QueuedExportOptions {
  exportApi: (params: Record<string, unknown>) => Promise<ApiResponse<unknown>>
  getParams?: () => Record<string, unknown>
  redirectPath?: string
  queuedMessageKey?: string
  failedMessageKey?: string
  requireExportId?: boolean
}

export function useQueuedExport(options: QueuedExportOptions) {
  const {
    exportApi,
    getParams = () => ({}),
    redirectPath = '/exports',
    queuedMessageKey = 'common.queued',
    failedMessageKey = 'common.output_failed',
    requireExportId = true,
  } = options

  const navigate = useNavigate()
  const { t } = useTranslation()
  const { message } = App.useApp()
  const [exporting, setExporting] = useState(false)

  const handleExport = async () => {
    if (exporting) return
    setExporting(true)
    try {
      const response = await exportApi(getParams())
      const data = (response.data || {}) as Record<string, unknown>
      const nested = (data.data || {}) as Record<string, unknown>
      const exportId = data.export_id || nested.export_id

      if (requireExportId && !exportId) {
        message.error(t(failedMessageKey, { defaultValue: t('common.operation_failed') }))
        return
      }

      message.success(t(queuedMessageKey))
      navigate(redirectPath)
    } catch (error) {
      const err = error as { response?: { status?: number }; __handled?: boolean }
      if (err.response?.status === 429) {
        message.warning(t('common.already_queued', { defaultValue: '已有导出任务在队列中' }))
      } else if (!err.__handled) {
        message.error(t('common.operation_failed'))
      }
    } finally {
      setExporting(false)
    }
  }

  return { exporting, handleExport }
}
