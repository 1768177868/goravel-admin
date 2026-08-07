import { useState } from 'react'
import { App, Space, Switch, Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useTranslation } from 'react-i18next'
import { deleteArticle, getArticleList, updateArticle } from '@/api/article'
import { useListPage } from '@/hooks/useListPage'
import { handlePaginatedTableChange } from '@/utils/tableChange'
import { useCrudActions } from '@/hooks/useCrudActions'
import { usePermission } from '@/hooks/usePermission'
import { useUnhandledError } from '@/hooks/useUnhandledError'
import PageContainer from '@/components/PageContainer'
import SearchForm from '@/components/SearchForm'
import PermissionButton from '@/components/PermissionButton'
import logger from '@/utils/logger'
import { extractTextFromMarkdown } from '@/utils/markdown'
import ArticleFormModal from './ArticleFormModal'
import { articleInitialSearchForm, getAdminDisplayName, transformArticleRow, type ArticleRow } from './article.config'

export default function ArticleList() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const { getButtonState } = usePermission()
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<string | number | null>(null)

  const {
    tableData,
    loading,
    pagination,
    searchForm,
    onSearchFormChange,
    loadData,
    handleSearch,
    handleReset,
    handleSortChange,
    refresh,
  } = useListPage<ArticleRow>({
    fetchApi: getArticleList,
    initialSearchForm: articleInitialSearchForm,
    normalizeRows: false,
    transformData: (row) => transformArticleRow(row),
  })

  const { toolbar, confirmDelete } = useCrudActions({
    createPermission: 'article.store',
    onRefresh: refresh,
    onCreate: () => {
      setEditId(null)
      setOpen(true)
    },
    deleteApi: deleteArticle,
  })

  const handleStatusChange = async (row: ArticleRow, checked: boolean) => {
    try {
      await updateArticle(row.id, { status: checked ? 1 : 0 })
      message.success(t('common.update_success'))
      await refresh()
    } catch (error) {
      logger.error('Status change error:', error)
      showError(error, t('common.operation_failed'))
    }
  }

  const columns: ColumnsType<ArticleRow> = [
    { title: t('table.id'), dataIndex: 'id', width: 80, sorter: true },
    {
      title: t('admin_id'),
      dataIndex: 'admin_id',
      width: 140,
      render: (_, row) => getAdminDisplayName(row),
    },
    { title: t('title'), dataIndex: 'title', ellipsis: true },
    {
      title: t('content'),
      dataIndex: 'content',
      ellipsis: true,
      render: (content: string) => extractTextFromMarkdown(content).slice(0, 120),
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      width: 100,
      render: (status: number, row) => (
        <Switch
          checked={Number(status ?? 1) === 1}
          disabled={getButtonState('article.update').disabled}
          onChange={(checked) => void handleStatusChange(row, checked)}
        />
      ),
    },
    { title: t('table.updated_at'), dataIndex: 'updated_at', width: 180, sorter: true },
    { title: t('table.created_at'), dataIndex: 'created_at', width: 180, sorter: true },
    {
      title: t('common.operation'),
      key: 'operation',
      width: 160,
      fixed: 'end',
      render: (_, row) => (
        <Space>
          {getButtonState('article.update').show && (
            <PermissionButton
              permission="article.update"
              type="link"
              onClick={() => {
                setEditId(row.id)
                setOpen(true)
              }}
            >
              {t('common.edit')}
            </PermissionButton>
          )}
          {getButtonState('article.destroy').show && (
            <PermissionButton
              permission="article.destroy"
              type="link"
              danger
              onClick={() => confirmDelete(row.id, row.title)}
            >
              {t('common.delete')}
            </PermissionButton>
          )}
        </Space>
      ),
    },
  ]

  return (
    <PageContainer title={t('menu.article')} extra={toolbar}>
      <SearchForm
        fields={[
          { name: 'admin_id', label: t('admin_id') },
          { name: 'title', label: t('title') },
          { name: 'content', label: t('content'), advanced: true },
          {
            name: 'status',
            label: t('common.status'),
            type: 'select',
            advanced: true,
            options: [
              { label: t('common.enabled'), value: 1 },
              { label: t('common.disabled'), value: 0 },
            ],
          },
        ]}
        values={searchForm}
        onChange={onSearchFormChange}
        onSearch={handleSearch}
        onReset={handleReset}
      />
      <Table<ArticleRow>
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={tableData}
        scroll={{ x: 1200 }}
        pagination={{
          current: pagination.page,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          showTotal: (total) => t('common.total', { total }),
        }}
        onChange={(pager, _f, sorter) =>
          handlePaginatedTableChange({ pager, sorter, pagination, loadData, handleSortChange })
        }
      />
      <ArticleFormModal
        open={open}
        editId={editId}
        onClose={() => setOpen(false)}
        onSuccess={() => void refresh()}
      />
    </PageContainer>
  )
}
