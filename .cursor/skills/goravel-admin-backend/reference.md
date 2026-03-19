## Module names for `ErrorWithLog` (reuse these)
Prefer reusing existing module strings so logs are grouped consistently.

Observed in controllers:
- `auth`
- `captcha`
- `admin`
- `online_admin`
- `user`
- `password`
- `menu`
- `role`
- `permission`
- `department`
- `dictionary`
- `blacklist`
- `order`
- `payment`
- `payment_method`
- `attachment`
- `export`
- `import`
- `config`
- `login-log`
- `operation-log`
- `system-log`
- `menu-test` (debug/test)

Guideline:
- If a module already exists for the domain, **do not invent a new one**.
- If adding a new domain, prefer **kebab-case** for new module names and keep it stable:
  - ✅ kebab-case: `payment-method`, `online-admin`, `login-log`
  - ⚠️ existing snake_case modules exist (`payment_method`, `online_admin`) — keep them as-is, but don’t introduce more snake_case.
  - Rule of thumb: use the same word separator style as the closest existing module; otherwise default to kebab-case.

## What to put in `attrs` (safe, useful, consistent)
`ErrorWithLog` will auto-add `error` if you don’t set it. Provide only small, non-sensitive context.

Good attrs:
- identifiers: `id`, `<resource>_id`, `user_id`, `admin_id`, `role_id`, `menu_id`, `order_id`
- lookup keys: `username`, `slug`, `email`, `phone`
- action context: `action`, `status`, `type`
- pagination: `page`, `page_size`

Avoid attrs:
- passwords, tokens, Authorization headers
- full request bodies (too big + may contain secrets)
- personal data beyond what’s required to debug (minimize)

## Decision table (which response helper to use)
- Request validation failed (`errors != nil`): `response.ValidationError(ctx, 400, "validation_failed", errors.All())`
- Business rule failure (expected): `response.Error(ctx, 400/401/403/404, businessErrOrCode)`
  - Prefer passing the **error** when using `WithParams` so formatting is preserved:
    - ✅ `response.Error(ctx, http.StatusBadRequest, err)`
    - ⚠️ `response.Error(ctx, http.StatusBadRequest, businessErr.Code)` (loses params formatting)
- Unexpected infra failure (DB/IO/remote): `response.ErrorWithLog(ctx, module, err, attrs)` (usually 500)

## Response payload key conventions (observed)
- Paginated lists (most controllers): `list`, `total`, `page`, `page_size`
- Detail endpoints: singular key, e.g. `user`, `role`, `permission`, `dictionary`, `department`, `menu`
- Tree endpoints:
  - Menu: returns `menus` and `list` with the same value (compat)
  - Department: returns `list` only (tree or flat search results)
- “lookup” endpoints may use plural nouns, e.g. `dictionaries`, `types`

## Copy/paste examples

### Example: standard request validation
```go
var req adminrequests.MenuCreate
errors, err := ctx.Request().ValidateRequest(&req)
if err != nil {
    return response.Error(ctx, http.StatusBadRequest, err.Error())
}
if errors != nil {
    return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
}
```

### Example: logging unexpected errors safely
```go
if err := facades.Orm().Query().Save(&model); err != nil {
    return response.ErrorWithLog(ctx, "menu", err, map[string]any{
        "menu_id": model.ID,
        "slug":    model.Slug,
    })
}
```

### Example: BusinessError with placeholders
```go
// service layer:
return apperrors.ErrInsufficientBalance.WithParams(map[string]any{"balance": currentBalance})

// controller boundary (preserves WithParams formatting):
if err != nil {
    return response.Error(ctx, http.StatusBadRequest, err)
}
```

