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
    </el-card>

    ArticleForm
      ref="formRef"
      v-model="dialogVisible"
      :edit-id="editId"
      @success="handleFormSuccess"
    />
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

  name: '',
  status: '',
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
    prop: 'name',
    label: t('article.name'),
    type: 'input',


    width: '200px',
    advanced: false
  },
  {
    prop: 'status',
    label: t('article.status'),
    type: 'select',
    apiUrl: '/options?type=status',

    
    apiUrl: '/options?type=dictionary&dictionary_type=status',
    

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
    field: 'name',

    title: t('article.name'),

    sortable: true
  },

  {
    field: 'status',

    title: t('article.status'),

    sortable: true
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
  try {
    await ElMessageBox.confirm(
      t('common.delete_confirm'),
      t('common.confirm'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    await handleDeleteCrud(row, loadData)
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ErrorHandler.handle(error)
    }
  }
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