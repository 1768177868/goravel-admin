import { useEffect, useMemo, useState } from 'react'
import { Button, Checkbox, Modal, Space, Tooltip } from 'antd'
import { ArrowDownOutlined, ArrowUpOutlined, VerticalLeftOutlined, VerticalRightOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnSettingItem, ColumnSettingValue } from '@/hooks/useColumnSetting'

interface ColumnSettingDialogProps {
  open: boolean
  onClose: () => void
  allColumns: ColumnSettingItem[]
  visibleColumns: string[]
  columnOrder: string[]
  fixedColumns: Record<string, 'left' | 'right'>
  onConfirm: (value: ColumnSettingValue) => void
}

export default function ColumnSettingDialog({
  open,
  onClose,
  allColumns,
  visibleColumns,
  columnOrder,
  fixedColumns,
  onConfirm,
}: ColumnSettingDialogProps) {
  const { t } = useTranslation()
  const [localVisible, setLocalVisible] = useState<string[]>(visibleColumns)
  const [localOrder, setLocalOrder] = useState<string[]>(columnOrder)
  const [localFixed, setLocalFixed] = useState<Record<string, 'left' | 'right'>>(fixedColumns)

  useEffect(() => {
    if (!open) return
    setLocalVisible(visibleColumns)
    setLocalOrder(columnOrder.length ? columnOrder : allColumns.map((c) => c.key))
    setLocalFixed(fixedColumns)
  }, [open, visibleColumns, columnOrder, fixedColumns, allColumns])

  const columnMap = useMemo(() => new Map(allColumns.map((c) => [c.key, c])), [allColumns])
  const orderedItems = useMemo(
    () => localOrder.map((key) => columnMap.get(key)).filter((item): item is ColumnSettingItem => !!item),
    [localOrder, columnMap],
  )
  const toggleableItems = useMemo(() => orderedItems.filter((item) => !item.required), [orderedItems])
  const allChecked = toggleableItems.length > 0 && toggleableItems.every((item) => localVisible.includes(item.key))
  const indeterminate = !allChecked && toggleableItems.some((item) => localVisible.includes(item.key))

  const handleSelectAll = (checked: boolean) => {
    const toggleKeys = new Set(toggleableItems.map((item) => item.key))
    setLocalVisible((prev) =>
      checked ? Array.from(new Set([...prev, ...toggleKeys])) : prev.filter((key) => !toggleKeys.has(key)),
    )
  }

  const toggleColumn = (key: string, checked: boolean) => {
    setLocalVisible((prev) => (checked ? Array.from(new Set([...prev, key])) : prev.filter((k) => k !== key)))
  }

  const moveItem = (index: number, direction: -1 | 1) => {
    const target = index + direction
    if (target < 0 || target >= localOrder.length) return
    const next = [...localOrder]
    ;[next[index], next[target]] = [next[target], next[index]]
    setLocalOrder(next)
  }

  const toggleFixed = (key: string, side: 'left' | 'right') => {
    setLocalFixed((prev) => {
      const next = { ...prev }
      if (next[key] === side) delete next[key]
      else next[key] = side
      return next
    })
  }

  const handleReset = () => {
    const keys = allColumns.map((c) => c.key)
    setLocalVisible(keys)
    setLocalOrder(keys)
    setLocalFixed({})
  }

  const handleOk = () => {
    onConfirm({ visibleColumns: localVisible, columnOrder: localOrder, fixedColumns: localFixed })
  }

  return (
    <Modal
      open={open}
      title={t('common.column_setting')}
      onCancel={onClose}
      destroyOnHidden
      width={420}
      footer={[
        <Button key="reset" onClick={handleReset}>
          {t('common.reset')}
        </Button>,
        <Button key="cancel" onClick={onClose}>
          {t('common.cancel')}
        </Button>,
        <Button key="confirm" type="primary" onClick={handleOk}>
          {t('common.confirm')}
        </Button>,
      ]}
    >
      <div style={{ marginBottom: 10, fontSize: 12, color: 'var(--ant-color-text-tertiary, #999)' }}>
        {t('common.column_setting_tip')}
      </div>
      <div
        style={{
          paddingBottom: 8,
          marginBottom: 4,
          borderBottom: '1px solid var(--ant-color-border-secondary, #f0f0f0)',
        }}
      >
        <Checkbox checked={allChecked} indeterminate={indeterminate} onChange={(e) => handleSelectAll(e.target.checked)}>
          {t('common.select_all')}
        </Checkbox>
      </div>
      <div style={{ maxHeight: 380, overflowY: 'auto' }}>
        {orderedItems.map((item, index) => {
          const fixed = localFixed[item.key]
          return (
            <div
              key={item.key}
              style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '6px 2px' }}
            >
              <Checkbox
                checked={localVisible.includes(item.key)}
                disabled={item.required}
                onChange={(e) => toggleColumn(item.key, e.target.checked)}
              />
              <span
                style={{
                  flex: 1,
                  fontSize: 13,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {item.title}
              </span>
              <Space size={2}>
                <Tooltip title={t('common.drag_to_sort')}>
                  <Button
                    size="small"
                    type="text"
                    icon={<ArrowUpOutlined />}
                    disabled={index === 0}
                    onClick={() => moveItem(index, -1)}
                  />
                </Tooltip>
                <Button
                  size="small"
                  type="text"
                  icon={<ArrowDownOutlined />}
                  disabled={index === orderedItems.length - 1}
                  onClick={() => moveItem(index, 1)}
                />
                <Tooltip title={fixed === 'left' ? t('common.unfreeze') : t('common.freeze_left')}>
                  <Button
                    size="small"
                    type={fixed === 'left' ? 'primary' : 'text'}
                    icon={<VerticalRightOutlined />}
                    onClick={() => toggleFixed(item.key, 'left')}
                  />
                </Tooltip>
                <Tooltip title={fixed === 'right' ? t('common.unfreeze') : t('common.freeze_right')}>
                  <Button
                    size="small"
                    type={fixed === 'right' ? 'primary' : 'text'}
                    icon={<VerticalLeftOutlined />}
                    onClick={() => toggleFixed(item.key, 'right')}
                  />
                </Tooltip>
              </Space>
            </div>
          )
        })}
      </div>
    </Modal>
  )
}
