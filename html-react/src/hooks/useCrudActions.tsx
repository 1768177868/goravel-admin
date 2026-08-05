import { useCallback, useState } from 'react'
import { App, Button, Space } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import PermissionButton from '@/components/PermissionButton'
import { useUnhandledError } from '@/hooks/useUnhandledError'

interface UseCrudActionsOptions {
  createPermission?: string
  deletePermission?: string
  onRefresh: () => Promise<void> | void
  onCreate?: () => void
  deleteApi?: (id: string | number) => Promise<unknown>
}

export function useCrudActions(options: UseCrudActionsOptions) {
  const { t } = useTranslation()
  const { modal, message } = App.useApp()
  const showError = useUnhandledError()
  const [deleting, setDeleting] = useState(false)

  const confirmDelete = useCallback(
    (id: string | number, name?: string) => {
      if (!options.deleteApi) return
      modal.confirm({
        title: t('common.delete_confirm'),
        content: name ? `${name}` : undefined,
        okType: 'danger',
        onOk: async () => {
          setDeleting(true)
          try {
            await options.deleteApi?.(id)
            message.success(t('common.delete_success'))
            await options.onRefresh()
          } catch (error) {
            showError(error, t('common.operation_failed'))
          } finally {
            setDeleting(false)
          }
        },
      })
    },
    [modal, message, options, showError, t],
  )

  const toolbar = (
    <Space>
      {options.onCreate && (
        <PermissionButton
          permission={options.createPermission || ''}
          type="primary"
          icon={<PlusOutlined />}
          onClick={options.onCreate}
        >
          {t('common.add')}
        </PermissionButton>
      )}
      <Button icon={<ReloadOutlined />} onClick={() => void options.onRefresh()}>
        {t('common.refresh')}
      </Button>
    </Space>
  )

  return { toolbar, confirmDelete, deleting }
}
