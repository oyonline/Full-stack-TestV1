# 用户管理 CRUD 自动化测试 - 健壮版

## 核心特性

| 特性 | 实现方式 |
|------|---------|
| **失败不中断** | 移除 `set -e`，显式错误处理，每个测试 `|| true` 继续执行 |
| **强制清理** | `trap cleanup EXIT` 捕获所有退出，确保数据清理 |
| **JSON结构化解析** | 优先 `jq` 解析，降级到字符串匹配 |
| **结果校验** | 编辑/删除后验证数据实际变更，不只是接口返回 |

---

## 执行步骤

### 1. 确保服务运行

```bash
# 终端1: 后端
cd ~/Desktop/Full-stack-TestV1/go-admin
go run main.go

# 终端2: 前端
cd ~/Desktop/Full-stack-TestV1/vue-vben-admin
pnpm dev
```

### 2. 运行测试

```bash
cd ~/Desktop/Full-stack-TestV1/test-automation
./run-crud-test.sh
```

### 3. 查看结果

```bash
# 测试报告
cat reports/crud-test-report.md

# API 响应详情
cat logs/api-*.json

# UI 截图
open reports/*.png
```

---

## 测试覆盖

### API 层 (4项)

| 测试 | 验证内容 |
|------|---------|
| 登录获取Token | Bearer Token 提取 |
| 用户列表查询 | `code=200`, `data.total`, `data.list` 结构校验 |
| 新增用户 | 接口响应 + **入库验证**（查询确认） |
| 编辑用户 | 接口响应 + **变更验证**（昵称修改确认） |
| 删除用户 | 接口响应 + **清除验证**（查询返回非200） |

### UI 层 (5项)

| 测试 | 验证内容 |
|------|---------|
| 登录流程 | URL 跳转验证 + 截图 |
| 进入用户管理 | 页面元素检测 + 截图 |
| 新增用户 | 表单提交 + **成功提示/列表验证** + 截图 |
| 编辑用户 | 查找用户 + 修改 + **刷新验证** + 截图 |
| 删除用户 | 查找用户 + 确认 + **消失验证** + 截图 |

---

## 技术实现

### 1. 失败不中断机制

```bash
# 不用 set -e
set -u
set -o pipefail

# 每个测试函数返回状态
test_api_list() {
    # ... 测试逻辑 ...
    [ "$status" = "PASS" ]  # 返回 0 或 1
}

# main 中统一处理，失败继续
test_api_list || true
test_api_create || true
# ... 即使失败，后续测试仍会执行
```

### 2. 强制清理机制

```bash
# EXIT trap 捕获所有退出情况
trap cleanup EXIT

cleanup() {
    # 总是执行，无论正常退出还是异常
    # 1. 清理所有测试用户
    # 2. 关闭浏览器
    # 3. 生成报告
}
```

### 3. JSON 结构化解析

```bash
# 优先使用 jq
if command -v jq &> /dev/null; then
    code=$(jq -r '.code // 0' "$response_file")
    total=$(jq -r '.data.total // 0' "$response_file")
    list_length=$(jq -r '.data.list | length // 0' "$response_file")
else
    # 降级到字符串匹配
    code=$(grep -o '"code":200' "$response_file" ...)
fi
```

### 4. 结果校验示例

**编辑验证：**
```bash
# 1. 调用编辑接口
curl -X PUT ... /api/v1/user/$user_id

# 2. 验证接口返回
code=$(jq -r '.code' "$response_file")

# 3. 查询验证数据实际变更
verify_response=$(curl -X GET ... /api/v1/user/$user_id)
actual_nickname=$(jq -r '.data.nickName' "$verify_file")

# 4. 对比期望值
if [ "$actual_nickname" = "$expected_nickname" ]; then
    status="PASS"
fi
```

**删除验证：**
```bash
# 1. 调用删除接口
curl -X DELETE ... /api/v1/user/$user_id

# 2. 再次查询
verify_response=$(curl -X GET ... /api/v1/user/$user_id)
verify_code=$(jq -r '.code' "$verify_file")

# 3. 验证返回非200（数据已删除）
if [ "$verify_code" != "200" ]; then
    status="PASS"
fi
```

---

## 文件结构

```
test-automation/
├── run-crud-test.sh         # 主测试脚本（唯一入口）
├── README.md                # 本文档
├── reports/                 # 测试报告和截图
│   ├── crud-test-report.md
│   ├── ui-login.png
│   ├── ui-navigate.png
│   ├── ui-create.png
│   ├── ui-update.png
│   └── ui-delete.png
└── logs/                    # 详细日志
    ├── api-login.json
    ├── api-list.json
    ├── api-list-verify.json
    ├── api-create.json
    ├── api-create-verify.json
    └── ...
```

---

## 报告格式

```markdown
# CRUD 自动化测试报告

**测试时间**: 2026-03-24 10:30:00
**测试范围**: 用户管理模块 (API + UI)

## 执行摘要

| 项目 | 值 |
|------|-----|
| 总体状态 | PASSED / FAILED |
| API 通过 | 4/4 |
| UI 通过 | 5/5 |
| 总失败 | 0 |

## 详细测试结果

### ✅ [API] 用户列表查询
- **状态**: PASS
- **详情**: 返回 10 条记录，总计 100 条
- **证据**: logs/api-list.json

### ❌ [API] 新增用户
- **状态**: FAIL
- **详情**: code=500, 数据库连接失败
- **证据**: logs/api-create.json

...

## 失败分析
- API.新增用户: 数据库连接异常
- UI.编辑用户: 无法定位编辑按钮
```

---

## 问题发现与修复流程

```
运行测试
   ↓
查看 reports/crud-test-report.md
   ↓
识别失败项
   ↓
查看 logs/api-*.json 或 reports/*.png
   ↓
定位根因（API路由？前端组件？）
   ↓
小七根据日志修复代码
   ↓
重新运行测试验证
```

---

## 扩展指南

### 添加新 API 测试

```bash
test_api_custom() {
    log_step "API: 自定义测试"
    
    local response_file="$TEST_DIR/logs/api-custom.json"
    local status="FAIL"
    local details=""
    
    # 发送请求
    curl -s -X GET ... > "$response_file"
    
    # 解析验证
    if command -v jq &> /dev/null; then
        code=$(jq -r '.code' "$response_file")
        # ... 验证逻辑 ...
    fi
    
    record_test_result "API" "自定义测试" "$status" "$details" "$response_file"
    [ "$status" = "PASS" ]
}

# 在 main() 中添加调用
test_api_custom || true
```

### 添加新 UI 测试

```bash
test_ui_custom() {
    log_step "UI: 自定义测试"
    
    local status="FAIL"
    local details=""
    local screenshot="$TEST_DIR/reports/ui-custom.png"
    
    # UI 操作
    agent-browser find text "菜单" click 2>/dev/null || {
        details="无法点击菜单"
        record_test_result "UI" "自定义测试" "$status" "$details" "$screenshot"
        return 1
    }
    
    # 验证
    # ... 验证逻辑 ...
    
    agent-browser screenshot "$screenshot" 2>/dev/null || true
    record_test_result "UI" "自定义测试" "$status" "$details" "$screenshot"
    [ "$status" = "PASS" ]
}

# 在 main() 中添加调用
test_ui_custom || true
```

---

## 注意事项

1. **验证码**: 当前使用固定值 `1234`，如需真实验证码需接入识别服务
2. **测试数据**: 所有用户带 `autotest_` 或 `uitest_` 前缀，便于识别
3. **并发**: 当前为顺序执行，避免数据冲突
4. **jq 依赖**: 如未安装会自动尝试安装，失败则使用降级方案
