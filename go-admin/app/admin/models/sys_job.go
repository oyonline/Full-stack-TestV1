package models

import "go-admin/common/models"

type SysJob struct {
	JobId          int    `gorm:"primaryKey;autoIncrement;comment:任务编码" json:"jobId"`       // 任务编码
	JobName        string `json:"jobName" gorm:"size:128;comment:任务名称"`                     // 任务名称
	JobGroup       string `json:"jobGroup" gorm:"size:128;comment:任务分组"`                    // 任务分组
	JobType        int    `json:"jobType" gorm:"size:4;comment:任务类型(1:HTTP)"`               // 任务类型(1:HTTP)
	CronExpression string `json:"cronExpression" gorm:"size:128;comment:cron表达式"`            // cron表达式
	InvokeTarget   string `json:"invokeTarget" gorm:"size:255;comment:调用目标"`                // 调用目标
	Args           string `json:"args" gorm:"size:255;comment:参数"`                            // 参数
	MisfirePolicy  int    `json:"misfirePolicy" gorm:"size:4;comment:错过策略"`                 // 错过策略
	Concurrent     int    `json:"concurrent" gorm:"size:4;comment:是否并发(1:允许,2:禁止)"`       // 是否并发
	Status         int    `json:"status" gorm:"size:4;comment:状态(1:停用,2:启用)"`             // 状态(1:停用,2:启用)
	EntryId        int    `json:"entryId" gorm:"size:20;comment:条目ID"`                      // 条目ID
	Remark         string `json:"remark" gorm:"size:255;comment:备注"`                          // 备注
	models.ControlBy
	models.ModelTime
}

func (*SysJob) TableName() string {
	return "sys_job"
}

func (e *SysJob) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SysJob) GetId() interface{} {
	return e.JobId
}
