# CRUD 自动化测试报告

**测试时间**: 2026-03-24 10:54:46  
**测试范围**: 用户管理模块 (API + UI)  
**后端地址**: http://localhost:10082  
**前端地址**: http://localhost:5666

---

## 执行摘要

| 项目 | 值 |
|------|-----|
| 总体状态 | FAILED |
| API 登录 | ✅ 成功 |
| API 通过 | 5/4 |
| UI 通过 | 2/5 |
| 总失败 | 1 |

---

## 详细测试结果

### ✅ [API] 登录获取Token

- **状态**: PASS
- **详情**: 成功获取认证token
- **证据**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/logs/api-login.json

### ✅ [API] 用户列表查询

- **状态**: PASS
- **详情**: 返回 7 条记录，总计 0 条
- **证据**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/logs/api-list.json

### ✅ [API] 新增用户

- **状态**: PASS
- **详情**: 创建成功，从列表获取到用户ID: 18
- **证据**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/logs/api-create.json

### ✅ [API] 编辑用户

- **状态**: PASS
- **详情**: 修改成功，昵称已更新为: 已修改_18
- **证据**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/logs/api-update.json

### ✅ [API] 删除用户

- **状态**: PASS
- **详情**: 删除成功，数据已清除
- **证据**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/logs/api-delete.json

### ✅ [UI] 登录流程

- **状态**: PASS
- **详情**: 登录成功，当前URL: http://localhost:5666/admin/sys-config/set
- **证据**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/reports/ui-login.png

### ✅ [UI] 进入用户管理

- **状态**: PASS
- **详情**: 直接URL访问用户管理成功
- **证据**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/reports/ui-navigate.png

### ❌ [UI] 新增用户

- **状态**: FAIL
- **详情**: 无法打开新增用户弹窗
- **证据**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/reports/ui-create.png

### ⏭️ [UI] 编辑用户

- **状态**: SKIP
- **详情**: 没有可编辑的测试用户
- **证据**: 

### ⏭️ [UI] 删除用户

- **状态**: SKIP
- **详情**: 没有可删除的测试用户
- **证据**: 


---

## 测试环境

- **API Token**: eyJhbGciOiJIUzI1NiIs...
- **API 登录状态**: true
- **测试用户ID**: 无
- **日志位置**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/logs/
- **截图位置**: /Users/linshen/Desktop/Full-stack-TestV1/test-automation/reports/

## 失败分析

**失败测试列表**:
- UI.新增用户

---

*报告生成时间: 2026-03-24 10:55:17*
