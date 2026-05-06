package audit

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBuildMeta_CreateMapsContractFields(t *testing.T) {
	entry := Entry{
		Title:  "岗位管理",
		Action: ActionCreate,
		Target: Target{Type: CategoryPost, ID: 42, Label: "运维"},
		Method: "admin.sysPost.insert",
		After:  map[string]interface{}{"postName": "运维", "status": 2},
	}

	meta := BuildMeta(entry)

	if meta.Title != "岗位管理" {
		t.Fatalf("Title not mapped: %q", meta.Title)
	}
	if meta.BusinessType != ActionCreate {
		t.Fatalf("BusinessType not mapped: %q", meta.BusinessType)
	}
	if meta.BusinessTypes != CategoryPost {
		t.Fatalf("BusinessTypes not mapped: %q", meta.BusinessTypes)
	}
	if meta.Method != "admin.sysPost.insert" {
		t.Fatalf("Method not mapped: %q", meta.Method)
	}
	if meta.OperatorType != OperatorManage {
		t.Fatalf("OperatorType should default to %q, got %q", OperatorManage, meta.OperatorType)
	}
	if meta.Remark == "" {
		t.Fatalf("Remark should summarise target, got empty")
	}

	var payload struct {
		Target struct {
			Type  string      `json:"type"`
			ID    interface{} `json:"id"`
			Label string      `json:"label"`
		} `json:"target"`
		Before interface{}            `json:"before,omitempty"`
		After  map[string]interface{} `json:"after,omitempty"`
		Extra  map[string]interface{} `json:"extra,omitempty"`
	}
	if err := json.Unmarshal([]byte(meta.OperParam), &payload); err != nil {
		t.Fatalf("OperParam not valid JSON: %v (%s)", err, meta.OperParam)
	}
	if payload.Target.Type != CategoryPost || payload.Target.Label != "运维" {
		t.Fatalf("OperParam target wrong: %+v", payload.Target)
	}
	if payload.Before != nil {
		t.Fatalf("Before should be omitted on create, got %v", payload.Before)
	}
	if payload.After["postName"] != "运维" {
		t.Fatalf("After.postName not preserved: %+v", payload.After)
	}
}

func TestBuildMeta_UpdateCarriesBeforeAndAfter(t *testing.T) {
	entry := Entry{
		Title:  "岗位管理",
		Action: ActionUpdate,
		Target: Target{Type: CategoryPost, ID: 42, Label: "运维"},
		Method: "admin.sysPost.update",
		Before: map[string]interface{}{"status": 2},
		After:  map[string]interface{}{"status": 1},
		Extra:  map[string]interface{}{"reason": "下线"},
	}

	meta := BuildMeta(entry)

	var payload struct {
		Before map[string]interface{} `json:"before"`
		After  map[string]interface{} `json:"after"`
		Extra  map[string]interface{} `json:"extra"`
	}
	if err := json.Unmarshal([]byte(meta.OperParam), &payload); err != nil {
		t.Fatalf("OperParam JSON: %v", err)
	}
	if payload.Before["status"].(float64) != 2 {
		t.Fatalf("Before.status not preserved: %+v", payload.Before)
	}
	if payload.After["status"].(float64) != 1 {
		t.Fatalf("After.status not preserved: %+v", payload.After)
	}
	if payload.Extra["reason"] != "下线" {
		t.Fatalf("Extra.reason not preserved: %+v", payload.Extra)
	}
}

func TestBuildMeta_DeleteOmitsAfter(t *testing.T) {
	entry := Entry{
		Title:  "岗位管理",
		Action: ActionDelete,
		Target: Target{Type: CategoryPost, ID: []int{1, 2}, Label: ""},
		Method: "admin.sysPost.delete",
		Before: map[string]interface{}{"ids": []int{1, 2}},
	}

	meta := BuildMeta(entry)

	if meta.BusinessType != ActionDelete {
		t.Fatalf("BusinessType: %q", meta.BusinessType)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(meta.OperParam), &payload); err != nil {
		t.Fatalf("OperParam JSON: %v", err)
	}
	if _, hasAfter := payload["after"]; hasAfter {
		t.Fatalf("After should be omitted on delete, got %+v", payload["after"])
	}
	if _, hasBefore := payload["before"]; !hasBefore {
		t.Fatalf("Before should be present on delete")
	}
}

func TestLog_NilContextIsNoOp(t *testing.T) {
	// 应该不 panic、不写入。
	Log(nil, Entry{Title: "x", Action: ActionCreate, Target: Target{Type: "t"}})
}

func TestLog_WritesAllFieldsIntoGinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	LogUpdate(c, "岗位管理",
		Target{Type: CategoryPost, ID: 7, Label: "运维"},
		map[string]interface{}{"status": 2},
		map[string]interface{}{"status": 1},
		"admin.sysPost.update",
	)

	wantPairs := map[string]string{
		TitleKey:         "岗位管理",
		BusinessTypeKey:  ActionUpdate,
		BusinessTypesKey: CategoryPost,
		MethodKey:        "admin.sysPost.update",
		OperatorTypeKey:  OperatorManage,
	}
	for k, want := range wantPairs {
		got, ok := c.Get(k)
		if !ok {
			t.Fatalf("ctx missing key %q", k)
		}
		if got != want {
			t.Fatalf("ctx[%q] = %v, want %v", k, got, want)
		}
	}

	op, _ := c.Get(OperParamKey)
	if op == nil {
		t.Fatalf("OperParam missing from ctx")
	}
	var payload struct {
		Target struct {
			ID float64 `json:"id"`
		} `json:"target"`
	}
	if err := json.Unmarshal([]byte(op.(string)), &payload); err != nil {
		t.Fatalf("OperParam JSON: %v", err)
	}
	if payload.Target.ID != 7 {
		t.Fatalf("Target.ID not 7: %v", payload.Target.ID)
	}
}

func TestBuildMeta_RespectsExplicitOperator(t *testing.T) {
	meta := BuildMeta(Entry{
		Title:    "审批",
		Action:   ActionApprove,
		Target:   Target{Type: CategoryWorkflow, ID: 1},
		Operator: "USER",
	})
	if meta.OperatorType != "USER" {
		t.Fatalf("OperatorType not respected: %q", meta.OperatorType)
	}
}
