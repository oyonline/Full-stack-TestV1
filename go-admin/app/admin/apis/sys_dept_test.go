package apis

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"go-admin/app/admin/service/dto"
)

// TestRoleDeptTreeRespShape 锁定 GetDeptTreeRoleSelect 的响应 shape。
// 唯一调用方（vue-vben-admin/apps/web-antd/src/views/admin/sys-role/workspace.vue
// 通过 getRoleDeptTreeselect）依赖 depts + checkedKeys 成对返回；任何字段名
// 变更都会让前端 deptCheckedKeys / rawDeptTree 渲染失效。本测试若失败，
// 说明你正在改一个有外部契约的 shape，请同步修改前端。
func TestRoleDeptTreeRespShape(t *testing.T) {
	resp := roleDeptTreeResp{
		Depts: []dto.DeptLabel{
			{Id: 1, Label: "总部", Children: []dto.DeptLabel{}},
		},
		CheckedKeys: []int{1, 2, 3},
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	gotKeys := make([]string, 0, len(decoded))
	for k := range decoded {
		gotKeys = append(gotKeys, k)
	}
	sort.Strings(gotKeys)
	wantKeys := []string{"checkedKeys", "depts"}
	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("response keys mismatch: got %v want %v", gotKeys, wantKeys)
	}

	// CheckedKeys 必须是 JSON 数组（不能因 nil 退化为 null，前端会崩）
	var checked []int
	if err := json.Unmarshal(decoded["checkedKeys"], &checked); err != nil {
		t.Fatalf("checkedKeys not an int array: %v", err)
	}
	if !reflect.DeepEqual(checked, []int{1, 2, 3}) {
		t.Fatalf("checkedKeys content: got %v want [1 2 3]", checked)
	}
}

// TestRoleDeptTreeRespEmptyCheckedKeys 验证未传 roleId（id=0 分支）时
// CheckedKeys 仍序列化为 []，避免前端拿到 null 引发 .map of undefined。
func TestRoleDeptTreeRespEmptyCheckedKeys(t *testing.T) {
	resp := roleDeptTreeResp{
		Depts:       []dto.DeptLabel{},
		CheckedKeys: make([]int, 0),
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if got, want := string(raw), `{"depts":[],"checkedKeys":[]}`; got != want {
		t.Fatalf("empty payload mismatch:\ngot:  %s\nwant: %s", got, want)
	}
}
