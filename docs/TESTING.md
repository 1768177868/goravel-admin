# 单元测试指南

本项目已添加前后端单元测试支持，以提升代码质量和可维护性。

## 后端测试 (Go - Goravel)

后端测试分为两类：

- **快速测试**：默认质量门禁，只运行当前根模块内的包，不能依赖外部服务。
- **集成测试 / 冒烟测试**：需要数据库等依赖；CI 的 `backend-integration` job 会跑 `./tests/feature/...`。

### 当前实际结构

```
tests/
├── test_case.go                      # Goravel 测试基类
└── feature/
    ├── admin_auth_test.go            # 登录校验、未授权访问
    ├── admin_smoke_test.go           # 冒烟：登录成功、info、menus/tree
    └── api_auth_test.go              # health、用户注册/登录校验

app/http/helpers/
└── time_converter_test.go            # 快速 Go 测试

app/utils/
├── sharding_helper_test.go
└── traceid/traceid_test.go

driver/
├── kafka/queue_test.go               # 集成测试：需要本地 Kafka
├── nsq/queue_test.go                 # 集成测试：需要本地 nsqd
├── rabbitmq/queue_test.go            # 集成测试：需要本地 RabbitMQ
├── redisstream/queue_test.go         # 驱动子模块测试，使用 miniredis
└── dm/integration_test.go            # 集成测试：需要 -tags dm 和 DM_TEST_DSN
```

### 运行测试

```bash
# 快速后端门禁（CI 默认使用，排除 tests/feature）
PKGS=$(go list ./... | grep -v '/tests/feature$')
go test -v -timeout=2m $PKGS

# 冒烟 / feature 集成测试（需 migrate 后的数据库）
go test -v -timeout=2m ./tests/feature/...

# 仅 helper / utils
go test -v -timeout=30s ./app/http/helpers ./app/utils/...
```

### 集成测试（按需运行）

`driver/*` 下的 Kafka、NSQ、RabbitMQ、Redis Stream 是独立 Go module，默认 `go test ./...` 不会进入这些目录。需要验证队列驱动时，进入对应目录单独运行：

```bash
(cd driver/redisstream && go test -v ./...)
(cd driver/kafka && go test -v -timeout=2m ./...)     # 需要 127.0.0.1:9092
(cd driver/nsq && go test -v -timeout=2m ./...)       # 需要 127.0.0.1:4150
(cd driver/rabbitmq && go test -v -timeout=2m ./...)  # 需要 127.0.0.1:5672
```

达梦驱动集成测试需要本地官方驱动、`dm` build tag 和可访问的达梦实例：

```bash
set DM_TEST_DSN=dm://SYSDBA:SYSDBA@127.0.0.1:5236
go test -tags dm ./driver/dm/... -run TestDMCrudAndTransaction -v
```

### 卡住问题排查顺序

如果 `go test ./...` 变慢或无输出，按下面顺序缩小范围：

```bash
# 1. 只编译，不运行测试，判断是否是包初始化/编译问题
go test -run TestNonExistent -count=0 -v ./...

# 2. 分区检查
go test -run TestNonExistent -count=0 -v ./app/...
go test -run TestNonExistent -count=0 -v ./routes ./config ./bootstrap ./database/...

# 3. 再运行真实快速测试
go test -v -timeout=2m ./...
```

注意：不要把需要外部服务的集成测试加入默认快速门禁；它们应独立运行并设置明确超时。

### 创建测试

使用 Artisan 命令创建测试：

```bash
go run . artisan make:test unit/MyServiceTest
go run . artisan make:test feature/MyFeatureTest
```

### 测试示例

```go
// tests/unit/token_service_test.go
package unit

import (
    "testing"
    "github.com/stretchr/testify/suite"
    "goravel/tests"
)

type TokenServiceTestSuite struct {
    suite.Suite
    tests.TestCase
}

func TestTokenServiceTestSuite(t *testing.T) {
    suite.Run(t, new(TokenServiceTestSuite))
}

func (s *TokenServiceTestSuite) SetupTest() {
    // 每个测试前执行
}

func (s *TokenServiceTestSuite) TearDownTest() {
    // 每个测试后执行
}

func (s *TokenServiceTestSuite) TestHashToken() {
    // 测试 token 哈希
    s.Equal(64, len(hashToken("test")))
}

func (s *TokenServiceTestSuite) TestGenerateRandomToken() {
    token1 := generateRandomToken()
    token2 := generateRandomToken()
    
    s.Len(token1, 40)
    s.NotEqual(token1, token2)
}
```

---

## 前端测试 (Vitest)

### 目录结构

前端测试支持两种组织方式：

#### 方式一：共置模式（当前使用）✅ 推荐

测试文件放在被测模块旁边的 `__tests__` 目录：

```
html/src/
├── composables/
│   ├── useCrud.js
│   ├── useDebounce.js
│   └── __tests__/              # 测试放在模块旁边
│       ├── useCrud.test.js
│       └── useDebounce.test.js
├── utils/
│   ├── validation.js
│   ├── storage.js
│   └── __tests__/
│       ├── validation.test.js
│       └── storage.test.js
└── components/
    └── __tests__/
        └── ...
```

**优点：**
- 测试文件与源码紧密关联，便于查找和维护
- 符合 Vue/React 社区最佳实践
- 模块独立性强，便于移动或删除

#### 方式二：集中模式

所有测试放在根目录的 `tests` 目录：

```
html/
├── src/
│   ├── composables/
│   ├── utils/
│   └── components/
└── tests/                      # 所有测试集中存放
    ├── unit/
    │   ├── composables/
    │   │   ├── useCrud.test.js
    │   │   └── useDebounce.test.js
    │   └── utils/
    │       ├── validation.test.js
    │       └── storage.test.js
    └── integration/
        └── ...
```

**优点：**
- 测试代码与源码分离
- 与后端测试结构一致

### 运行测试

```bash
cd html

# 交互式监听模式
npm test

# 单次运行
npm run test:run

# 生成覆盖率报告
npm run test:coverage
```

### 当前测试文件

| 测试文件 | 覆盖内容 | 测试数量 |
|----------|----------|----------|
| `composables/__tests__/useDebounce.test.js` | 防抖功能 | 12 |
| `composables/__tests__/useCrud.test.js` | CRUD 操作 | 19 |
| `utils/__tests__/validation.test.js` | 验证器函数 | 41 |
| `utils/__tests__/storage.test.js` | Storage 工具 | 13 |
| `utils/__tests__/markdown.test.js` | Markdown/HTML 净化 | 4 |

### 测试示例

```javascript
// src/utils/__tests__/validation.test.js
import { describe, it, expect } from 'vitest'
import { validators } from '../validation'

describe('validators', () => {
  describe('required', () => {
    it('应该拒绝空字符串', () => {
      expect(validators.required('')).not.toBe(true)
    })

    it('应该接受有效字符串', () => {
      expect(validators.required('hello')).toBe(true)
    })
  })

  describe('email', () => {
    it('应该接受有效邮箱', () => {
      expect(validators.email('test@example.com')).toBe(true)
    })
  })
})
```

---

## 测试最佳实践

### 1. 命名约定

```go
// Go: TestSuiteName + TestMethodName
func (s *TokenServiceTestSuite) TestHashToken_ValidInput()

// JavaScript: describe + it
describe('validators', () => {
  describe('required', () => {
    it('应该拒绝空字符串', () => {})
  })
})
```

### 2. 表格驱动测试 (Go)

```go
func (s *TokenServiceTestSuite) TestHashToken() {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"正常 token", "test-token", 64},
        {"空 token", "", 64},
    }
    
    for _, tt := range tests {
        s.Run(tt.name, func() {
            got := hashToken(tt.input)
            s.Len(got, tt.expected)
        })
    }
}
```

### 3. Mock 使用 (JavaScript)

```javascript
import { vi } from 'vitest'

// Mock 模块
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn()
  }
}))

// 验证调用
expect(ElMessage.success).toHaveBeenCalledWith('操作成功')
```

### 4. 异步测试

```javascript
// async/await
it('应该处理异步操作', async () => {
  const result = await asyncFunction()
  expect(result).toBe('expected')
})

// Fake timers
beforeEach(() => vi.useFakeTimers())
afterEach(() => vi.restoreAllMocks())
```

---

## CI/CD 集成

### GitHub Actions 示例

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - run: go test -v ./tests/...

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: cd html && npm install
      - run: cd html && npm run test:run
```

---

## 测试覆盖率目标

| 模块类型 | 当前状态 | 目标覆盖率 |
|----------|----------|------------|
| 工具函数 | ✅ 已添加 | 80%+ |
| Composables | ✅ 已添加 (TypeScript) | 70%+ |
| Services | ✅ 已添加 | 60%+ |
| Controllers | ✅ 已添加（集成测试） | 40%+ |

---

## Controller 集成测试

后端 Controller 集成测试位于 `tests/feature/` 目录：

### 测试文件

| 文件 | 覆盖内容 |
|------|----------|
| `admin_api_test.go` | 管理员登录、信息、列表、角色、菜单、部门、日志等 |
| `blacklist_api_test.go` | 黑名单 CRUD、IP格式验证、批量删除 |
| `permission_test.go` | 权限列表、角色权限绑定 |

### 运行集成测试

```bash
# 运行所有集成测试（需要 Docker）
go test -v ./tests/feature/...

# 运行特定测试
go test -v -run TestAdminApiTestSuite ./tests/feature/...
go test -v -run TestBlacklistApiTestSuite ./tests/feature/...
```

### 测试示例

```go
func (s *AdminApiTestSuite) TestLogin_Success() {
    body := strings.NewReader(`{"username":"admin","password":"admin123"}`)
    resp, err := s.Http(s.T()).
        WithHeader("Content-Type", "application/json").
        Post("/api/admin/login", body)

    s.Require().NoError(err)
    resp.AssertSuccessful()

    content, err := resp.Content()
    s.Require().NoError(err)

    var result map[string]any
    json.Unmarshal([]byte(content), &result)
    s.Equal(float64(200), result["code"])
}
```

### 注意事项

- 集成测试需要 Docker 环境（自动创建测试数据库和 Redis）
- 每个测试会执行 `RefreshDatabase()` 重置数据库
- 使用 `Seed()` 填充测试数据

---

## 常见问题

### Q: 前端测试应该放在哪里？

**A:** 推荐使用共置模式（`__tests__` 目录），这是 Vue/React 社区的最佳实践。

### Q: 后端测试为什么放在 tests 目录？

**A:** 这是 Goravel 框架的推荐做法，参考 [官方文档](https://www.goravel.dev/zh_CN/testing/getting-started.html)。

### Q: 如何添加新测试？

**A:** 
- 后端：`go run . artisan make:test unit/MyTest`
- 前端：在对应模块的 `__tests__` 目录创建 `.test.js` 文件
