---
name: goravel-admin-backend
description: Implements backend features in this Goravel admin project using the existing HTTP response helpers (response.Success/Error/ValidationError/ErrorWithLog/Paginate/FindByID/PaginateQuery), BusinessError codes with i18n (trans.Get), and consistent HTTP status mapping. Use when adding or modifying controllers, requests validation, services, repositories/ORM queries, auth/permission middleware, exports/imports, uploads, or when fixing bugs around API responses and error codes.
---

# Goravel Admin Backend

## Canonical response format (must match)
All API responses are JSON and use the helpers in `app/http/response/response.go`.

## Additional resources
- For module names, logging attrs conventions, and copy-paste examples, see [reference.md](reference.md).
- For controller/service skeletons (CRUD, pagination, error handling), see [examples.md](examples.md).

### Success JSON
- `code`: `200`
- `message`: translated via `trans.Get(ctx, messageKey)` (default `"success"`)
- `data`: optional
- `trace_id`: auto-included when present in context

Use:
- `response.Success(ctx)` / `response.Success(ctx, data)` / `response.Success(ctx, "message_key", data)`
- `response.SuccessWithHeader(ctx, "message_key", headerKey, headerValue, data)`
- `response.Paginate(ctx, ...)` or `response.PaginateQuery(ctx, query, &list, options)`

### Error JSON
- `code`: HTTP status code (and also the JSON `code` field)
- `message`: user-facing message (translated)
- `error_code`: stable string key for frontend branching
- `trace_id`: auto-included when present in context

Use:
- `response.Error(ctx, httpStatus, messageKeyOrErr)`
  - if `messageKeyOrErr` is `error` and it’s a `*errors.BusinessError`, it will use `GetFormattedMessage(ctx)` and set `error_code = BusinessError.Code`.
  - if `messageKeyOrErr` is a string, it is treated as a **translation key** (aka error_code).
- `response.ErrorWithLog(ctx, ...)` / `response.ErrorWithLogAuto(ctx, module, ...)` for 500+ errors with automatic error logging.

### Validation error JSON
- same as Error JSON plus:
  - `errors`: field-level map (from `errors.All()`)
  - `message`: first field error message if available (otherwise translated base message)

Use:
- `response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())`

## Error model (BusinessError + i18n)
The project’s business error type is `*errors.BusinessError` in `app/errors/errors.go`.

### Rules
- Prefer returning `*errors.BusinessError` for expected, user-facing failures (validation-like, not-found, conflicts, forbidden, business rules).
- Error “code” is the contract: use stable snake_case keys.
- User-facing message must be i18n-friendly:
  - prefer `trans.Get(ctx, code)` resolution via `BusinessError.GetFormattedMessage(ctx)`
  - if translation missing, it must fall back to the default `BusinessError.Message`
- For dynamic content in messages, use `WithParams(...)` so placeholders are replaced.
  - Supported placeholders: `{key}` and `${key}`
  - Numeric formatting: float32/float64 -> 2 decimals; ints -> integer string

### Don’t duplicate codes
Before adding a new business error, search existing codes in `app/errors/errors.go` and reuse when semantics match.

## Controller workflow (match existing style)

### Request validation pattern (standard)
For create/update endpoints, follow this pattern:
1. `errors, err := ctx.Request().ValidateRequest(&req)`
2. If `err != nil`: return `response.Error(ctx, http.StatusBadRequest, err.Error())`
3. If `errors != nil`: return `response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())`

This is used in controllers like:
- `app/http/controllers/api/auth_controller.go`
- `app/http/controllers/admin/menu_controller.go`

### HTTP status mapping (default choices)
Use these defaults unless an existing controller in the same module does otherwise:
- **400 Bad Request**
  - request validation failures (`response.ValidationError`)
  - invalid/required params (e.g., `id_required`, `params_error`)
  - “already exists” conflicts handled as business rule (e.g., `*_exists`, `menu_slug_exists`)
- **401 Unauthorized**
  - not logged in / invalid credentials (e.g., `not_logged_in`, `username_or_password_error`)
- **403 Forbidden**
  - disabled account / permission forbidden / protected actions (e.g., `account_disabled`, `protected_*`)
- **404 Not Found**
  - resource not found (use `response.FindByID` or `response.Error(ctx, http.StatusNotFound, "<resource>_not_found")`)
- **500 Internal Server Error**
  - unexpected DB/IO/infra failures; prefer `response.ErrorWithLog(ctx, module, err, attrs)` so it’s logged

### Use response helpers instead of hand-rolled JSON
Do not craft response JSON manually in controllers. Use:
- `response.Success`, `response.SuccessWithHeader`
- `response.Error`, `response.ErrorWithLog`, `response.ErrorWithLogAuto`
- `response.ValidationError`
- `response.FindByID` (preferred for show/update/destroy)
- `response.PaginateQuery` (preferred for index lists)

## Common patterns to follow

### Find by ID (preferred)
For show/update/destroy, prefer:
- `model, resp := response.FindByID[models.X](ctx, id, options)`
- If `resp != nil` return it

Notes:
- id==0 -> `400` with `id_required`
- not found -> `404` with `record_not_found` or `options.NotFoundMessageKey`

### System errors should be logged (500+)
If an operation can fail due to DB/IO/remote calls, prefer:
- `response.ErrorWithLog(ctx, "<module>", err, map[string]any{...context...})`
so `response.Error` will include a user-safe message and log details server-side.

Keep attrs small and safe (no secrets, passwords, tokens).

### Business errors should be returned as codes
If a service returns a `*errors.BusinessError`, bubble it up and respond using its code:
- Recommended: `return response.Error(ctx, http.StatusBadRequest, err)` (lets `Error` detect BusinessError and format message/params)
- Acceptable (existing style in some controllers): `return response.Error(ctx, http.StatusBadRequest, businessErr.Code)`

Prefer passing the `error` (not only the code) when you rely on `WithParams` formatting.

## Deliverable expectations (how the agent should report work)
When implementing changes, always include:
- **Changed files**: list + 1-line purpose each
- **API behavior**: success payload keys + relevant error cases (`error_code`, status codes)
- **New/updated error codes**: only if added; explain why not reusable
- **Test plan**: commands + at least 2-3 manual request examples (happy path + failure path)

## References
- `app/http/response/response.go`: canonical response helpers, logging behavior, validation error payload
- `app/errors/errors.go`: `BusinessError`, `GetFormattedMessage(ctx)`, placeholder replacement via `WithParams`
