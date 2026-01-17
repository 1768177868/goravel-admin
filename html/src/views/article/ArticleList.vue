<template>
  <div class="article-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('menu.article') }}</span>
          
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
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
      />

      <VxeTable
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
      >
        <template #operation="{ row }">
          
          <el-button type="primary" size="small" @click="handleEdit(row)">
            {{ $t('common.edit') }}
          </el-button>
          
          
          <el-button type="danger" size="small" @click="handleDelete(row)">
            {{ $t('common.delete') }}
          </el-button>
          
        </template>
      </VxeTable>

      <Pagination
        v-model="pagination"
        :auto-load="true"
        :on-page-change="loadData"
      />

      <ArticleForm
        ref="formRef"
        v-model="dialogVisible"
        :edit-id="editId"
        @success="handleFormSuccess"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SearchForm from '../../components/SearchForm.vue'
import Pagination from '../../components/Pagination.vue'
import VxeTable from '../../components/VxeTable.vue'
import ArticleForm from './ArticleForm.vue'
import { useListPage } from '../../composables/useListPage'
import { useCrud } from '../../composables/useCrud'
import {
  getArticleList,
  deleteArticle
} from '../../api/article'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

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
  handleSortChange
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
    apiUrl: '/options?type=status',

    
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
    sortable: false
  },

  {
    field: 'admin.name',
    title: t('article.admin_id'),
    sortable: false
  },

  {
    field: 'created_at',
    title: t('article.created_at'),
    sortable: false
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
    width: 180,
    fixed: 'right',
    slot: 'operation'
  }
])

const handleEdit = (row) => {
  editId.value = row.id
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  await handleDeleteCrud(row, loadData)
}

const handleFormSuccess = () => {
  handleClose()
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.article-list {
  padding: 20px;
}
</style>