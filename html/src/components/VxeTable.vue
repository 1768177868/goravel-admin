<!--
  封装的 VXE Table 组件
  简化列表页的表格渲染逻辑
  
  使用示例：
  <VxeTable
    ref="tableRef"
    :data="tableData"
    :loading="loading"
    :columns="tableColumns"
    :height="600"
    @sort-change="handleSortChange"
  >
    <template #status="{ row }">
      <el-tag>...</el-tag>
    </template>
    <template #operation="{ row }">
      <TableActionButtons ... />
    </template>
  </VxeTable>
-->
<template>
  <vxe-table
    ref="tableRef"
    :data="data"
    :loading="loading"
    border
    :column-config="{ resizable: true }"
    :height="height"
    :tree-config="treeConfig"
    :sort-config="{ multiple: false, trigger: 'default' }"
    @sort-change="handleSortChange"
    @checkbox-change="handleCheckboxChange"
    @checkbox-all="handleCheckboxAll"
  >
    <template v-for="(column, index) in columns" :key="column.field || column.slot || index">
      <vxe-column
        :type="column.type"
        :field="column.field"
        :title="column.title"
        :width="column.width"
        :min-width="column.minWidth"
        :sortable="column.sortable"
        :fixed="column.fixed"
        :formatter="column.formatter"
        :tree-node="column.treeNode"
      >
        <template v-if="column.slot" #default="{ row }">
          <slot :name="column.slot" :row="row" />
        </template>
      </vxe-column>
    </template>
  </vxe-table>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  data: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  columns: {
    type: Array,
    required: true
  },
  height: {
    type: [String, Number],
    default: 600
  },
  treeConfig: {
    type: [Object, Boolean],
    default: null
  }
})

const emit = defineEmits(['sort-change', 'checkbox-change', 'checkbox-all'])

const tableRef = ref(null)

const handleSortChange = (params) => {
  emit('sort-change', params)
}

const handleCheckboxChange = (params) => {
  emit('checkbox-change', params)
}

const handleCheckboxAll = (params) => {
  emit('checkbox-all', params)
}

defineExpose({
  get tableRef() {
    return tableRef.value
  }
})
</script>

