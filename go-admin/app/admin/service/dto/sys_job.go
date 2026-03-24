package dto

import (
	"go-admin/app/admin/models"

	"go-admin/common/dto"
	common "go-admin/common/models"
)

type SysJobGetPageReq struct {
	dto.Pagination `search:"-"`
	JobId          int    `form:"jobId" search:"type:exact;column:job_id;table:sys_job" comment:"任务ID"`
	JobName        string `form:"jobName" search:"type:contains;column:job_name;table:sys_job" comment:"任务名称"`
	JobGroup       string `form:"jobGroup" search:"type:exact;column:job_group;table:sys_job" comment:"任务组"`
	CronExpression string `form:"cronExpression" search:"type:contains;column:cron_expression;table:sys_job" comment:"cron表达式"`
	InvokeTarget   string `form:"invokeTarget" search:"type:contains;column:invoke_target;table:sys_job" comment:"调用目标"`
	Status         int    `form:"status" search:"type:exact;column:status;table:sys_job" comment:"状态"`
}

func (m *SysJobGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type SysJobById struct {
	dto.ObjectById
	common.ControlBy
}

func (s *SysJobById) GetId() interface{} {
	if len(s.Ids) > 0 {
		s.Ids = append(s.Ids, s.Id)
		return s.Ids
	}
	return s.Id
}

func (s *SysJobById) GenerateM() (common.ActiveRecord, error) {
	return &models.SysJob{}, nil
}

type SysJobInsertReq struct {
	JobId          int    `json:"jobId" comment:"任务ID"`
	JobName        string `json:"jobName" comment:"任务名称" vd:"len($)>0"`
	JobGroup       string `json:"jobGroup" comment:"任务组"`
	JobType        int    `json:"jobType" comment:"任务类型"`
	CronExpression string `json:"cronExpression" comment:"cron表达式" vd:"len($)>0"`
	InvokeTarget   string `json:"invokeTarget" comment:"调用目标" vd:"len($)>0"`
	Args           string `json:"args" comment:"参数"`
	MisfirePolicy  int    `json:"misfirePolicy" comment:"执行策略"`
	Concurrent     int    `json:"concurrent" comment:"是否并发"`
	Status         int    `json:"status" comment:"状态"`
	EntryId        int    `json:"entryId" comment:"启动时任务ID"`
	common.ControlBy
}

func (s *SysJobInsertReq) Generate(model *models.SysJob) {
	if s.JobId != 0 {
		model.JobId = s.JobId
	}
	model.JobName = s.JobName
	model.JobGroup = s.JobGroup
	model.JobType = s.JobType
	model.CronExpression = s.CronExpression
	model.InvokeTarget = s.InvokeTarget
	model.Args = s.Args
	model.MisfirePolicy = s.MisfirePolicy
	model.Concurrent = s.Concurrent
	model.Status = s.Status
	model.EntryId = s.EntryId
	model.CreateBy = s.CreateBy
}

func (s *SysJobInsertReq) GetId() interface{} {
	return s.JobId
}

type SysJobUpdateReq struct {
	JobId          int    `json:"jobId" comment:"任务ID" vd:"$>0"`
	JobName        string `json:"jobName" comment:"任务名称" vd:"len($)>0"`
	JobGroup       string `json:"jobGroup" comment:"任务组"`
	JobType        int    `json:"jobType" comment:"任务类型"`
	CronExpression string `json:"cronExpression" comment:"cron表达式" vd:"len($)>0"`
	InvokeTarget   string `json:"invokeTarget" comment:"调用目标" vd:"len($)>0"`
	Args           string `json:"args" comment:"参数"`
	MisfirePolicy  int    `json:"misfirePolicy" comment:"执行策略"`
	Concurrent     int    `json:"concurrent" comment:"是否并发"`
	Status         int    `json:"status" comment:"状态"`
	EntryId        int    `json:"entryId" comment:"启动时任务ID"`
	common.ControlBy
}

func (s *SysJobUpdateReq) Generate(model *models.SysJob) {
	if s.JobId != 0 {
		model.JobId = s.JobId
	}
	model.JobName = s.JobName
	model.JobGroup = s.JobGroup
	model.JobType = s.JobType
	model.CronExpression = s.CronExpression
	model.InvokeTarget = s.InvokeTarget
	model.Args = s.Args
	model.MisfirePolicy = s.MisfirePolicy
	model.Concurrent = s.Concurrent
	model.Status = s.Status
	model.EntryId = s.EntryId
	model.UpdateBy = s.UpdateBy
}

func (s *SysJobUpdateReq) GetId() interface{} {
	return s.JobId
}

// SysJobByIds 批量删除请求
type SysJobByIds struct {
	dto.IdsReq
	common.ControlBy
}

func (s *SysJobByIds) GetIds() interface{} {
	return s.Ids
}

// UpdateSysJobStatusReq 更新状态请求
type UpdateSysJobStatusReq struct {
	JobId  int `json:"jobId" comment:"任务ID" vd:"$>0"`
	Status int `json:"status" comment:"状态"`
	common.ControlBy
}

func (s *UpdateSysJobStatusReq) GetId() interface{} {
	return s.JobId
}
