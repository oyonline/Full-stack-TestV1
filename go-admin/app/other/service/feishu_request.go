package service

import (
	"errors"
	"fmt"
	emod "go-admin/app/admin/models"
	esvc "go-admin/app/admin/service"
	"go-admin/app/other/service/dto"
	"go-admin/common/utils"
	"go-admin/common/utils/structsUtils"
	"strconv"
	"strings"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type FeishuRequest struct {
	service.Service
}

func (e *FeishuRequest) ProcessCallback(c *dto.FeishuApiResponse) error {
	if c.Code == 0 {
		isNewReq := true
		respData := c.Data
		appRecord := emod.FeishuApprovalRecord{}
		err := e.Orm.Model(&emod.FeishuApprovalRecord{}).Where("instance_code = ? AND approval_code = ?", respData.InstanceCode, respData.ApprovalCode).First(&appRecord).Error
		if err == nil && appRecord.ID > 0 {
			isNewReq = false
			appRecord.Status = respData.Status
		} else {
			structsUtils.CopyBeanProp(&appRecord, respData)
		}
		if respData.StartTime != "" && respData.StartTime != "0" {
			timeStamp, _ := strconv.ParseInt(respData.StartTime, 10, 64)
			seconds := timeStamp / 1000
			appRecord.StartDate = time.Unix(seconds, 0).Format(time.DateTime)
		}
		if respData.EndTime != "" && respData.EndTime != "0" {
			timeStamp, _ := strconv.ParseInt(respData.EndTime, 10, 64)
			seconds := timeStamp / 1000
			appRecord.EndDate = time.Unix(seconds, 0).Format(time.DateTime)
		}
		db := e.Orm.Begin()
		// 更新飞书实例数据
		err = db.Save(&appRecord).Error
		if err != nil {
			db.Rollback()
			return err
		}
		lines := respData.Timeline
		if len(lines) > 0 {
			// 更新审批流
			timeLines := make([]emod.FeishuTimeline, 0)
			for _, l := range lines {
				line := emod.FeishuTimeline{
					InstanceCode: appRecord.InstanceCode,
					Type:         l.Type,
					CreateDate:   time.Now().Format(time.DateTime),
					UserId:       l.UserId,
					OpenId:       l.OpenId,
					TaskId:       l.TaskId,
					Comment:      l.Comment,
					Ext:          l.Ext,
					NodeKey:      l.NodeKey,
				}
				if l.CreateTime != "" && l.CreateTime != "0" {
					timeStamp, _ := strconv.ParseInt(l.CreateTime, 10, 64)
					seconds := timeStamp / 1000
					line.CreateDate = time.Unix(seconds, 0).Format(time.DateTime)
				}
				timeLines = append(timeLines, line)
			}
			appRecord.Timeline = timeLines
			err = db.Save(&timeLines).Error
			if err != nil {
				db.Rollback()
				return err
			}
		}
		// 如果已经存在审批实例数据，说明是有状态变更
		if !isNewReq {
			var logs []emod.FeeRequestLog
			db.Model(&emod.FeeRequestLog{}).Where("instance_code = ?", appRecord.InstanceCode).Find(&logs)
			if len(logs) > 0 {
				budgets := make([]emod.CostBudgetVersionDetail, 0)
				// 更新费用申请日志状态
				for i, l := range logs {
					l.Status = appRecord.Status
					// 实例状态为 APPROVE 扣除对应的费用预算
					if appRecord.Status == "APPROVED" {
						var budget emod.CostBudgetVersionDetail
						e.Orm.Model(&budget).First(&budget, "id = ?", l.Id)
						if budget.Id > 0 {
							l.BudgetUsed += l.RequestAmount
							budget.BudgetUsed += l.RequestAmount
							budgets = append(budgets, budget)
						}
					}
					logs[i] = l
				}
				if len(budgets) > 0 {
					err = db.Save(&logs).Error
					if err != nil {
						db.Rollback()
						return err
					}
					err = db.Save(&budgets).Error
					if err != nil {
						db.Rollback()
						return err
					}
				}
			}
			db.Commit()
			return nil
		}
		// 新的审批实例 保存表单数据
		forms := respData.Form
		approvalForm := emod.FeishuApprovalForm{Currency: "CNY", InstanceCode: appRecord.InstanceCode}
		for _, f := range forms {
			switch f.Name {
			case "费用平台":
				val, _ := utils.ToString(f.Value)
				if f.Option != nil && f.Option.Key != "" {
					val = f.Option.Key
				}
				approvalForm.Platform = val
			case "付款主体":
				val, _ := utils.ToString(f.Value)
				if f.Option != nil && f.Option.Key != "" {
					val = f.Option.Key
				}
				approvalForm.OrgCode = val
			case "承担部门":
				val, _ := utils.ToString(f.Value)
				if f.Option != nil && f.Option.Key != "" {
					val = f.Option.Key
					deptArr := strings.Split(val, ":")
					val = deptArr[1]
					approvalForm.KingdeeDepartmentCode = deptArr[0]
				}
				approvalForm.DepartmentId = val
			case "费用汇总":
				val := utils.ToFloat64(f.Value)
				if f.Ext != nil {
					if utils.IsMap(f.Ext) {
						ext := f.Ext.(map[string]interface{})
						if ca, ok := ext["capitalValue"]; ok {
							approvalForm.Capital, _ = utils.ToString(ca)
						}
					}
				}
				approvalForm.TotalAmount = val
			case "附件":
				if f.Value != nil && utils.IsArray(f.Value) {
					attachments := make([]emod.FeishuAttachment, 0)
					filePaths := f.Value.([]interface{})
					if len(filePaths) > 0 {
						fileNames := make([]string, 0)
						if f.Ext != nil {
							fileNameStr, _ := utils.ToString(f.Ext)
							if fileNameStr != "" {
								fileNames = strings.Split(fileNameStr, ",")
							}
						}
						for idx, fp := range filePaths {
							fpStr, _ := utils.ToString(fp)
							att := emod.FeishuAttachment{
								InstanceCode: appRecord.InstanceCode,
								FileUrl:      fpStr,
							}
							if len(fileNames) >= idx+1 {
								att.FileName = fileNames[idx]
							}
							attachments = append(attachments, att)
						}
						if len(attachments) > 0 {
							approvalForm.Attachments = attachments
						}
					}
				}
			case "费用明细":
				val := f.Value
				if utils.IsArray(val) {
					details := make([]emod.FeeDetail, 0)
					for _, vs := range val.([]interface{}) {
						if utils.IsArray(vs) {
							detail := emod.FeeDetail{InstanceCode: appRecord.InstanceCode, FeeDate: time.Now().Format(time.DateTime)}
							for _, v := range vs.([]interface{}) {
								if utils.IsMap(v) {
									vmap := v.(map[string]interface{})
									na, ok := vmap["name"]
									naStr, _ := utils.ToString(na)
									co, ok1 := vmap["value"]
									ext, ok2 := vmap["ext"]
									var coStr string
									if ok1 {
										coStr, _ = utils.ToString(co)
									}
									if ok && naStr == "费用类型" && ok1 {
										detail.FeeCode = coStr
										option, ok3 := vmap["option"]
										if ok3 && utils.IsMap(option) && len(option.(map[string]interface{})) > 0 {
											optionMap := option.(map[string]interface{})
											if feeType, ok4 := optionMap["key"]; ok4 {
												detail.FeeCode, _ = utils.ToString(feeType)
											}
										}
									}
									if ok && naStr == "报销事由" && ok1 {
										detail.Reason = coStr
									}
									if ok && naStr == "金额" && ok1 {
										detail.FeeAmount = utils.ToFloat64(co)
									}
									if ok2 && ext != nil {
										if utils.IsMap(ext) {
											extMap := ext.(map[string]interface{})
											if cu, ok4 := extMap["currency"]; ok4 {
												detail.Currency, _ = utils.ToString(cu)
												approvalForm.Currency = detail.Currency
											}
										}
									}
								}
							}
							if detail.FeeDate != "" && detail.FeeAmount > 0 {
								details = append(details, detail)
							}
						}
					}
					if len(details) > 0 {
						approvalForm.Details = details
					}
				}
			}
		}
		var details []emod.FeeDetail
		if len(approvalForm.Details) > 0 {
			err = db.Save(&approvalForm).Error
			if err != nil {
				db.Rollback()
				return err
			}
			details = approvalForm.Details
			err = db.Save(&details).Error
			if err != nil {
				db.Rollback()
				return err
			}
			atts := approvalForm.Attachments
			if len(atts) > 0 {
				err = db.Save(&atts).Error
				if err != nil {
					db.Rollback()
					return err
				}
			}
			appRecord.Form = approvalForm
		}
		db.Commit()
		// 根据费用明细 平台，费用类型 查找当前生效的预算数据
		ym := time.Now().Format("2006-01")
		if len(details) > 0 {
			var user emod.CurrentUser
			e.Orm.Model(&user).Preload("MainDept").Where("open_id = ?", appRecord.OpenId).First(&user)
			logs := make([]emod.FeeRequestLog, 0)
			feishuSvc := esvc.FeishuService{}
			feishuSvc.Orm = e.Orm
			var department emod.CurrentDept
			for _, d := range details {
				var budgetData emod.FeeRequestLog
				err = e.Orm.Table("budget_fee_category_details bd").
					Joins("INNER JOIN budget_fee_category bc ON bc.id=bd.budget_fee_category_id and bc.view_type=1 and  bc.deleted_at is null").
					Joins("INNER JOIN cost_budget_version_detail cd ON cd.budget_fee_category_id = bc.id and cd.deleted_at is null").
					Joins("INNER JOIN cost_budget_version bv ON cd.cost_budget_version_id = bv.id AND bv.status=2 AND effective_date <= CURDATE() AND bv.deleted_at  is null").
					Joins("INNER JOIN cost_center_info ci ON ci.id = bv.cost_center_info_id").
					Joins("INNER JOIN cost_center_related_customer cr ON cr.cost_center_info_id = ci.id").
					Joins("INNER JOIN kingdee_customer_group kcg ON kcg.group_id =cr.group_id").
					Select("bv.id as budget_version_id,cd.id as budget_detail_id, bd.fee_code,bd.fee_name,bv.cost_center_info_id,"+
						"ci.cost_center_name,ci.dept_id,cd.budget_amount,cd.budget_used, cd.years_month as budget_years_month,kcg.group_name,kcg.group_number").
					Where("bd.fee_code = ? AND cd.years_month = ? AND kcg.group_number = ?", d.FeeCode, ym, appRecord.Form.Platform).First(&budgetData).Error
				if err == nil && budgetData.CostCenterName != "" {
					budgetData.RequestAmount = d.FeeAmount
					budgetData.CreateBy = user.UserId
					budgetData.InstanceCode = appRecord.InstanceCode
					budgetData.UserDeptId = user.DeptId
					budgetData.Currency = approvalForm.Currency
					budgetData.OrgCode = approvalForm.OrgCode
					budgetData.ReqUserOpenid = appRecord.OpenId
					logs = append(logs, budgetData)
					if department.DeptId == 0 {
						e.Orm.Model(&department).Preload("LeaderUser").First(&department, "dept_id=?", budgetData.DepartmentId)
					}
				} else {
					return errors.New("未找到完整的成本中心预算信息")
				}
			}
			if len(logs) > 0 {
				e.Orm.Create(&logs)
				userName := fmt.Sprintf("%s(%s)", user.CnName, user.MainDept.DeptName)
				for _, l := range logs {
					// 如果 已使用预算 + 当前申请费用 >= 预算的80% 则给部门负责人报警
					if l.BudgetUsed+l.RequestAmount >= l.BudgetAmount*0.8 {
						params := map[string]string{
							"amount":         fmt.Sprintf("%s %v", l.Currency, l.RequestAmount),
							"deptartment":    department.DeptName,
							"budget_amount":  fmt.Sprintf("%s %v", l.Currency, l.BudgetAmount),
							"budget_balance": fmt.Sprintf("%s %v", l.Currency, l.BudgetAmount-l.BudgetUsed),
							"real_name":      userName,
							"date_str":       appRecord.StartDate,
							"instance_code":  appRecord.InstanceCode,
						}
						err = feishuSvc.SendMessage(user.MainDept.LeaderUser.OpenId, "budgetInsufficient ", params)
					}
				}
			}

		}

	}
	return nil
}
