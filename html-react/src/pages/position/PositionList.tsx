import { useTranslation } from 'react-i18next'
import SimpleCrudPage from '@/components/SimpleCrudPage'
import { createPosition, deletePosition, getPositionList, updatePosition } from '@/api/position'
import StatusTag from '@/components/StatusTag'

interface PositionRow {
  id: number | string
  name?: string
  sort?: number
  status?: number
  created_at?: string
}

export default function PositionList() {
  const { t } = useTranslation()

  return (
    <SimpleCrudPage<PositionRow>
      title={t('menu.position_management')}
      permissions={{
        store: 'position.store',
        update: 'position.update',
        destroy: 'position.destroy',
      }}
      fetchApi={getPositionList}
      createApi={createPosition}
      updateApi={updatePosition}
      deleteApi={deletePosition}
      createTitle={t('position.add')}
      editTitle={t('position.edit')}
      initialSearchForm={{ name: '', status: '' }}
      searchFields={[
        { name: 'name', label: t('common.name') },
        {
          name: 'status',
          label: t('common.status'),
          type: 'select',
          options: [
            { label: t('common.enabled'), value: 1 },
            { label: t('common.disabled'), value: 0 },
          ],
        },
      ]}
      formFields={[
        { name: 'name', label: t('common.name'), required: true },
        { name: 'sort', label: t('common.sort'), type: 'number' },
        { name: 'status', label: t('common.status'), type: 'status' },
      ]}
      columns={[
        { title: t('table.id'), dataIndex: 'id', width: 80, sorter: true },
        { title: t('common.name'), dataIndex: 'name' },
        { title: t('common.sort'), dataIndex: 'sort', width: 80, sorter: true },
        {
          title: t('common.status'),
          dataIndex: 'status',
          width: 100,
          render: (status: number) => <StatusTag status={status} />,
        },
        { title: t('table.created_at'), dataIndex: 'created_at', width: 180, sorter: true },
      ]}
    />
  )
}
