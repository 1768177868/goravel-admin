import { useEffect, useState } from 'react'
import { App, Checkbox, InputNumber, Modal } from 'antd'
import { useTranslation } from 'react-i18next'

export const CLEAN_ALL_DAYS = 36500

interface LogCleanModalProps {
  open: boolean
  loading?: boolean
  onClose: () => void
  onConfirm: (days: number) => Promise<void>
}

export default function LogCleanModal({ open, loading, onClose, onConfirm }: LogCleanModalProps) {
  const { t } = useTranslation()
  const { modal } = App.useApp()
  const [days, setDays] = useState(30)
  const [cleanAll, setCleanAll] = useState(false)

  useEffect(() => {
    if (open) {
      setDays(30)
      setCleanAll(false)
    }
  }, [open])

  const handleOk = () => {
    const effectiveDays = cleanAll ? CLEAN_ALL_DAYS : days
    modal.confirm({
      title: t('log.clean'),
      content: cleanAll
        ? t('log.clean_confirm')
        : t('log.clean_days_confirm', { days: effectiveDays }),
      okType: 'danger',
      onOk: async () => {
        await onConfirm(effectiveDays)
      },
    })
  }

  return (
    <Modal
      open={open}
      title={t('log.clean')}
      confirmLoading={loading}
      onCancel={onClose}
      onOk={handleOk}
      destroyOnClose
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div>
          <div style={{ marginBottom: 8 }}>{t('log.clean_days_label')}</div>
          <InputNumber
            min={1}
            max={36500}
            value={days}
            disabled={cleanAll}
            style={{ width: '100%' }}
            onChange={(value) => setDays(typeof value === 'number' ? value : 30)}
          />
          <div style={{ marginTop: 8, color: 'var(--ant-color-text-secondary)', fontSize: 12 }}>
            {t('log.clean_days_hint')}
          </div>
        </div>
        <Checkbox checked={cleanAll} onChange={(e) => setCleanAll(e.target.checked)}>
          {t('log.clean_all')}
        </Checkbox>
      </div>
    </Modal>
  )
}
