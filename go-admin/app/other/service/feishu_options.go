package service

import (
	"fmt"
	"go-admin/app/other/service/dto"
	"strconv"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type FeishuOptions struct {
	service.Service
}

func (e *FeishuOptions) OrgList() dto.FeishuResponseData {
	type orgEnt struct {
		Id     int    `gorm:"column:forgid"`
		Name   string `gorm:"column:fname"`
		EnName string `gorm:"column:f_pxzo_en"`
	}
	var orgs []orgEnt
	e.Orm.Table("kingdee_organize_info").Where("fforbidstatus = 'A' AND fdocumentstatus='C'").Select("forgid,fname,f_pxzo_en").Find(&orgs)
	textMap := make(map[string]string)
	//textEnMap := make(map[string]string)
	options := []dto.Option{}
	for _, o := range orgs {
		key := fmt.Sprintf("@i18n@%d", o.Id)
		textMap[key] = o.Name
		//textEnMap[key] = o.EnName
		option := dto.Option{
			Id:        strconv.Itoa(o.Id),
			Value:     key,
			IsDefault: false,
		}
		options = append(options, option)
	}
	i18ns := []dto.I18nResource{
		{
			Location:  "zh_cn",
			IsDefault: true,
			Texts:     textMap,
		},
		//{
		//	Location:  "en_us",
		//	IsDefault: false,
		//	Texts:     textEnMap,
		//},
	}
	response := dto.FeishuResponseData{}
	response.SetData(i18ns, options)
	return response
}

func (e *FeishuOptions) DepartmentList(orgName, parentName string) dto.FeishuResponseData {
	var fid int64
	if orgName != "" {
		e.Orm.Table("kingdee_organize_info").Where("fforbidstatus = 'A' AND fdocumentstatus='C' AND fname = ?", orgName).Select("forgid").Scan(&fid)
	} else {
		e.Orm.Table("kingdee_dept").Where("dept_status = 'C' AND forbid_status = 'A' AND dept_name=?", parentName).Select("id").Find(&fid)
	}
	response := dto.FeishuResponseData{}
	if fid > 0 {
		type deptEnt struct {
			DeptId     int    `gorm:"column:dept_id"`
			DeptNumber string `gorm:"column:dept_number"`
			DeptName   string `gorm:"column:dept_name"`
		}
		var depts []deptEnt
		tx := e.Orm.Table("kingdee_dept k").Joins("sys_dept d ON k.dept_name = d.dept_name").Where("k.dept_status = 'C' AND k.forbid_status = 'A'")
		if orgName != "" {
			tx.Where("AND k.use_org_id=? and k.parent_id=0", fid)
		} else {
			tx.Where("AND k.parent_id=?", fid)
		}
		tx.Select("d.dept_id,k.dept_number,k.dept_name").Find(&depts)
		textMap := make(map[string]string)
		options := []dto.Option{}
		for _, o := range depts {
			key := fmt.Sprintf("@i18n@%s", o.DeptNumber)
			keyId := fmt.Sprintf("%s:%d", o.DeptNumber, o.DeptId)
			textMap[key] = o.DeptName
			option := dto.Option{
				Id:        keyId,
				Value:     key,
				IsDefault: false,
			}
			options = append(options, option)
		}
		i18ns := []dto.I18nResource{
			{
				Location:  "zh_cn",
				IsDefault: true,
				Texts:     textMap,
			},
		}
		response.SetData(i18ns, options)
	}
	return response
}

func (e *FeishuOptions) PlatformList(departmentName string) dto.FeishuResponseData {
	var deptId int
	response := dto.FeishuResponseData{}
	e.Orm.Table("sys_dept").Where("dept_name=?", departmentName).Select("dept_id").Scan(&deptId)
	if deptId > 0 {
		type platformEnt struct {
			GroupNumber string `gorm:"column:group_number"`
			GroupName   string `gorm:"column:group_name"`
		}
		var platforms []platformEnt
		e.Orm.Table("kingdee_customer_group").Where("group_id in (select group_id from kingdee_customer where dept_id=?)", deptId).Select("group_number,group_name").Find(&platforms)
		if len(platforms) > 0 {
			textMap := make(map[string]string)
			options := []dto.Option{}
			for _, o := range platforms {
				key := fmt.Sprintf("@i18n@%s", o.GroupNumber)
				textMap[key] = o.GroupName
				option := dto.Option{
					Id:        o.GroupNumber,
					Value:     key,
					IsDefault: false,
				}
				options = append(options, option)
			}
			i18ns := []dto.I18nResource{
				{
					Location:  "zh_cn",
					IsDefault: true,
					Texts:     textMap,
				},
			}
			response.SetData(i18ns, options)
		}
	}
	return response
}

func (e *FeishuOptions) FeeCodeList(platformName string) dto.FeishuResponseData {
	type feeEnt struct {
		FeeCode string `gorm:"column:fee_code"`
		FeeName string `gorm:"column:fee_name"`
	}
	response := dto.FeishuResponseData{}
	var feeCodes []feeEnt
	e.Orm.Table("budget_fee_category_details").Where("platform like ?", fmt.Sprintf("/%s/", platformName)).Select("fee_code,fee_name").Find(&feeCodes)
	if len(feeCodes) > 0 {
		textMap := make(map[string]string)
		options := []dto.Option{}
		for _, o := range feeCodes {
			key := fmt.Sprintf("@i18n@%s", o.FeeCode)
			textMap[key] = o.FeeName
			option := dto.Option{
				Id:        o.FeeCode,
				Value:     key,
				IsDefault: false,
			}
			options = append(options, option)
		}
		i18ns := []dto.I18nResource{
			{
				Location:  "zh_cn",
				IsDefault: true,
				Texts:     textMap,
			},
		}
		response.SetData(i18ns, options)
	}
	return response
}
