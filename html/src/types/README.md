# TypeScript 类型定义

本目录包含项目的 TypeScript 类型定义文件。

## 文件说明

- `index.d.ts` - 通用类型和业务实体类型定义
- `composables.d.ts` - Composables 函数的类型定义

## 使用方式

### 在 JavaScript 文件中使用（JSDoc）

```javascript
/**
 * @typedef {import('../types').Admin} Admin
 * @typedef {import('../types').Pagination} Pagination
 */

/**
 * @param {Admin} admin
 * @returns {string}
 */
function getAdminDisplayName(admin) {
  return admin.nickname || admin.username
}
```

### 在 TypeScript 文件中使用

```typescript
import type { Admin, Pagination } from '../types'

function getAdminDisplayName(admin: Admin): string {
  return admin.nickname || admin.username
}
```

### 在 Vue 组件中使用

```vue
<script setup lang="ts">
import type { Admin } from '../types'
import { ref } from 'vue'

const admin = ref<Admin | null>(null)
</script>
```

## 类型检查

运行类型检查（不影响构建）：

```bash
npm run type-check
```

## 注意事项

1. **不强制使用 TypeScript**：现有的 `.js` 文件可以继续使用，不需要修改
2. **渐进式采用**：新文件可以选择使用 `.ts` 或 `.js`
3. **类型定义共享**：`.d.ts` 文件中的类型可以在 JS 和 TS 文件中使用
4. **IDE 支持**：即使使用 JS 文件，配合 JSDoc 注释也能获得类型提示

