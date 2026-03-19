## Controller skeletons (copy/paste)
These examples match the patterns in:
- `app/http/controllers/admin/menu_controller.go`
- `app/http/controllers/api/auth_controller.go`
- `app/http/response/response.go`

### 1) Index: standard paginated list (most controllers)
Most list endpoints in this repo return `data` with these keys:
`list`, `total`, `page`, `page_size`.

```go
func (r *XController) Index(ctx http.Context) http.Response {
    page := helpers.GetIntQuery(ctx, "page", 1)
    pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

    // Pattern used widely: controller has buildFilters(ctx) -> services.<Domain>Filters
    // filters := r.buildFilters(ctx)
    // OR build inline if the module is small.
    filters := services.XFilters{
        Name:      ctx.Request().Query("name", ""),
        Status:    ctx.Request().Query("status", ""),
        StartTime: helpers.GetTimeQueryParam(ctx, "start_time"),
        EndTime:   helpers.GetTimeQueryParam(ctx, "end_time"),
        OrderBy:   ctx.Request().Query("order_by", ""),
    }

    list, total, err := r.xService.GetList(filters, page, pageSize)
    if err != nil {
        // If the service returns a BusinessError, prefer returning its Code for frontend branching.
        // For system errors, use ErrorWithLog with a stable module.
        if businessErr, ok := apperrors.GetBusinessError(err); ok {
            return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
        }
        return response.ErrorWithLog(ctx, "x", err, map[string]any{
            "page":      page,
            "page_size": pageSize,
        })
    }

    return response.Success(ctx, http.Json{
        "list":      list,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}
```

### 1.1) Index: optional shortcut via `response.Paginate(...)`
Some controllers use `response.Paginate(ctx, list, total, page, pageSize)` (e.g. online admins, export lists).

```go
return response.Paginate(ctx, list, total, page, pageSize)
```

### 2) Show: `FindByID` + success payload
```go
func (r *XController) Show(ctx http.Context) http.Response {
    id := helpers.GetUintRoute(ctx, "id")
    x, resp := response.FindByID[models.X](ctx, id, &response.FindByIDOptions{
        NotFoundMessageKey: apperrors.ErrRecordNotFound.Code, // or "x_not_found"
        WithRelations:      []string{"RelA"},                 // optional
    })
    if resp != nil {
        return resp
    }

    return response.Success(ctx, http.Json{
        "x": *x,
    })
}
```

### 2.1) Tree endpoints: return both `<plural>` and `list` for compatibility
Tree endpoints are not fully consistent across modules:
- Menu returns both `menus` and `list` (compat).
- Department returns only `list`.

Pick the module’s existing convention. If unsure, prefer returning `list`, and optionally add a plural alias for compatibility when the frontend expects it.

```go
func (r *XController) Tree(ctx http.Context) http.Response {
    treeData, err := r.xService.BuildTree(0)
    if err != nil {
        return response.ErrorWithLog(ctx, "x", err)
    }
    // Optionally apply filters (dev/monitor hidden, etc.)
    // treeData = r.applyTreeFilters(ctx, treeData)

    return response.Success(ctx, http.Json{
        "list": treeData,
        // Optional plural alias (menu-style compatibility):
        // "xs": treeData,
    })
}
```

### 3) Store: ValidateRequest + uniqueness check + service + ErrorWithLog
```go
func (r *XController) Store(ctx http.Context) http.Response {
    var req adminrequests.XCreate
    errors, err := ctx.Request().ValidateRequest(&req)
    if err != nil {
        return response.Error(ctx, http.StatusBadRequest, err.Error())
    }
    if errors != nil {
        return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
    }

    // Example: uniqueness check
    exists, err := facades.Orm().Query().Model(&models.X{}).Where("slug", req.Slug).Exists()
    if err != nil {
        return response.ErrorWithLog(ctx, "x", err, map[string]any{"slug": req.Slug})
    }
    if exists {
        return response.Error(ctx, http.StatusBadRequest, "x_slug_exists") // prefer reusing apperrors.* if available
    }

    x, err := r.xService.Create(req /* ... */)
    if err != nil {
        if businessErr, ok := apperrors.GetBusinessError(err); ok {
            // NOTE: many existing controllers pass `.Code` here.
            // If you rely on `WithParams(...)`, prefer `response.Error(ctx, 400, err)` to preserve placeholder formatting.
            return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
        }
        return response.ErrorWithLog(ctx, "x", err, map[string]any{"slug": req.Slug})
    }

    return response.Success(ctx, http.Json{"x": *x})
}
```

### 4) Update: partial update using `Request().All()` (pattern from Menu)
Use this when you support PATCH-like updates and must only modify provided fields.

```go
func (r *XController) Update(ctx http.Context) http.Response {
    id := helpers.GetUintRoute(ctx, "id")
    x, resp := response.FindByID[models.X](ctx, id, &response.FindByIDOptions{
        NotFoundMessageKey: "x_not_found",
    })
    if resp != nil {
        return resp
    }

    var req adminrequests.XUpdate
    errors, err := ctx.Request().ValidateRequest(&req)
    if err != nil {
        return response.Error(ctx, http.StatusBadRequest, err.Error())
    }
    if errors != nil {
        return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
    }

    allInputs := ctx.Request().All()

    if _, ok := allInputs["name"]; ok {
        (*x).Name = req.Name
    }
    if _, ok := allInputs["slug"]; ok {
        // optional uniqueness check:
        exists, err := facades.Orm().Query().Model(&models.X{}).
            Where("slug", req.Slug).Where("id != ?", id).Exists()
        if err != nil {
            return response.ErrorWithLog(ctx, "x", err, map[string]any{"x_id": id, "slug": req.Slug})
        }
        if exists {
            return response.Error(ctx, http.StatusBadRequest, "x_slug_exists")
        }
        (*x).Slug = req.Slug
    }

    if err := r.xService.Update(x); err != nil {
        return response.ErrorWithLog(ctx, "x", err, map[string]any{"x_id": (*x).ID})
    }

    return response.Success(ctx, http.Json{"x": *x})
}
```

### 5) Destroy: children/constraint checks + service delete
```go
func (r *XController) Destroy(ctx http.Context) http.Response {
    id := helpers.GetUintRoute(ctx, "id")
    x, resp := response.FindByID[models.X](ctx, id, &response.FindByIDOptions{
        NotFoundMessageKey: "x_not_found",
    })
    if resp != nil {
        return resp
    }

    // Optional constraint check (example)
    if hasChildren, err := r.xService.HasChildren(id); err != nil {
        return response.ErrorWithLog(ctx, "x", err, map[string]any{"x_id": id})
    } else if hasChildren {
        return response.Error(ctx, http.StatusBadRequest, "x_has_children")
    }

    if err := r.xService.Delete(x); err != nil {
        return response.ErrorWithLog(ctx, "x", err, map[string]any{"x_id": (*x).ID})
    }

    return response.Success(ctx)
}
```

## Auth-style patterns (header + 401/403)

### Login/register returning token in header
Matches `response.SuccessWithHeader` usage from `api/auth_controller.go`.

```go
return response.SuccessWithHeader(ctx, "login_success", "Authorization", "Bearer "+token, http.Json{
    "token": token,
    "user": http.Json{
        "id": user.ID,
        // ...
    },
})
```

### Common status choices in auth flows
- Invalid credentials: `response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUsernameOrPasswordErr.Code)`
- Disabled account: `response.Error(ctx, http.StatusForbidden, apperrors.ErrAccountDisabled.Code)`
- Not logged in: `response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)`

