<template>
  <ListPage
    ref="listPageRef"
    page-class="article"
    :title="$t('menu.article')"
    :add-button-text="$t('common.add')"
    :add-button-disabled="getButtonState('article.store').disabled"
    :search-form="searchForm"
    :search-fields="searchFields"
    :initial-search-values="articleInitialSearchForm"
    i18n-prefix="article"
    :table-data="tableData"
    :loading="loading"
    :table-columns="tableColumns"
    :pagination="pagination"
    :show-toolbar="true"
    @add="handleAdd"
    @search="handleSearch"
    @reset="handleReset"
    @refresh="loadData"
    @page-change="loadData"
    @sort-change="handleSortChange"
  >
    <template #admin_id="{ row }">
      {{ getadminDisplayName(row.admin || row.admin_id) }}
    </template>

    <template #operation="{ row }">
      <TableActionButtons
        :row="row"
        :primary-actions="operationActions"
        :get-button-state="getButtonState"
      />
    </template>

    <template #form>
      <ArticleForm
        v-model="dialogVisible"
        :edit-id="editId"
        @success="handleFormSuccess"
      />
    </template>
  </ListPage>
</template>

<script setup>
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import ListPage from "@/components/ListPage.vue";
import TableActionButtons from "@/components/TableActionButtons.vue";
import ArticleForm from "./ArticleForm.vue";
import { useStandardListPage } from "@/composables/useStandardListPage";
import { createCrudActions } from "@/utils/listPageHelpers";
import { getArticleList, deleteArticle } from "@/api/article";
import {
  articleInitialSearchForm,
  buildArticleListParams,
  createArticleSearchFields,
  createArticleTableColumns,
  getadminDisplayName,
} from "./article.config";

const { t } = useI18n();
const listPageRef = ref(null);

const {
  pagination,
  tableData,
  loading,
  searchForm,
  dialogVisible,
  editId,
  loadData,
  handleSearch,
  handleReset,
  handleSortChange,
  handleAdd,
  handleEdit,
  handleFormSuccess,
  handleDelete,
  getButtonState,
} = useStandardListPage({
  fetchApi: getArticleList,
  initialSearchForm: articleInitialSearchForm,
  buildParams: buildArticleListParams,
  defaultSort: "id:desc",
  deleteApi: deleteArticle,
  tableRef: computed(() => listPageRef.value?.tableRef?.tableRef),
  normalizeRows: false,
});

const searchFields = computed(() => createArticleSearchFields(t));
const tableColumns = computed(() =>
  createArticleTableColumns(t, { enableBatchActions: false }),
);

const operationActions = computed(() =>
  createCrudActions(t, "article", {
    onEdit: handleEdit,
    onDelete: handleDelete,
  }),
);
</script>
