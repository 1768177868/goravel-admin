# useListPage Composable 使用说明

`useListPage` 是一个用于封装列表页面通用逻辑的 composable，可以大大减少重复代码。

## 基本用法

```javascript
import { useListPage } from '@/composables/useListPage'
import { getAdminList } from '@/api/admin'

// 使用 composable
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getAdminList,
  initialSearchForm: {
    username: '',
    status: '',
    role_id: ''
  }
})

// 在 onMounted 中初始化
onMounted(() => {
  initDefaultSort?.()
  loadData()
})
```

## 带排序的用法

```javascript
import { ref } from 'vue'
import { useListPage } from '@/composables/useListPage'
import { getAdminList } from '@/api/admin'

const tableRef = ref(null)

const fieldMapping = {
  'id': 'id',
  'username': 'username',
  'created_at': 'created_at'
}

const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSortChange,
  initDefaultSort
} = useListPage({
  fetchApi: getAdminList,
  initialSearchForm: {
    username: '',
    status: ''
  },
  sortOptions: {
    tableRef,
    fieldMapping,
    defaultSort: 'id:desc'
  }
})
```

## 自定义参数构建

```javascript
const {
  pagination,
  tableData,
  loading,
  searchForm,
  loadData
} = useListPage({
  fetchApi: getAdminList,
  initialSearchForm: {
    username: '',
    status: ''
  },
  buildParams: (searchForm, pagination, orderBy) => {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    
    if (orderBy) {
      params.order_by = orderBy
    }
    
    // 自定义参数构建逻辑
    if (searchForm.username) {
      params.username = searchForm.username.trim()
    }
    
    return params
  }
})
```

## 数据转换

```javascript
const {
  tableData,
  loadData
} = useListPage({
  fetchApi: getAdminList,
  initialSearchForm: {},
  transformData: (item) => {
    // 转换数据格式
    return {
      id: item.id || item.ID,
      username: item.username || item.Username,
      // ... 其他字段
    }
  }
})
```

## 回调函数

```javascript
const {
  loadData
} = useListPage({
  fetchApi: getAdminList,
  initialSearchForm: {},
  onLoadSuccess: (res, list) => {
    // 加载成功后的处理
    console.log('加载成功', list)
  },
  onLoadError: (error) => {
    // 加载失败后的处理
    console.error('加载失败', error)
  }
})
```

## 返回值说明

- `pagination`: 分页状态对象（reactive）
  - `page`: 当前页码
  - `pageSize`: 每页数量
  - `total`: 总记录数

- `tableData`: 表格数据（ref）

- `loading`: 加载状态（ref）

- `searchForm`: 搜索表单数据（reactive）

- `loadData()`: 加载数据方法

- `handleSearch()`: 搜索处理方法（重置页码并加载）

- `handleReset()`: 重置搜索条件方法（重置表单、排序、页码并加载）

- `handlePageChange({ currentPage, pageSize })`: 分页变化处理方法

- `handleSortChange()`: 排序变化处理方法（如果提供了 sortOptions）

- `initDefaultSort()`: 初始化默认排序方法（如果提供了 sortOptions）

## 注意事项

1. `fetchApi` 是必需的参数
2. `initialSearchForm` 定义了搜索表单的初始值
3. 如果提供了 `sortOptions`，会自动集成排序功能
4. `buildParams`、`transformData`、`onLoadSuccess`、`onLoadError` 都是可选参数

