package dto

import (
	"go-admin/common/dto"
)

type FeeRequestPageReq struct {
	dto.Pagination  `search:"-"`
	Status          string `json:"status" search:"column:status;type:exact;table:fee_request_log"`
	OrgCode         string `json:"orgCode" search:"column:org_code;type:exact;table:fee_request_log" comment:"付款主体"`
	YearsMonth      string `json:"yearsMonth" search:"column:budget_years_month;type:exact;table:fee_request_log" comment:"预算年月"`
	CostCenterId    int64  `json:"costCenterId" search:"column:cost_center_info_id;type:exact;table:fee_request_log" comment:"成本中心ID"`
	CostCenterName  string `json:"costCenterName" search:"column:cost_center_name;type:contains;table:fee_request_log" comment:"成本中心名称"`
	BudgetVersionId int64  `json:"budgetVersionId" search:"column:budget_version_id;type:exact;table:fee_request_log" comment:"预算版本ID"`
	GroupName       string `json:"groupName" search:"column:group_name;type:contains;table:fee_request_log" comment:"客户分组名称"`
	GroupCode       string `json:"groupCode" search:"column:group_number;type:exact;table:fee_request_log" comment:"客服分组编码"`
	BudgetDetailId  int64  `json:"budgetDetailId" search:"column:budget_detail_id;type:exact;table:fee_request_log" comment:"预算详情ID"`
	FeeCode         string `json:"feeCode" search:"column:fee_code;type:exact;table:fee_request_log" comment:"费用编码"`
	FeeName         string `json:"feeName" search:"column:fee_name;type:contains;table:fee_request_log" comment:"费用名称"`
	UserDeptId      int    `json:"userDeptId" search:"column:user_dept_id;type:exact;table:fee_request_log" comment:"申请部门ID"`
	DepartmentId    int    `json:"departmentId" search:"column:dept_id;type:exact;table:fee_request_log" comment:"承担部门ID"`
	RequestTimeFrom string `json:"requestTimeFrom" search:"column:request_time;type:gte;table:fee_request_log" comment:"费用申请时间"`
	RequestTimeTo   string `json:"requestTimeTo" search:"column:request_time;type:lte;table:fee_request_log" comment:"费用申请时间"`
}

func (m *FeeRequestPageReq) GetNeedSearch() interface{} {
	return *m
}

type FeeRequestGet struct {
	Id int64 `uri:"id"`
}
