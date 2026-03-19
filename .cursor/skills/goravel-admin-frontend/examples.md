## API module templates

### 1) Standard CRUD module (matches existing `user.js` pattern)
This repo often exports named functions like `getXList/getXDetail/createX/updateX/deleteX`.

Create a module like `html/src/api/role.js` style:

```js
import { createCRUDApi } from '../utils/apiFactory'

const roleApi = createCRUDApi('roles')

export const {
  list: getRoleList,
  detail: getRoleDetail,
  create: createRole,
  update: updateRole,
  delete: deleteRole
} = roleApi
```

Use in a page:

```js
const res = await getRoleList({ page: 1, page_size: 10 })
// res is the backend envelope: { code, message, data, trace_id }
// list endpoints typically return: res.data.list / res.data.total / res.data.page / res.data.page_size
```

### 2) Extend CRUD with custom endpoints
```js
import request from '../utils/request'
import { createCRUDApi, extendApi } from '../utils/apiFactory'

const base = createCRUDApi('menus')

export const menusApi = extendApi(base, {
  // Example that matches existing `menu.js` endpoints:
  tree: () =>
    request({
      url: '/menus/tree',
      method: 'get'
    })
})
```

## Page wiring templates

### 3) Use `useApiRequest()` for cancellable calls
```js
import { useApiRequest } from '@/composables/useApiRequest'
import { getUserList } from '@/api/user'

const { request, loading, error } = useApiRequest()

const fetchUsers = () =>
  request(() => getUserList({ page: 1, page_size: 10 }))

try {
  const res = await fetchUsers()
  if (!res) return // cancelled
  tableData.value = res.data.list
} catch (err) {
  // If err.__handled === true, global toast already happened.
  // Only show local UI if needed (e.g. inline error state).
}
```

### 3.1) Preferred list-page pattern: `useListPage` + `Pagination` auto-load
Most list views in this repo (e.g. `views/user/UserList.vue`) use:
- `useListPage({ fetchApi, initialSearchForm, defaultSort, tableRef, transformData })`
- `Pagination v-model="pagination" :auto-load="true" :on-page-change="loadData"`
- `SearchForm @search="handleSearch" @reset="handleReset"`

Minimal skeleton:

```js
import { computed, ref } from 'vue'
import { useListPage } from '@/composables/useListPage'
import { getXList } from '@/api/x'

const tableRef = ref(null)

const initialSearchForm = {
  name: '',
  status: ''
}

const transformRow = (row) => ({
  // normalize backend fields (supports id/ID, created_at/CreatedAt, etc.)
  id: row.id || row.ID,
  name: row.name || row.Name || ''
})

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
  fetchApi: getXList,
  initialSearchForm,
  defaultSort: 'id:desc',
  tableRef: computed(() => tableRef.value?.tableRef),
  transformData: transformRow
})

// In template:
// <SearchForm :model="searchForm" ... @search="handleSearch" @reset="handleReset" />
// <VxeTable ref="tableRef" :data="tableData" :loading="loading" @sort-change="handleSortChange" />
// <Pagination v-model="pagination" :auto-load="true" :on-page-change="loadData" />
```

Notes:
- `useListPage` builds params as: `{ page, page_size, order_by, ...searchForm }` via `buildSearchParams`.
- `useTableData` expects list payload at `res.data.list` (fallback `res.data.data`) and updates `pagination.total` from `res.data.total`.

### 3.2) Standard list-page “extras”: permissions + toolbar + column setting + row actions
Pattern used by many list pages (e.g. `views/user/UserList.vue`):
- Permissions:
  - `const { getButtonState } = usePermission()`
  - Disable buttons: `:disabled="getButtonState('user.store').disabled"`
- Column setting:
  - `const { tableColumns, visibleColumns, allColumns, defaultVisibleColumns, columnOrder, fixedColumns, handleColumnSettingConfirm } = useColumnSetting('<page_key>', allTableColumns)`
- Toolbar:
  - `<TableToolbar :on-refresh="handleRefresh" fullscreen-target=".<page-root>" ... :on-column-setting-confirm="handleColumnSettingConfirm" />`
- Row actions:
  - `<TableActionButtons :row="row" :primary-actions="getPrimaryActions(row)" :more-actions="getMoreActions(row)" :get-button-state="getButtonState" @action="handleAction" />`

Minimal skeleton:

```js
import { computed, ref } from 'vue'
import TableToolbar from '@/components/TableToolbar.vue'
import TableActionButtons from '@/components/TableActionButtons.vue'
import { usePermission } from '@/composables/usePermission'
import { useColumnSetting } from '@/composables/useColumnSetting'

const { getButtonState } = usePermission()

const allTableColumns = computed(() => [
  { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
  { field: 'name', title: t('table.name'), key: 'name' },
  { slot: 'operation', title: t('common.operation'), key: 'operation', width: 180, fixed: 'right' }
])

const {
  tableColumns,               // use this in <VxeTable :columns="tableColumns" />
  visibleColumns,
  allColumns,
  defaultVisibleColumns,
  columnOrder,
  fixedColumns,
  handleColumnSettingConfirm
} = useColumnSetting('x_list', allTableColumns)

const handleRefresh = () => loadData()

const getPrimaryActions = (row) => ([
  { key: 'edit', label: t('common.edit'), type: 'primary', permission: 'x.update', handler: () => handleEdit(row) },
  { key: 'delete', label: t('common.delete'), type: 'danger', permission: 'x.destroy', handler: () => handleDelete(row) }
])

const getMoreActions = (row) => ([
  { key: 'detail', label: t('common.detail'), permission: 'x.show', handler: () => handleDetail(row) }
])
```

Template wiring:

```vue
<TableToolbar
  :on-refresh="handleRefresh"
  fullscreen-target=".x-list"
  :visible-columns="visibleColumns"
  :all-columns="allColumns"
  :default-visible-columns="defaultVisibleColumns"
  :column-order="columnOrder"
  :fixed-columns="fixedColumns"
  :on-column-setting-confirm="handleColumnSettingConfirm"
/>

<VxeTable :columns="tableColumns" ...>
  <template #operation="{ row }">
    <TableActionButtons
      :row="row"
      :primary-actions="getPrimaryActions(row)"
      :more-actions="getMoreActions(row)"
      :get-button-state="getButtonState"
    />
  </template>
</VxeTable>
```

### 4) Handling errors in Login view (auth endpoints)
Auth endpoints are treated specially: global interceptor avoids showing toasts and sets `__handled=false` so the page can render a precise message.

```js
try {
  await login(form) // from `html/src/api/auth.js`
  // navigate
} catch (err) {
  // Here you should show err.message / err.translatedMessage locally.
  // err.errorCode contains backend error_code for branching (e.g. google_code_required).
}
```

