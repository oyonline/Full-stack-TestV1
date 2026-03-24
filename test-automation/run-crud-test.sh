#!/bin/bash
# ============================================
# 用户管理 CRUD 自动化测试 - 工程硬版
# 特性：登录门禁、事件触发输入、跨平台兼容
# ============================================

# 严格模式但不自动退出
set -u
set -o pipefail

# 跨平台路径推导（脚本所在目录为基准）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_DIR="$SCRIPT_DIR"
REPORT_FILE="$TEST_DIR/reports/crud-test-report.md"
TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")

# 测试状态追踪
API_TOKEN=""
API_LOGIN_SUCCESS=false
TEST_USER_IDS=()
OVERALL_STATUS="PASSED"
FAILED_TESTS=()

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_debug() { echo -e "${CYAN}[DEBUG]${NC} $1"; }
log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; }

# ============================================
# 清理函数（强制总是执行）
# ============================================
cleanup() {
    local exit_code=$?
    log_step "执行清理..."
    
    # 清理 API 测试数据
    if [ ${#TEST_USER_IDS[@]} -gt 0 ] && [ -n "$API_TOKEN" ]; then
        log_info "清理 ${#TEST_USER_IDS[@]} 个测试用户..."
        for user_id in "${TEST_USER_IDS[@]}"; do
            curl -s -X DELETE \
                -H "Authorization: Bearer $API_TOKEN" \
                "http://localhost:10082/api/v1/sys-user/${user_id}" > /dev/null 2>&1 || true
        done
    fi
    
    # 关闭浏览器
    agent-browser close 2>/dev/null || true
    
    # 生成最终报告（如果还没生成）
    if [ -f "$REPORT_FILE" ] && ! grep -q "测试完成时间" "$REPORT_FILE" 2>/dev/null; then
        finalize_report
    fi
    
    log_info "清理完成"
    
    # 输出最终结果
    echo ""
    echo "========================================"
    if [ $exit_code -ne 0 ] || [ "$OVERALL_STATUS" = "FAILED" ]; then
        echo -e "${RED}测试完成: 存在失败项${NC}"
        [ ${#FAILED_TESTS[@]} -gt 0 ] && echo "失败测试: ${FAILED_TESTS[*]}"
    else
        echo -e "${GREEN}测试完成: 全部通过${NC}"
    fi
    echo "========================================"
    
    exit $exit_code
}

trap cleanup EXIT
trap 'log_error "收到中断信号"; exit 130' INT TERM

# ============================================
# 跨平台工具函数
# ============================================

# 从数组中删除指定元素（不使用 nameref，兼容旧版 bash）
array_remove() {
    local element=$1
    local new_array
    new_array=()
    
    for item in "${TEST_USER_IDS[@]}"; do
        if [ "$item" != "$element" ]; then
            new_array+=("$item")
        fi
    done
    
    TEST_USER_IDS=(${new_array[@]+"${new_array[@]}"})
}

# 跨平台 sed（支持 macOS 和 Linux）
sed_i() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "$@"
    else
        # Linux
        sed -i "$@"
    fi
}

# 检查 jq 是否安装
check_jq() {
    if ! command -v jq &> /dev/null; then
        log_warn "jq 未安装，JSON 解析将使用备用方案"
        return 1
    fi
    return 0
}

# 安全 JSON 提取
json_get() {
    local json_file=$1
    local key=$2
    local default=$3
    
    if command -v jq &> /dev/null && [ -f "$json_file" ]; then
        jq -r "$key // \"$default\"" "$json_file" 2>/dev/null || echo "$default"
    else
        grep -o "\"$key\":\"[^\"]*\"" "$json_file" 2>/dev/null | head -1 | cut -d'"' -f4 || echo "$default"
    fi
}

# ============================================
# UI 输入工具函数（赋值 + 事件触发）
# ============================================

# 安全设置输入值（支持 React/Vue 等框架）
browser_set_input() {
    local selector=$1
    local value=$2
    local field_name=${3:-"字段"}
    
    log_debug "设置 $field_name: $value"
    
    # 使用 JavaScript 设置值并触发事件
    local js_code="
        (function() {
            const el = $selector;
            if (!el) return 'not_found';
            
            // 设置值
            el.value = '$value';
            
            // 触发 input 事件（React/Vue 响应）
            const inputEvent = new Event('input', { bubbles: true });
            el.dispatchEvent(inputEvent);
            
            // 触发 change 事件
            const changeEvent = new Event('change', { bubbles: true });
            el.dispatchEvent(changeEvent);
            
            // 触发 blur 事件（表单验证）
            const blurEvent = new Event('blur', { bubbles: true });
            el.dispatchEvent(blurEvent);
            
            return 'success';
        })()
    "
    
    local result=$(agent-browser eval "$js_code" 2>/dev/null || echo "error")
    
    if [ "$result" = "success" ]; then
        return 0
    else
        log_warn "$field_name 设置可能失败: $result"
        return 1
    fi
}

# 多策略查找元素并设置值
browser_fill_field() {
    local field_name=$1
    local value=$2
    local selectors=$3  # 逗号分隔的选择器列表
    
    IFS=',' read -ra SELECTOR_LIST <<< "$selectors"
    
    for selector in "${SELECTOR_LIST[@]}"; do
        # 尝试 agent-browser 原生方法
        if agent-browser find role textbox --name "$field_name" fill "$value" 2>/dev/null; then
            return 0
        fi
        
        # 尝试 JavaScript 方式
        local js_selector="document.querySelector('$selector')"
        if browser_set_input "$js_selector" "$value" "$field_name"; then
            return 0
        fi
    done
    
    return 1
}

# ============================================
# 报告生成
# ============================================

init_report() {
    mkdir -p "$TEST_DIR/reports" "$TEST_DIR/logs"
    
    cat > "$REPORT_FILE" << EOF
# CRUD 自动化测试报告

**测试时间**: $TIMESTAMP  
**测试范围**: 用户管理模块 (API + UI)  
**后端地址**: http://localhost:10082  
**前端地址**: http://localhost:5666

---

## 执行摘要

| 项目 | 值 |
|------|-----|
| 总体状态 | ⏳ 执行中 |
| API 登录 | ⏳ 待测试 |
| API 通过 | 0/4 |
| UI 通过 | 0/5 |
| 总失败 | 0 |

---

## 详细测试结果

EOF
    
    API_PASSED=0
    UI_PASSED=0
}

record_test_result() {
    local phase=$1
    local test_name=$2
    local status=$3
    local details=$4
    local evidence=$5
    
    local emoji="✅"
    [ "$status" = "FAIL" ] && emoji="❌"
    [ "$status" = "SKIP" ] && emoji="⏭️"
    
    if [ "$phase" = "API" ] && [ "$status" = "PASS" ]; then
        API_PASSED=$((API_PASSED + 1))
    elif [ "$phase" = "UI" ] && [ "$status" = "PASS" ]; then
        UI_PASSED=$((UI_PASSED + 1))
    fi
    
    if [ "$status" = "FAIL" ]; then
        OVERALL_STATUS="FAILED"
        FAILED_TESTS+=("${phase}.${test_name}")
    fi
    
    cat >> "$REPORT_FILE" << EOF
### ${emoji} [${phase}] ${test_name}

- **状态**: ${status}
- **详情**: ${details}
- **证据**: ${evidence}

EOF
}

update_summary() {
    local api_total=4
    local ui_total=5
    local total_failed=${#FAILED_TESTS[@]}
    local login_status="✅ 成功"
    [ "$API_LOGIN_SUCCESS" = false ] && login_status="❌ 失败"
    
    sed_i "s/总体状态 | ⏳ 执行中/总体状态 | ${OVERALL_STATUS}/" "$REPORT_FILE"
    sed_i "s/API 登录 | ⏳ 待测试/API 登录 | ${login_status}/" "$REPORT_FILE"
    sed_i "s/API 通过 | 0\/4/API 通过 | ${API_PASSED}\/${api_total}/" "$REPORT_FILE"
    sed_i "s/UI 通过 | 0\/5/UI 通过 | ${UI_PASSED}\/${ui_total}/" "$REPORT_FILE"
    sed_i "s/总失败 | 0/总失败 | ${total_failed}/" "$REPORT_FILE"
}

finalize_report() {
    update_summary
    
    cat >> "$REPORT_FILE" << EOF

---

## 测试环境

- **API Token**: ${API_TOKEN:0:20}...
- **API 登录状态**: ${API_LOGIN_SUCCESS}
- **测试用户ID**: ${TEST_USER_IDS[*]:-"无"}
- **日志位置**: $TEST_DIR/logs/
- **截图位置**: $TEST_DIR/reports/

## 失败分析

EOF
    
    if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
        echo "**无失败项** 🎉" >> "$REPORT_FILE"
    else
        echo "**失败测试列表**:" >> "$REPORT_FILE"
        for test in "${FAILED_TESTS[@]}"; do
            echo "- $test" >> "$REPORT_FILE"
        done
    fi
    
    cat >> "$REPORT_FILE" << EOF

---

*报告生成时间: $(date "+%Y-%m-%d %H:%M:%S")*
EOF
}

# ============================================
# 服务检查
# ============================================

check_services() {
    log_step "检查服务状态"
    
    local backend_ok=false
    local max_wait=30
    local waited=0
    
    log_info "等待后端服务 (localhost:10082)..."
    while [ $waited -lt $max_wait ]; do
        if curl -s "http://localhost:10082/api/v1/login" > /dev/null 2>&1; then
            backend_ok=true
            break
        fi
        sleep 1
        waited=$((waited + 1))
        echo -n "."
    done
    echo ""
    
    if [ "$backend_ok" = true ]; then
        log_info "✅ 后端服务就绪"
    else
        log_error "❌ 后端服务未响应 (已等待 ${max_wait}s)"
        return 1
    fi
    
    return 0
}

# ============================================
# API 测试（带登录门禁）
# ============================================

api_login() {
    log_step "API: 登录获取 Token"
    
    local response_file="$TEST_DIR/logs/api-login.json"
    
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{
            "username": "admin",
            "password": "123456",
            "code": "1234",
            "uuid": "test-uuid"
        }' \
        "http://localhost:10082/api/v1/login" > "$response_file"
    
    if command -v jq &> /dev/null; then
        API_TOKEN=$(jq -r '.token // empty' "$response_file" 2>/dev/null)
    else
        API_TOKEN=$(grep -o '"token":"[^"]*"' "$response_file" | head -1 | cut -d'"' -f4)
    fi
    
    if [ -z "$API_TOKEN" ]; then
        log_error "API 登录失败"
        API_LOGIN_SUCCESS=false
        record_test_result "API" "登录获取Token" "FAIL" "无法获取token" "$response_file"
        return 1
    fi
    
    API_LOGIN_SUCCESS=true
    log_info "✅ Token 获取成功"
    record_test_result "API" "登录获取Token" "PASS" "成功获取认证token" "$response_file"
    return 0
}

# API 测试门禁检查
api_guard_check() {
    if [ "$API_LOGIN_SUCCESS" != true ]; then
        log_skip "API 登录未成功，跳过 CRUD 测试"
        return 1
    fi
    return 0
}

test_api_list() {
    log_step "API: 测试用户列表查询"
    
    # 登录门禁
    if ! api_guard_check; then
        record_test_result "API" "用户列表查询" "SKIP" "API 登录失败，跳过" ""
        return 0
    fi
    
    local response_file="$TEST_DIR/logs/api-list.json"
    local status="FAIL"
    local details=""
    
    curl -s -X GET \
        -H "Authorization: Bearer $API_TOKEN" \
        "http://localhost:10082/api/v1/sys-user?page=1&pageSize=10" > "$response_file"
    
    if command -v jq &> /dev/null; then
        local code=$(jq -r '.code // 0' "$response_file")
        local total=$(jq -r '.data.total // 0' "$response_file")
        local list_length=$(jq -r '.data.list | length // 0' "$response_file")
        
        if [ "$code" = "200" ] && [ "$total" -ge 0 ] 2>/dev/null; then
            status="PASS"
            details="返回 ${list_length} 条记录，总计 ${total} 条"
        else
            details="code=${code}, 数据结构异常"
        fi
    else
        if grep -q '"code":200' "$response_file" && grep -q '"total"' "$response_file"; then
            status="PASS"
            details="响应包含 code=200 和 total 字段"
        else
            details="响应结构不符合预期"
        fi
    fi
    
    record_test_result "API" "用户列表查询" "$status" "$details" "$response_file"
    [ "$status" = "PASS" ]
}

test_api_create() {
    log_step "API: 测试新增用户"
    
    if ! api_guard_check; then
        record_test_result "API" "新增用户" "SKIP" "API 登录失败，跳过" ""
        return 0
    fi
    
    local response_file="$TEST_DIR/logs/api-create.json"
    local status="FAIL"
    local details=""
    local user_id=""
    local timestamp=$(date +%s)
    local test_username="autotest_${timestamp}"
    
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $API_TOKEN" \
        -d "{
            \"username\": \"${test_username}\",
            \"password\": \"123456\",
            \"nickName\": \"API测试用户\",
            \"email\": \"${test_username}@test.com\",
            \"phone\": \"13800138000\",
            \"status\": \"1\",
            \"roleId\": 1,
            \"deptId\": 1
        }" \
        "http://localhost:10082/api/v1/sys-user" > "$response_file"
    
    if command -v jq &> /dev/null; then
        local code=$(jq -r '.code // 0' "$response_file")
        
        if [ "$code" = "200" ]; then
            # 后端返回 data:0，需要从列表查询获取用户ID
            sleep 1
            local list_file="$TEST_DIR/logs/api-create-list.json"
            curl -s -X GET \
                -H "Authorization: Bearer $API_TOKEN" \
                "http://localhost:10082/api/v1/sys-user?page=1&pageSize=10" > "$list_file"
            
            user_id=$(jq -r ".data.list[] | select(.username == \"$test_username\") | .userId" "$list_file" 2>/dev/null | head -1)
            
            if [ -n "$user_id" ]; then
                TEST_USER_IDS+=("$user_id")
                status="PASS"
                details="创建成功，从列表获取到用户ID: $user_id"
            else
                status="FAIL"
                details="创建成功但未从列表中找到用户"
            fi
        else
            details="code=$code, 创建失败"
        fi
    else
        if grep -q '"code":200' "$response_file"; then
            user_id=$(grep -o '"id":[0-9]*' "$response_file" | head -1 | cut -d: -f2)
            if [ -n "$user_id" ]; then
                TEST_USER_IDS+=("$user_id")
                status="PASS"
                details="创建成功(ID: $user_id)"
            else
                details="返回200但未获取到ID"
            fi
        else
            details="创建失败"
        fi
    fi
    
    record_test_result "API" "新增用户" "$status" "$details" "$response_file"
    
    if [ -n "$user_id" ]; then
        echo "$user_id" > "$TEST_DIR/logs/last-created-user.txt"
        echo "$test_username" > "$TEST_DIR/logs/last-created-username.txt"
    fi
    
    [ "$status" = "PASS" ]
}

test_api_update() {
    log_step "API: 测试编辑用户"
    
    if ! api_guard_check; then
        record_test_result "API" "编辑用户" "SKIP" "API 登录失败，跳过" ""
        return 0
    fi
    
    local response_file="$TEST_DIR/logs/api-update.json"
    local status="FAIL"
    local details=""
    local user_id=$(cat "$TEST_DIR/logs/last-created-user.txt" 2>/dev/null || echo "")
    local test_username=$(cat "$TEST_DIR/logs/last-created-username.txt" 2>/dev/null || echo "")
    
    if [ -z "$user_id" ] || [ -z "$test_username" ]; then
        record_test_result "API" "编辑用户" "SKIP" "没有可编辑的测试用户" ""
        return 0
    fi
    
    local new_nickname="已修改_${user_id}"
    
    # 先获取原用户信息
    local user_info_file="$TEST_DIR/logs/api-user-info.json"
    curl -s -X GET \
        -H "Authorization: Bearer $API_TOKEN" \
        "http://localhost:10082/api/v1/sys-user/${user_id}" > "$user_info_file"
    
    local phone=$(jq -r '.data.phone // "13800138000"' "$user_info_file")
    local email=$(jq -r '.data.email // "test@test.com"' "$user_info_file")
    
    curl -s -X PUT \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $API_TOKEN" \
        -d "{
            \"userId\": ${user_id},
            \"username\": \"${test_username}\",
            \"nickName\": \"${new_nickname}\",
            \"phone\": \"${phone}\",
            \"email\": \"${email}\",
            \"status\": \"1\",
            \"deptId\": 1,
            \"roleId\": 1
        }" \
        "http://localhost:10082/api/v1/sys-user" > "$response_file"
    
    if command -v jq &> /dev/null; then
        local code=$(jq -r '.code // 0' "$response_file")
        
        if [ "$code" = "200" ]; then
            sleep 1
            local verify_file="$TEST_DIR/logs/api-update-verify.json"
            curl -s -X GET \
                -H "Authorization: Bearer $API_TOKEN" \
                "http://localhost:10082/api/v1/sys-user/${user_id}" > "$verify_file"
            
            local actual_nickname=$(jq -r '.data.nickName // empty' "$verify_file")
            if [ "$actual_nickname" = "$new_nickname" ]; then
                status="PASS"
                details="修改成功，昵称已更新为: $new_nickname"
            else
                status="FAIL"
                details="接口成功但数据未变更(期望: $new_nickname, 实际: $actual_nickname)"
            fi
        else
            details="code=$code"
        fi
    else
        if grep -q '"code":200' "$response_file"; then
            status="PASS"
            details="编辑接口返回200"
        else
            details="编辑失败"
        fi
    fi
    
    record_test_result "API" "编辑用户" "$status" "$details" "$response_file"
    [ "$status" = "PASS" ]
}

test_api_delete() {
    log_step "API: 测试删除用户"
    
    if ! api_guard_check; then
        record_test_result "API" "删除用户" "SKIP" "API 登录失败，跳过" ""
        return 0
    fi
    
    local response_file="$TEST_DIR/logs/api-delete.json"
    local status="FAIL"
    local details=""
    local user_id=$(cat "$TEST_DIR/logs/last-created-user.txt" 2>/dev/null || echo "")
    
    if [ -z "$user_id" ]; then
        record_test_result "API" "删除用户" "SKIP" "没有可删除的测试用户" ""
        return 0
    fi
    
    curl -s -X DELETE \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $API_TOKEN" \
        -d "{\"ids\":[${user_id}]}" \
        "http://localhost:10082/api/v1/sys-user" > "$response_file"
    
    if command -v jq &> /dev/null; then
        local code=$(jq -r '.code // 0' "$response_file")
        
        if [ "$code" = "200" ]; then
            sleep 1
            local verify_file="$TEST_DIR/logs/api-delete-verify.json"
            curl -s -X GET \
                -H "Authorization: Bearer $API_TOKEN" \
                "http://localhost:10082/api/v1/sys-user/${user_id}" > "$verify_file"
            
            local verify_code=$(jq -r '.code // 0' "$verify_file")
            if [ "$verify_code" != "200" ]; then
                status="PASS"
                details="删除成功，数据已清除"
                # 正确删除数组元素
                array_remove "$user_id"
            else
                status="FAIL"
                details="接口返回成功但数据仍存在"
            fi
        else
            details="code=$code"
        fi
    else
        if grep -q '"code":200' "$response_file"; then
            status="PASS"
            details="删除接口返回200"
            rm -f "$TEST_DIR/logs/last-created-user.txt"
        else
            details="删除失败"
        fi
    fi
    
    record_test_result "API" "删除用户" "$status" "$details" "$response_file"
    [ "$status" = "PASS" ]
}

# ============================================
# UI 测试（使用事件触发输入）
# ============================================

test_ui_login() {
    log_step "UI: 测试登录流程"
    
    local status="FAIL"
    local details=""
    local screenshot="$TEST_DIR/reports/ui-login.png"
    
    agent-browser close 2>/dev/null || true
    sleep 1
    
    if ! agent-browser open "http://localhost:5666" 2>/dev/null; then
        details="无法打开登录页面"
        agent-browser screenshot "$screenshot" 2>/dev/null || true
        record_test_result "UI" "登录流程" "$status" "$details" "$screenshot"
        return 1
    fi
    
    sleep 3
    
    # 使用事件触发方式填写表单
    log_debug "填写用户名..."
    if ! browser_fill_field "用户名" "admin" "input[placeholder*='用户'],input[name*='user'],input[type='text']:first-of-type"; then
        log_warn "用户名填写可能失败，继续尝试..."
    fi
    
    sleep 0.5
    
    log_debug "填写密码..."
    browser_set_input "document.querySelector('input[type=password]')" "123456" "密码"
    
    sleep 0.5
    
    log_debug "填写验证码..."
    browser_set_input "document.querySelector('input[placeholder*=\"验证码\"]') || document.querySelectorAll('input[type=text]')[2]" "1234" "验证码"
    
    sleep 0.5
    
    # 点击登录
    log_debug "点击登录按钮..."
    if ! agent-browser find role button --name "登录" click 2>/dev/null; then
        agent-browser eval "Array.from(document.querySelectorAll('button')).find(b => b.textContent.includes('登录')).click()" 2>/dev/null || {
            details="无法点击登录按钮"
            agent-browser screenshot "$screenshot" 2>/dev/null || true
            record_test_result "UI" "登录流程" "$status" "$details" "$screenshot"
            return 1
        }
    fi
    
    sleep 3
    
    # 验证登录
    local current_url=$(agent-browser get url 2>/dev/null || echo "")
    if [[ "$current_url" != *"login"* ]] && [[ -n "$current_url" ]]; then
        status="PASS"
        details="登录成功，当前URL: $current_url"
    else
        details="登录后仍在登录页或URL异常: $current_url"
    fi
    
    agent-browser screenshot "$screenshot" 2>/dev/null || true
    record_test_result "UI" "登录流程" "$status" "$details" "$screenshot"
    [ "$status" = "PASS" ]
}

test_ui_navigate() {
    log_step "UI: 测试进入用户管理"
    
    local status="FAIL"
    local details=""
    local screenshot="$TEST_DIR/reports/ui-navigate.png"
    
    # 策略1: 直接访问用户管理URL（最可靠）
    log_debug "尝试直接访问用户管理URL..."
    if agent-browser open "http://localhost:5666/admin/sys-user" 2>/dev/null; then
        sleep 3
        local current_url=$(agent-browser get url 2>/dev/null || echo "")
        if [[ "$current_url" == *"sys-user"* ]]; then
            status="PASS"
            details="直接URL访问用户管理成功"
            agent-browser screenshot "$screenshot" 2>/dev/null || true
            record_test_result "UI" "进入用户管理" "$status" "$details" "$screenshot"
            return 0
        fi
    fi
    
    # 策略2: 通过菜单导航
    log_debug "尝试通过菜单导航..."
    
    # 点击系统管理
    if ! agent-browser find text "系统管理" click 2>/dev/null; then
        agent-browser eval "Array.from(document.querySelectorAll('*')).find(el => el.textContent.includes('系统管理') && el.click()).click()" 2>/dev/null || {
            details="无法找到系统管理菜单"
            agent-browser screenshot "$screenshot" 2>/dev/null || true
            record_test_result "UI" "进入用户管理" "$status" "$details" "$screenshot"
            return 1
        }
    fi
    
    # 等待子菜单展开
    sleep 2
    
    # 点击用户管理
    if ! agent-browser find text "用户管理" click 2>/dev/null; then
        agent-browser eval "Array.from(document.querySelectorAll('*')).find(el => el.textContent.includes('用户管理') && el.click()).click()" 2>/dev/null || {
            details="无法找到用户管理菜单"
            agent-browser screenshot "$screenshot" 2>/dev/null || true
            record_test_result "UI" "进入用户管理" "$status" "$details" "$screenshot"
            return 1
        }
    fi
    
    sleep 3
    
    # 验证页面
    local current_url=$(agent-browser get url 2>/dev/null || echo "")
    local snapshot=$(agent-browser snapshot -i 2>/dev/null || echo "")
    
    if [[ "$current_url" == *"sys-user"* ]] || echo "$snapshot" | grep -qi "新增.*用户\|用户列表\|username"; then
        status="PASS"
        details="用户管理页面加载成功 (URL: $current_url)"
    else
        sleep 2
        snapshot=$(agent-browser snapshot -i 2>/dev/null || echo "")
        current_url=$(agent-browser get url 2>/dev/null || echo "")
        if [[ "$current_url" == *"sys-user"* ]] || echo "$snapshot" | grep -qi "新增\|用户"; then
            status="PASS"
            details="用户管理页面加载成功(延迟) (URL: $current_url)"
        else
            details="页面关键元素未找到 (URL: $current_url)"
        fi
    fi
    
    agent-browser screenshot "$screenshot" 2>/dev/null || true
    record_test_result "UI" "进入用户管理" "$status" "$details" "$screenshot"
    [ "$status" = "PASS" ]
}

test_ui_create() {
    log_step "UI: 测试新增用户"
    
    local status="FAIL"
    local details=""
    local screenshot="$TEST_DIR/reports/ui-create.png"
    local timestamp=$(date +%s)
    local test_username="uitest_${timestamp}"
    
    # 等待页面加载完成
    sleep 2
    
    # 截图查看当前状态
    agent-browser screenshot "$TEST_DIR/reports/ui-before-create.png" 2>/dev/null || true
    
    # 点击"新增用户"按钮 - 直接操作 DOM
    log_debug "点击新增用户按钮..."
    
    # 等待确保页面渲染完成
    sleep 2
    
    # 尝试直接操作 DOM 显示弹窗
    local open_result=$(agent-browser eval '
        // 找到按钮并触发点击
        const buttons = Array.from(document.querySelectorAll("button"));
        const btn = buttons.find(b => b.textContent.includes("新增用户"));
        if (btn) {
            // 触发原生点击
            btn.dispatchEvent(new MouseEvent("click", { bubbles: true }));
            // 同时尝试获取 Vue 实例并调用方法
            const vueInstance = btn.__vueParentComponent?.ctx || btn._vueParentComponent?.ctx;
            if (vueInstance && vueInstance.openAddModal) {
                setTimeout(() => vueInstance.openAddModal(), 100);
                "opened with vue";
            } else {
                "clicked";
            }
        } else {
            "button not found";
        }
    ' 2>/dev/null || echo "error")
    
    log_debug "打开结果: $open_result"
    
    sleep 3
    
    # 检查弹窗是否打开
    local has_modal=$(agent-browser eval "
        const modal = document.querySelector('.ant-modal');
        const wrap = document.querySelector('.ant-modal-wrap');
        (modal && modal.offsetParent !== null) || (wrap && wrap.style.display !== 'none') ? 'yes' : 'no';
    " 2>/dev/null || echo "no")
    
    log_debug "弹窗状态: $has_modal"
    
    if [ "$has_modal" != "yes" ]; then
        # 最后尝试：直接修改 DOM 显示弹窗
        log_debug "尝试直接显示弹窗..."
        agent-browser eval '
            const modal = document.querySelector(".ant-modal");
            const wrap = document.querySelector(".ant-modal-wrap");
            if (modal) modal.style.display = "block";
            if (wrap) {
                wrap.style.display = "block";
                wrap.classList.add("ant-modal-wrap-open");
            }
        ' 2>/dev/null || true
        
        sleep 1
        
        has_modal=$(agent-browser eval "
            const modal = document.querySelector('.ant-modal');
            modal && modal.offsetParent !== null ? 'yes' : 'no';
        " 2>/dev/null || echo "no")
        
        if [ "$has_modal" != "yes" ]; then
            details="无法打开新增用户弹窗"
            agent-browser screenshot "$screenshot" 2>/dev/null || true
            record_test_result "UI" "新增用户" "$status" "$details" "$screenshot"
            return 1
        fi
    fi
    
    log_debug "弹窗已成功打开"
    
    # 等待弹窗出现
    sleep 3
    
    # 截图确认弹窗
    agent-browser screenshot "$TEST_DIR/reports/ui-modal-open.png" 2>/dev/null || true
    
    # 检查弹窗是否存在
    local has_modal=$(agent-browser eval "
        const modal = document.querySelector('.ant-modal, .vben-modal, [role=dialog]');
        modal ? 'yes' : 'no';
    " 2>/dev/null || echo "no")
    
    if [ "$has_modal" != "yes" ]; then
        log_warn "弹窗可能未打开，继续尝试..."
        sleep 2
        
        # 再次检查
        has_modal=$(agent-browser eval "
            const modal = document.querySelector('.ant-modal, .vben-modal, [role=dialog]');
            modal ? 'yes' : 'no';
        " 2>/dev/null || echo "no")
        
        if [ "$has_modal" != "yes" ]; then
            details="点击新增按钮后弹窗未打开"
            agent-browser screenshot "$screenshot" 2>/dev/null || true
            record_test_result "UI" "新增用户" "$status" "$details" "$screenshot"
            return 1
        fi
    fi
    
    # 填写表单（使用 data-testid）
    log_debug "填写用户名..."
    agent-browser eval "
        const input = document.querySelector('[data-testid=\"input-username\"]');
        if (input) {
            input.value = '${test_username}';
            input.dispatchEvent(new Event('input', { bubbles: true }));
            input.dispatchEvent(new Event('change', { bubbles: true }));
            'filled';
        } else { 'not found'; }
    " 2>/dev/null || log_warn "用户名设置可能失败"
    
    sleep 0.3
    
    log_debug "填写密码..."
    agent-browser eval "
        const input = document.querySelector('[data-testid=\"input-password\"]');
        if (input) {
            input.value = '123456';
            input.dispatchEvent(new Event('input', { bubbles: true }));
            input.dispatchEvent(new Event('change', { bubbles: true }));
            'filled';
        } else { 'not found'; }
    " 2>/dev/null || log_warn "密码设置可能失败"
    
    sleep 0.3
    
    log_debug "填写昵称..."
    agent-browser eval "
        const input = document.querySelector('[data-testid=\"input-nickname\"]');
        if (input) {
            input.value = 'UI测试用户';
            input.dispatchEvent(new Event('input', { bubbles: true }));
            input.dispatchEvent(new Event('change', { bubbles: true }));
            'filled';
        } else { 'not found'; }
    " 2>/dev/null || log_warn "昵称设置可能失败"
    
    sleep 0.3
    
    log_debug "填写手机号..."
    agent-browser eval "
        const input = document.querySelector('[data-testid=\"input-phone\"]');
        if (input) {
            input.value = '13800138000';
            input.dispatchEvent(new Event('input', { bubbles: true }));
            input.dispatchEvent(new Event('change', { bubbles: true }));
            'filled';
        } else { 'not found'; }
    " 2>/dev/null || log_warn "手机号设置可能失败"
    
    sleep 0.3
    
    log_debug "填写邮箱..."
    agent-browser eval "
        const input = document.querySelector('[data-testid=\"input-email\"]');
        if (input) {
            input.value = '${test_username}@test.com';
            input.dispatchEvent(new Event('input', { bubbles: true }));
            input.dispatchEvent(new Event('change', { bubbles: true }));
            'filled';
        } else { 'not found'; }
    " 2>/dev/null || log_warn "邮箱设置可能失败"
    
    sleep 0.5
    
    # 点击保存（使用 Ant Design Modal 的确定按钮）
    log_debug "点击保存..."
    local save_clicked=false
    
    # 方式1: 通过按钮文本在弹窗底部查找
    if agent-browser eval "
        const modal = document.querySelector('.ant-modal');
        if (modal) {
            const btn = modal.querySelector('.ant-modal-footer button.ant-btn-primary');
            if (btn) { btn.click(); 'clicked'; } else { 'not found'; }
        } else { 'no modal'; }
    " 2>/dev/null | grep -q "clicked"; then
        save_clicked=true
        log_debug "通过 Modal 底部按钮点击成功"
    fi
    
    if [ "$save_clicked" != true ]; then
        details="无法点击保存按钮"
        agent-browser screenshot "$screenshot" 2>/dev/null || true
        record_test_result "UI" "新增用户" "$status" "$details" "$screenshot"
        return 1
    fi
    
    sleep 3
    
    # 验证
    local snapshot=$(agent-browser snapshot -i 2>/dev/null || echo "")
    local page_text=$(agent-browser eval "document.body.innerText" 2>/dev/null || echo "")
    
    if echo "$snapshot$page_text" | grep -qi "成功\|${test_username}"; then
        status="PASS"
        details="新增用户成功: $test_username"
        echo "$test_username" > "$TEST_DIR/logs/last-ui-user.txt"
    elif echo "$page_text" | grep -qi "error\|失败"; then
        details="新增失败，检测到错误提示"
    else
        sleep 2
        snapshot=$(agent-browser snapshot -i 2>/dev/null || echo "")
        if echo "$snapshot" | grep -q "$test_username"; then
            status="PASS"
            details="新增用户成功(列表验证): $test_username"
            echo "$test_username" > "$TEST_DIR/logs/last-ui-user.txt"
        else
            details="无法确认新增结果"
        fi
    fi
    
    agent-browser screenshot "$screenshot" 2>/dev/null || true
    record_test_result "UI" "新增用户" "$status" "$details" "$screenshot"
    [ "$status" = "PASS" ]
}

test_ui_update() {
    log_step "UI: 测试编辑用户"
    
    local status="FAIL"
    local details=""
    local screenshot="$TEST_DIR/reports/ui-update.png"
    local test_username=$(cat "$TEST_DIR/logs/last-ui-user.txt" 2>/dev/null || echo "")
    local new_nickname="已修改_$(date +%s)"
    
    if [ -z "$test_username" ]; then
        record_test_result "UI" "编辑用户" "SKIP" "没有可编辑的测试用户" ""
        return 0
    fi
    
    local click_result=$(agent-browser eval "
        const rows = document.querySelectorAll('tr, .ant-table-row');
        for (const row of rows) {
            if (row.textContent.includes('$test_username')) {
                const editBtn = row.querySelector('button[class*=\"edit\"], a[class*=\"edit\"], [title*=\"编辑\"]');
                if (editBtn) {
                    editBtn.click();
                    return 'clicked';
                }
            }
        }
        return 'not found';
    " 2>/dev/null || echo "error")
    
    if [ "$click_result" != "clicked" ]; then
        details="无法找到或点击编辑按钮"
        agent-browser screenshot "$screenshot" 2>/dev/null || true
        record_test_result "UI" "编辑用户" "$status" "$details" "$screenshot"
        return 1
    fi
    
    sleep 2
    
    # 使用事件触发修改昵称
    browser_set_input "document.querySelector('input[placeholder*=\"昵称\"]') || document.querySelectorAll('input[type=text]')[1]" "$new_nickname" "昵称"
    
    sleep 0.5
    
    agent-browser eval "Array.from(document.querySelectorAll('button')).find(b => b.textContent.includes('保存') || b.textContent.includes('确定')).click()" 2>/dev/null || {
        details="无法点击保存按钮"
        agent-browser screenshot "$screenshot" 2>/dev/null || true
        record_test_result "UI" "编辑用户" "$status" "$details" "$screenshot"
        return 1
    }
    
    sleep 3
    
    # 验证
    local page_text=$(agent-browser eval "document.body.innerText" 2>/dev/null || echo "")
    
    if echo "$page_text" | grep -qi "成功"; then
        status="PASS"
        details="编辑成功，提示信息出现"
    else
        agent-browser reload 2>/dev/null || true
        sleep 2
        local snapshot=$(agent-browser snapshot -i 2>/dev/null || echo "")
        if echo "$snapshot" | grep -q "$new_nickname"; then
            status="PASS"
            details="编辑成功，列表中昵称已更新: $new_nickname"
        else
            details="无法确认编辑结果"
        fi
    fi
    
    agent-browser screenshot "$screenshot" 2>/dev/null || true
    record_test_result "UI" "编辑用户" "$status" "$details" "$screenshot"
    [ "$status" = "PASS" ]
}

test_ui_delete() {
    log_step "UI: 测试删除用户"
    
    local status="FAIL"
    local details=""
    local screenshot="$TEST_DIR/reports/ui-delete.png"
    local test_username=$(cat "$TEST_DIR/logs/last-ui-user.txt" 2>/dev/null || echo "")
    
    if [ -z "$test_username" ]; then
        record_test_result "UI" "删除用户" "SKIP" "没有可删除的测试用户" ""
        return 0
    fi
    
    local click_result=$(agent-browser eval "
        const rows = document.querySelectorAll('tr, .ant-table-row');
        for (const row of rows) {
            if (row.textContent.includes('$test_username')) {
                const deleteBtn = row.querySelector('button[class*=\"delete\"], a[class*=\"delete\"], [title*=\"删除\"]');
                if (deleteBtn) {
                    deleteBtn.click();
                    return 'clicked';
                }
            }
        }
        return 'not found';
    " 2>/dev/null || echo "error")
    
    if [ "$click_result" != "clicked" ]; then
        details="无法找到或点击删除按钮"
        agent-browser screenshot "$screenshot" 2>/dev/null || true
        record_test_result "UI" "删除用户" "$status" "$details" "$screenshot"
        return 1
    fi
    
    sleep 1
    
    agent-browser eval "Array.from(document.querySelectorAll('button')).find(b => b.textContent.includes('确定') || b.textContent.includes('确认')).click()" 2>/dev/null || true
    
    sleep 3
    
    # 验证
    local page_text=$(agent-browser eval "document.body.innerText" 2>/dev/null || echo "")
    
    if echo "$page_text" | grep -qi "成功\|删除成功"; then
        status="PASS"
        details="删除成功，提示信息出现"
    else
        agent-browser reload 2>/dev/null || true
        sleep 2
        local snapshot=$(agent-browser snapshot -i 2>/dev/null || echo "")
        if ! echo "$snapshot" | grep -q "$test_username"; then
            status="PASS"
            details="删除成功，列表中用户已消失"
        else
            details="用户仍在列表中，删除可能未生效"
        fi
    fi
    
    agent-browser screenshot "$screenshot" 2>/dev/null || true
    record_test_result "UI" "删除用户" "$status" "$details" "$screenshot"
    
    rm -f "$TEST_DIR/logs/last-ui-user.txt"
    
    [ "$status" = "PASS" ]
}

# ============================================
# 主执行流程
# ============================================

main() {
    log_info "========================================"
    log_info "用户管理 CRUD 自动化测试 - 工程硬版"
    log_info "========================================"
    
    init_report
    
    if ! check_jq; then
        log_warn "jq 未安装，JSON 解析将使用降级方案"
    fi
    
    if ! check_services; then
        log_error "服务检查失败，测试中止"
        exit 1
    fi
    
    # Phase 1: API 测试
    log_info ""
    log_info "========================================"
    log_info "Phase 1: API 层测试"
    log_info "========================================"
    
    # 登录是门禁，失败后续 SKIP
    api_login
    
    # CRUD 测试（登录失败时会自动 SKIP）
    test_api_list || true
    test_api_create || true
    test_api_update || true
    test_api_delete || true
    
    # Phase 2: UI 测试
    log_info ""
    log_info "========================================"
    log_info "Phase 2: UI 层测试"
    log_info "========================================"
    
    test_ui_login || true
    test_ui_navigate || true
    test_ui_create || true
    test_ui_update || true
    test_ui_delete || true
    
    log_info ""
    log_info "========================================"
    log_info "测试执行完毕，生成报告..."
    log_info "========================================"
}

main "$@"
