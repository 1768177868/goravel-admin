<!--
  通用表格列组件
  用于简化 vxe-table 列的定义和渲染
  
  使用示例：
  <vxe-table>
    <template v-for="(column, index) in tableColumns" :key="column.field || column.slot || index">
      <TableColumn :column="column" :index="index">
        <template v-if="column.slot === 'status'" #status="{ row }">
          <el-tag>...</el-tag>
        </template>
        <template v-if="column.slot === 'operation'" #operation="{ row }">
          <TableActionButtons ... />
        </template>
      </TableColumn>
    </template>
  </vxe-table>
-->
<template>
  <vxe-column
    :type="column.type"
    :field="column.field"
    :title="column.title"
    :width="column.width"
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

<script setup>
defineProps({
  column: {
    type: Object,
    required: true
  },
  index: {
    type: Number,
    default: 0
  }
})
</script>

