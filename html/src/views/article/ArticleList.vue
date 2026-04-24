<template>
  <div class="article-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t("menu.article") }}</span>

          <el-button
            type="primary"
            :disabled="getButtonState('article.store').disabled"
            @click="handleAdd"
          >
            <el-icon><PlusIcon /></el-icon>
            {{ $t("common.add") }}
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
        <template #extra-buttons>
          <el-button
            v-if="enableBatchActions && hasSelection"
            type="danger"
            :disabled="getButtonState('article.destroy').disabled"
            @click="handleBatchDelete"
          >
            {{ `${$t("common.batch_delete")} (${selectedIds.length})` }}
          </el-button>

          <el-button
            v-if="enableBatchActions && hasSelection"
            @click="handleClearSelection"
          >
            {{ $t("common.reset") }}
          </el-button>

          <el-button
            type="success"
            :disabled="getButtonState('article.export').disabled || isExporting"
            :loading="isExporting"
            @click="handleExport"
          >
            {{ $t("common.export") }}
          </el-button>
        </template>
      </SearchForm>

      <VxeTable
        ref="tableRef"
        :data="tableData"
        :loading="loading"
        :columns="tableColumns"
        :height="600"
        @sort-change="handleSortChange"
        @checkbox-change="handleTableCheckboxChange"
        @checkbox-all="handleTableCheckboxAll"
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
import { ref, reactive, onMounted, computed, markRaw } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import SearchForm from "../../components/SearchForm.vue";
import Pagination from "../../components/Pagination.vue";
import VxeTable from "../../components/VxeTable.vue";
import TableActionButtons from "../../components/TableActionButtons.vue";
import ArticleForm from "./ArticleForm.vue";
import { useTable } from "../../composables/useTable";
import { usePermission } from "../../composables/usePermission";
import { useCrud } from "../../composables/useCrud";
import { buildSearchParams } from "../../utils/buildSearchParams";

import { getStatusOptions } from "../../utils/fieldOptions";
import {
  getArticleList,
  deleteArticle,
  updateArticle,
  exportArticle,
} from "../../api/article";
import logger from "../../utils/logger";
import ErrorHandler from "../../utils/errorHandler";

const PlusIcon = markRaw(Plus);

// Permission checks
const { getButtonState } = usePermission();

const { t } = useI18n();
const router = useRouter();
const tableRef = ref(null);
const formRef = ref(null);

const isExporting = ref(false);

const {
  dialogVisible,
  editId,
  handleAdd,
  handleClose,
  handleDelete: handleDeleteCrud,
} = useCrud({
  deleteApi: deleteArticle,
});

const initialSearchForm = {
  admin_id: "",
  title: "",
  content: "",
  status: "",
  created_at: [],
  updated_at: [],
};

const rangeSearchFields = ["created_at", "updated_at"];

const buildListParams = (form, baseParams) => {
  const params = buildSearchParams(form, baseParams);

  rangeSearchFields.forEach((fieldName) => {
    const rangeValue = form[fieldName];
    if (!Array.isArray(rangeValue) || rangeValue.length !== 2) {
      delete params[`${fieldName}_start`];
      delete params[`${fieldName}_end`];
      return;
    }

    const [start, end] = rangeValue;
    if (start) {
      params[`${fieldName}_start`] = start;
    } else {
      delete params[`${fieldName}_start`];
    }
    if (end) {
      params[`${fieldName}_end`] = end;
    } else {
      delete params[`${fieldName}_end`];
    }

    // Avoid submitting raw range arrays as query params directly.
    delete params[fieldName];
  });

  return params;
};

const {
  pagination,
  tableData,
  loading,
  searchForm,
  selectedIds,
  loadData,
  refresh,
  handleSearch,
  handleReset,
  handleSelectionChange,
  clearSelection,
  handleSortChange,
  initDefaultSort,
} = useTable({
  fetchApi: getArticleList,
  initialSearchForm,
  buildParams: buildListParams,
  fieldMapping: {},
  defaultSort: "id:desc",
  tableRef: computed(() => tableRef.value?.tableRef),
});

const enableBatchActions = true;
const hasSelection = computed(() => selectedIds.value.length > 0);

const searchFields = computed(() => [
  {
    prop: "admin_id",
    label: t("admin_id"),
    type: "input",
    clearable: true,
    width: "200px",
    advanced: false,
  },
  {
    prop: "title",
    label: t("title"),
    type: "input",
    clearable: true,
    width: "200px",
    advanced: false,
  },
  {
    prop: "content",
    label: t("content"),
    type: "input",
    clearable: true,
    width: "200px",
    advanced: false,
  },
  {
    prop: "status",
    label: t("status"),
    type: "input",
    clearable: true,
    width: "200px",
    advanced: false,
  },
  {
    prop: "created_at",
    label: t("common.created_at"),
    type: "datetimerange",
    clearable: true,
    props: {
      startPlaceholder: t("common.start_time"),
      endPlaceholder: t("common.end_time"),
      rangeSeparator: t("common.range_separator"),
    },
    width: "360px",
    advanced: false,
  },
  {
    prop: "updated_at",
    label: t("common.updated_at"),
    type: "datetimerange",
    clearable: true,
    props: {
      startPlaceholder: t("common.start_time"),
      endPlaceholder: t("common.end_time"),
      rangeSeparator: t("common.range_separator"),
    },
    width: "360px",
    advanced: false,
  },
]);

const tableColumns = computed(() => {
  const baseColumns = [
    {
      field: "id",
      title: t("table.id"),
      width: 80,
      sortable: true,
    },
    {
      field: "admin_id",
      title: t("admin_id"),
      slot: "admin_id",
      sortable: false,
    },
    {
      field: "title",
      title: t("title"),
      sortable: false,
    },
    {
      field: "content",
      title: t("content"),
      sortable: false,
    },
    {
      field: "status",
      title: t("status"),
      sortable: false,
    },
    {
      field: "updated_at",
      title: t("table.updated_at"),
      sortable: false,
    },
    {
      field: "created_at",
      title: t("table.created_at"),
      width: 180,
      sortable: true,
    },
    {
      field: "operation",
      title: t("table.operation"),
      width: 220,
      fixed: "right",
      slot: "operation",
    },
  ];

  if (!enableBatchActions) {
    return baseColumns;
  }

  return [
    {
      type: "checkbox",
      width: 52,
      fixed: "left",
    },
    ...baseColumns,
  ];
});

const getadminDisplayName = (admin_id) => {
  if (!admin_id) return "-";
  return admin_id.name || admin_id.admin || "-";
};

const handleEdit = (row) => {
  editId.value = row.id;
  dialogVisible.value = true;
};

const handleDelete = (row) => handleDeleteCrud(row, loadData);

const handleFormSuccess = () => {
  handleClose();
  clearSelection();
  refresh();
};

const handleTableCheckboxChange = ({ records }) => {
  handleSelectionChange(records);
};

const handleTableCheckboxAll = ({ records }) => {
  handleSelectionChange(records);
};

const handleClearSelection = () => {
  clearSelection();
  tableRef.value?.tableRef?.clearCheckboxRow?.();
};

const handleBatchDelete = async () => {
  if (!selectedIds.value.length) return;

  try {
    await ElMessageBox.confirm(
      t("common.batch_delete_confirm", { count: selectedIds.value.length }),
      t("common.warning"),
      {
        type: "warning",
        confirmButtonText: t("common.confirm"),
        cancelButtonText: t("common.cancel"),
      },
    );

    await Promise.all(selectedIds.value.map((id) => deleteArticle(id)));
    ElMessage.success(t("common.operation_success"));
    handleClearSelection();
    await refresh();
  } catch (error) {
    if (error === "cancel" || error === "close") {
      return;
    }
    logger.error("Batch delete error:", error);
    if (!error.__handled) {
      ErrorHandler.handle(error, { silent: true });
    }
  }
};

// Get primary action button config.
const getPrimaryActions = (row) => {
  return [
    {
      key: "edit",
      label: t("common.edit"),
      type: "primary",
      permission: "article.update",
      handler: handleEdit,
    },
    {
      key: "delete",
      label: t("common.delete"),
      type: "danger",
      permission: "article.destroy",
      handler: handleDelete,
    },
  ];
};

// Get secondary action button config (extend when needed).
const getMoreActions = (row) => {
  return [];
};

// Handle action events.
const handleAction = (command, row) => {
  switch (command) {
    case "edit":
      handleEdit(row);
      break;
    case "delete":
      handleDelete(row);
      break;
  }
};

const handleExport = async () => {
  if (isExporting.value) {
    return;
  }

  isExporting.value = true;

  try {
    const response = await exportArticle(searchForm);
    const exportId =
      response?.data?.export_id ||
      response?.data?.data?.export_id ||
      response?.export_id ||
      response?.data?.id;

    if (!exportId) {
      ElMessage.warning(t("common.operation_success"));
    } else {
      ElMessage.success(t("export.task_submitted"));
      router.push("/exports");
    }
  } catch (error) {
    logger.error("Export error:", error);
    if (error.response?.status === 429) {
      ElMessage.warning(t("common.already_queued"));
    } else if (!error.__handled) {
      ErrorHandler.handle(error, { silent: true });
    }
  } finally {
    isExporting.value = false;
  }
};

onMounted(async () => {
  try {
    initDefaultSort();
    await loadData();
  } catch (error) {
    logger.error("ListPage onMounted error:", error);
    ErrorHandler.handle(error);
  }
});
</script>

<style scoped>
.article-list {
}
</style>
