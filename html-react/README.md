# Goravel Admin (React)

Vue 版后台（`html/`）的 React 对照实现，对接同一套 Goravel Admin API。

## 技术栈

- React 19 + TypeScript + Vite
- Ant Design 6
- Zustand / React Router 7 / axios / react-i18next

## 快速开始

```bash
cd html-react
cp .env.example .env
npm install
npm run dev
```

开发地址默认：`http://localhost:3008`  
后端默认：`http://localhost:3000`（与 `.env` 中 `VITE_API_BASE_URL` 一致）

> 若登录出现「网络连接失败」，多半是后端 CORS 未放行 React 端口。确认根目录 `.env` 含：
> `http://localhost:3008`，以及请求头 `Accept-Language`，然后**重启 Go 后端**。

> 生产环境托管 `dist` 时需做 SPA fallback（否则刷新子路由 404），Nginx 示例：
> `try_files $uri $uri/ /index.html;`

默认账号与 Vue 版相同：`admin` / `admin123`

## 目录结构

```
src/
  api/          # 接口模块（createCRUDApi / extendApi）
  components/   # 通用 UI
  hooks/        # useListPage / usePermission 等
  i18n/         # 多语言
  layouts/      # 主布局、侧栏、多标签
  pages/        # 页面（菜单 component 字段映射到此）
  router/       # 静态路由 + 动态菜单路由
  stores/       # Zustand：user / app / tabs
  types/        # 类型与 API 契约
  utils/        # request / storage / normalize / tree ...
```

## 与 Vue 版对照

| Vue (`html/`) | React (`html-react/`) |
|---|---|
| Pinia | Zustand |
| Element Plus | Ant Design 6 |
| `useListPage` | `hooks/useListPage` |
| `utils/request.js` | `utils/request.ts` |
| `utils/apiFactory.js` | `utils/apiFactory.ts` |
| `views/**` | `pages/**` |

API 响应约定不变：`{ code, message, data, error_code, trace_id }`。

## 当前已实现模块

- 登录 / 布局 / 动态菜单 / 多标签 / 权限按钮 / 通知铃铛（WS）
- 布局壳层：锁屏、时区切换、全屏、字号、顶栏/侧栏菜单、水印、主题色、列设置
- Dashboard（KPI + ECharts）、个人中心（头像 / 谷歌 2FA）
- 系统：管理员、角色（权限树）、权限、菜单、部门、岗位、在线管理员、字典、配置、导出、附件、黑名单
- 业务：用户（余额/重置密码/余额日志）、订单、支付方式、支付记录、文章（WangEditor）
- 通知：创建（Markdown / 富文本）、列表、详情
- 日志：操作 / 登录 / 系统（详情、批量删除、清理）+ 观测中心
- 监控：服务监控（SSE + ECharts）
- 开发：表单演示（含 WangEditor / Markdown）

**未实现（按约定暂缓）：** 代码生成器（`dev/CodeGenerator`）仍走占位页。
## 脚本

- `npm run dev` — 开发
- `npm run build` — 类型检查 + 生产构建
- `npm run type-check` — 仅类型检查
- `npm run preview` — 预览构建产物

更多说明见 [DEVELOPMENT.md](./DEVELOPMENT.md)。
