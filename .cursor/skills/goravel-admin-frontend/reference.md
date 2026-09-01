## Critical frontend files (do not bypass)
- `html/src/utils/request.js`: axios instance + interceptors + token/i18n/timezone headers + global error handling
- `html/src/utils/apiFactory.js`: CRUD API factory (`createCRUDApi`, `extendApi`)
- `html/src/composables/useStandardListPage.js`: standard paginated CRUD list pages

## Base URL / routing conventions
- API base URL is composed by:
  - `VITE_API_BASE_URL` + `VITE_API_PREFIX` (default prefix: `/api/admin`)
  - if no base URL, uses prefix directly (relative)
- WebSocket path `/ws` is proxied in `vite.config.js` to `VITE_API_BASE_URL` (default `http://127.0.0.1:3000`)

## Where APIs live (observed)
- API modules: `html/src/api/*.js`
  - Example: `html/src/api/auth.js` uses `/login`, `/info`, `/logout` etc.
  - Example: `html/src/api/menu.js` uses `/menus`, `/menus/tree`
  - Example: `html/src/api/user.js` uses `createCRUDApi('users')` + `extendApi(...)`

## Error contract mapping (how to reason about it)
1) HTTP status error (axios error.response.status)
- 401: global logout+redirect, except auth endpoints handled by page
- 403: debounced global message, except auth endpoints handled by page
- 429: non-export endpoints show toast; export endpoints delegated to business code

2) HTTP 200 but business failure (res.code !== 200)
- Treated as error; creates a rejected Error with:
  - `err.code` = res.code (often 400/401/403/404/500)
  - `err.errorCode` = res.error_code
  - `err.message` = extracted/translated message
  - `err.__handled` = true (except auth endpoints)

## i18n translation strategy for backend error codes
`extractErrorInfo()` attempts translation when message looks like a code:
- Try `common.<error_code>` first
- Then `messages.<error_code>`

Implication:
- For new backend error codes, ensure frontend has at least one matching i18n entry
  (prefer `common` for generic keys, `messages` for domain keys), or ensure backend already returns localized `message`.

## Checklist: adding a new backend `error_code`
When backend introduces a new `error_code` (BusinessError.Code), follow this checklist:

1) Decide who owns the user-facing text:
- Preferred: backend returns localized `message` (based on `Accept-Language`), frontend only displays it.
- Acceptable: frontend provides i18n mapping for the `error_code` (when backend message may be code-like).

2) If frontend should translate it, add keys in BOTH locales:
- `html/src/i18n/locales/zh-CN.json`
- `html/src/i18n/locales/en-US.json`

3) Put the key in the right namespace:
- Generic/system keys → `common.<error_code>` (e.g. `operation_failed`, `query_failed`, `validation_failed`)
- Domain/business keys → `messages.<error_code>` (if you want to group them)
- Do NOT put backend `error_code` under `error.*` unless it’s truly a frontend-only network/UI error.

4) If the backend uses placeholders (e.g. `{balance}`), prefer backend-side formatting.
Frontend-side translation does not automatically substitute backend placeholders.

## Token behavior (important)
- Request sends `Authorization: Bearer <token>`
- Response updates token from:
  - response header `authorization` (case-insensitive) OR
  - payload `res.data.token`

Do not manually set Authorization headers in individual API modules unless required for special cases.

## List page data expectations (observed)
The shared list-page composables assume backend list responses look like:
`res.data.list` (default) or `res.data.data` (payment, payment_method, order) plus `res.data.total`.

Both Vue composables and React `normalizeListResponse` handle either rows key.
New backend modules should prefer `list`; match `data` only when editing payment/order modules.

Pagination component expects a model like:
`{ page, pageSize, total }` (note: `pageSize` in state maps to request param `page_size`).

## Permission + UI control conventions (observed)
- Permission checks are done via `usePermission()`:
  - `getButtonState('<resource>.<action>')` returns `{ show, disabled }`
  - Many pages always render buttons but use `:disabled="getButtonState(...).disabled"`
- Common permission strings follow backend route action names, e.g.:
  - `user.store`, `user.update`, `user.destroy`, `user.export`

## Column setting conventions (observed)
- List pages often use `useColumnSetting('<page_key>', allTableColumns)` and pass the returned state into `TableToolbar` → `ColumnSettingDialog`.

