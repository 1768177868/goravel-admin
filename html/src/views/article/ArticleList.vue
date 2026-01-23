<template>
  <div class="article-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.article') }}</span>
          
          <el-button 
            type="primary" 
            :disabled="getButtonState('article.store').disabled"
            @click="handleAdd"
          >
            <el-icon><PlusIcon /></el-icon>
            {{ $t('common.add') }}
          </el-button>
          
        </div>
      </template>

      <SearchForm
        :model="searchForm"
        :fields="searchFields"
        :initial-values="initialSearchForm"
        i18n-prefix="article"
        @search="handleSearch"
        @reset="handleReset"
      >
        
      </SearchForm>

      <VxeTable
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
      >
        
        <template #admin_id="{ row }">
          {{ getadminDisplayName(row.admin || row.admin_id) }}
        </template>
        <template #operation="{ row }">
          <TableActionButtons
            :row="row"
            :primary-actions="getPrimaryActions(row)"
            :more-actions="getMoreActions(row)"
            :get-button-state="getButtonState"
            @action="handleAction"
          />
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />
    </el-card>

    <ArticleForm
      ref="formRef"
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'

import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import TableActionButtons from '../../components/TableActionButtons.vue'
import ArticleForm from './ArticleForm.vue'
import { useListPage } from '../../composables/useListPage'
import { usePermission } from '../../composables/usePermission'
import { useCrud } from '../../composables/useCrud'

import {
  getArticleList,
  deleteArticle,
  updateArticle,
  
} from '../../api/article'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

const PlusIcon = markRaw(Plus)

// 权限控制
const { getButtonState } = usePermission()

const { t } = useI18n()

const tableRef = ref(null)
const formRef = ref(null)


const {
  dialogVisible,
  editId,
  handleAdd,
  handleClose,
  handleDelete: handleDeleteCrud
} = useCrud({
  deleteApi: deleteArticle
})

const initialSearchForm = {

  title: '',
  content: '',
  status: '',
  admin_id: '',
  created_at: '',
  updated_at: '',
}

const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getArticleList,
  initialSearchForm,
  fieldMapping: {},
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef)
})

const searchFields = computed(() => [

  {
    prop: 'title',
    label: t('article.title'),
    type: 'input',
    clearable: true,

    width: '200px',
    advanced: false
  },
  {
    prop: 'content',
    label: t('article.content'),
    type: 'input',
    clearable: true,

    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('article.status'),
    type: 'select',
    clearable: true,

    
    apiUrl: '/options?type=dictionary&dictionary_type=status',
    

    width: '200px',
    advanced: false
  },
  {
    prop: 'admin_id',
    label: t('article.admin_id'),
    type: 'input',
    clearable: true,

    width: '200px',
    advanced: false
  },
  {
    prop: 'created_at',
    label: t('article.created_at'),
    type: 'datetime',
    clearable: true,

    width: '200px',
    advanced: false
  },
  {
    prop: 'updated_at',
    label: t('article.updated_at'),
    type: 'datetime',
    clearable: true,

    width: '200px',
    advanced: false
  },
])

const tableColumns = computed(() => [
  {
    field: 'id',
    title: t('table.id'),
    width: 80,
    sortable: true
  },

  {
    field: 'title',
    title: t('article.title'),
    sortable: false
  },
  {
    field: 'content',
    title: t('article.content'),
    sortable: false
  },
  {
    field: 'status',
    title: t('article.status'),
    sortable: false
  },
  {
    field: 'admin_id',
    title: t('article.admin_id'),
    slot: 'admin_id',
    sortable: false
  },
  {
    field: 'created_at',
    title: t('article.created_at'),
    sortable: true
  },
  {
    field: 'updated_at',
    title: t('article.updated_at'),
    sortable: false
  },
  {
    field: 'created_at',
    title: t('table.created_at'),
    width: 180,
    sortable: true
  },
  {
    field: 'operation',
    title: t('table.operation'),
    width: 220,
    fixed: 'right',
    slot: 'operation'
  }
])


const getadminDisplayName = (admin_id) => {
  if (!admin_id) return '-'
  return admin_id.username || admin_id.admin || '-'
}





const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleDelete = (row) => handleDeleteCrud(row, loadData)

const handleFormSuccess = () => {
  handleClose()
  loadData()
}

// 获取主要操作按钮配置
const getPrimaryActions = (row) => {
  return [
    
    {
      key: 'edit',
      label: t('common.edit'),
      type: 'primary',
      permission: 'article.update',
      handler: handleEdit
    },
    
    
    {
      key: 'delete',
      label: t('common.delete'),
      type: 'danger',
      permission: 'article.destroy',
      handler: handleDelete
    }
    
  ]
}

// 获取更多操作按钮配置（可根据需要扩展）
const getMoreActions = (row) => {
  return []
}

// 处理操作事件
const handleAction = (command, row) => {
  switch (command) {
    case 'edit':
      handleEdit(row)
      break
    case 'delete':
      handleDelete(row)
      break
  }
}



onMounted(async () => {
  try {
    initDefaultSort()
    await loadData()
  } catch (error) {
    logger.error('ListPage onMounted error:', error)
    ErrorHandler.handle(error)
  }
})
</script>

<style scoped>
.article-list {
  padding: 20px;
}
</style>
