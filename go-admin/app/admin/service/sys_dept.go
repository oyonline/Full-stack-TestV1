package service

import (
	"errors"
	"fmt"
	"go-admin/app/admin/models"
	"go-admin/common/utils/feishuUtils"

	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"gorm.io/gorm"

	"go-admin/app/admin/service/dto"
	cDto "go-admin/common/dto"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type SysDept struct {
	service.Service
	Su *SysUser
}

// GetDepts 获取SysDept列表
func (e *SysDept) GetDepts(list *[]models.SysDept) error {
	var err error
	var data models.SysDept

	err = e.Orm.Model(&data).Find(list).Error
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// GetPage 获取SysDept列表
//func (e *SysDept) GetPage(c *dto.SysDeptGetPageReq, list *[]models.SysDept) error {
//	var err error
//	var data models.SysDept
//
//	err = e.Orm.Model(&data).
//		Scopes(
//			cDto.MakeCondition(c.GetNeedSearch()),
//		).
//		Find(list).Error
//	if err != nil {
//		e.Log.Errorf("db error:%s", err)
//		return err
//	}
//	return nil
//}

// Get 获取SysDept对象
func (e *SysDept) Get(d *dto.SysDeptGetReq, model *models.SysDept) error {
	var err error
	var data models.SysDept

	err = e.Orm.Model(&data).
		FirstOrInit(model, d.GetId()).
		Error
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		_ = e.AddError(err)
		return err
	}
	if model.DeptId == 0 {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("Service GetSysApi error: %s", err)
		_ = e.AddError(err)
		return err
	}
	return nil
}

// Insert 创建SysDept对象
func (e *SysDept) Insert(c *dto.SysDeptInsertReq) error {
	var err error
	var data models.SysDept
	c.Generate(&data)
	tx := e.Orm.Debug().Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = tx.Create(&data).Error
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	deptPath := pkg.IntToString(data.DeptId) + "/"
	if data.ParentId != 0 {
		var deptP models.SysDept
		tx.First(&deptP, data.ParentId)
		deptPath = deptP.DeptPath + deptPath
	} else {
		deptPath = "/0/" + deptPath
	}
	var mp = map[string]string{}
	mp["dept_path"] = deptPath
	if err = tx.Model(&data).Update("dept_path", deptPath).Error; err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	err = buildPathLevel(&data, tx)
	err = tx.Model(&data).Updates(map[string]interface{}{"dept_path": data.DeptPath, "dept_path_name": data.DeptPathName, "level": data.Level}).Error
	if err != nil {
		e.Log.Errorf("Updates error:%s", err)
		return err
	}
	return nil
}

// Update 修改SysDept对象
func (e *SysDept) Update(c *dto.SysDeptUpdateReq) error {
	var err error
	var model = models.SysDept{}
	tx := e.Orm.Debug().Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	tx.First(&model, c.GetId())
	c.Generate(&model)

	deptPath := pkg.IntToString(model.DeptId) + "/"
	if model.ParentId != 0 {
		var deptP models.SysDept
		tx.First(&deptP, model.ParentId)
		deptPath = deptP.DeptPath + deptPath
	} else {
		deptPath = "/0/" + deptPath
	}
	model.DeptPath = deptPath
	db := tx.Save(&model)
	if err = db.Error; err != nil {
		e.Log.Errorf("UpdateSysDept error:%s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// Remove 删除SysDept
func (e *SysDept) Remove(d *dto.SysDeptDeleteReq) error {
	var err error
	var data models.SysDept

	db := e.Orm.Model(&data).Delete(&data, d.GetId())
	if err = db.Error; err != nil {
		e.Log.Errorf("Delete error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		err = errors.New("无权删除该数据")
		return err
	}
	return nil
}

// GetSysDeptList 获取组织数据
func (e *SysDept) getList(c *dto.SysDeptGetPageReq, list *[]models.SysDept) error {
	var err error
	var data models.SysDept

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
		).
		Find(list).Error
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// SetDeptTree 设置组织数据
func (e *SysDept) SetDeptTree(c *dto.SysDeptGetPageReq) (m []dto.DeptLabel, err error) {
	var list []models.SysDept
	err = e.getList(c, &list)

	m = make([]dto.DeptLabel, 0)
	for i := 0; i < len(list); i++ {
		if list[i].ParentId != 0 {
			continue
		}
		e := dto.DeptLabel{}
		e.Id = list[i].DeptId
		e.Label = list[i].DeptName
		deptsInfo := deptTreeCall(&list, e)

		m = append(m, deptsInfo)
	}
	return
}

// Call 递归构造组织数据
func deptTreeCall(deptList *[]models.SysDept, dept dto.DeptLabel) dto.DeptLabel {
	list := *deptList
	childrenList := make([]dto.DeptLabel, 0)
	for j := 0; j < len(list); j++ {
		if dept.Id != list[j].ParentId {
			continue
		}
		mi := dto.DeptLabel{Id: list[j].DeptId, Label: list[j].DeptName, Children: []dto.DeptLabel{}}
		ms := deptTreeCall(deptList, mi)
		childrenList = append(childrenList, ms)
	}
	dept.Children = childrenList
	return dept
}

// SetDeptPage 设置dept页面数据
func (e *SysDept) SetDeptPage(c *dto.SysDeptGetPageReq) (m []models.SysDept, err error) {
	var list []models.SysDept
	err = e.getList(c, &list)
	for i := 0; i < len(list); i++ {
		if list[i].ParentId != 0 {
			continue
		}
		info := e.deptPageCall(&list, list[i])
		m = append(m, info)
	}
	return
}

func (e *SysDept) deptPageCall(deptlist *[]models.SysDept, menu models.SysDept) models.SysDept {
	list := *deptlist
	childrenList := make([]models.SysDept, 0)
	for j := 0; j < len(list); j++ {
		if menu.DeptId != list[j].ParentId {
			continue
		}
		mi := models.SysDept{}
		mi.DeptId = list[j].DeptId
		mi.ParentId = list[j].ParentId
		mi.DeptPath = list[j].DeptPath
		mi.DeptName = list[j].DeptName
		mi.Sort = list[j].Sort
		mi.Leader = list[j].Leader
		mi.Phone = list[j].Phone
		mi.Email = list[j].Email
		mi.Status = list[j].Status
		mi.CreatedAt = list[j].CreatedAt
		mi.Children = []models.SysDept{}
		ms := e.deptPageCall(deptlist, mi)
		childrenList = append(childrenList, ms)
	}
	menu.Children = childrenList
	return menu
}

// GetWithRoleId 获取角色的部门ID集合
func (e *SysDept) GetWithRoleId(roleId int) ([]int, error) {
	deptIds := make([]int, 0)
	deptList := make([]dto.DeptIdList, 0)
	if err := e.Orm.Table("sys_role_dept").
		Select("sys_role_dept.dept_id").
		Joins("LEFT JOIN sys_dept on sys_dept.dept_id=sys_role_dept.dept_id").
		Where("role_id = ? ", roleId).
		Where(" sys_role_dept.dept_id not in(select sys_dept.parent_id from sys_role_dept LEFT JOIN sys_dept on sys_dept.dept_id=sys_role_dept.dept_id where role_id =? )", roleId).
		Find(&deptList).Error; err != nil {
		return nil, err
	}
	for i := 0; i < len(deptList); i++ {
		deptIds = append(deptIds, deptList[i].DeptId)
	}
	return deptIds, nil
}

func (e *SysDept) SetDeptLabel() (m []dto.DeptLabel, err error) {
	list := make([]models.SysDept, 0)
	err = e.Orm.Find(&list).Error
	if err != nil {
		log.Error("find dept list error, %s", err.Error())
		return
	}
	m = make([]dto.DeptLabel, 0)
	var item dto.DeptLabel
	for i := range list {
		if list[i].ParentId != 0 {
			continue
		}
		item = dto.DeptLabel{}
		item.Id = list[i].DeptId
		item.Label = list[i].DeptName
		deptInfo := deptLabelCall(&list, item)
		m = append(m, deptInfo)
	}
	return
}

// deptLabelCall
func deptLabelCall(deptList *[]models.SysDept, dept dto.DeptLabel) dto.DeptLabel {
	list := *deptList
	var mi dto.DeptLabel
	childrenList := make([]dto.DeptLabel, 0)
	for j := 0; j < len(list); j++ {
		if dept.Id != list[j].ParentId {
			continue
		}
		mi = dto.DeptLabel{Id: list[j].DeptId, Label: list[j].DeptName, Children: []dto.DeptLabel{}}
		ms := deptLabelCall(deptList, mi)
		childrenList = append(childrenList, ms)
	}
	dept.Children = childrenList
	return dept
}

// PullDepartmentChildrens 拉取全部通讯录部门信息
func (e *SysDept) PullDepartmentChildrens(d *dto.DepartmentBatch) error {
	var err error
	tx := e.Orm.Debug().Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	client, err := feishuUtils.NewFeishuClient()
	if err != nil {
		return err
	}

	e.Su.Client = client
	err = e.DepartmentChildrenCall(d.OpenDepartmentIds[0], &d.ParentId, &d.CreateBy, tx)
	if err != nil {
		return err
	}

	// 根据用户修改部门leader信息
	// UPDATE sys_dept AS a, sys_user AS b
	// SET a.leader_uid = b.user_id, a.leader = b.cn_name, a.phone = b.phone, a.email = b.email
	// WHERE a.leader_user_id != '' AND a.leader_user_id = b.open_id
	rawSQL := "UPDATE sys_dept AS a, sys_user AS b SET a.leader_uid = b.user_id, a.leader = b.cn_name, a.phone = b.phone, a.email = b.email WHERE a.leader_user_id != '' AND a.leader_user_id = b.open_id"
	err = tx.Table("sys_dept").Exec(rawSQL).Error
	if err != nil {
		e.Log.Errorf("根据用户修改部门leader信息错误:%s", err)
		return err
	}

	return nil
}

// DepartmentChildrenCall 递归拉取子部门信息
func (e *SysDept) DepartmentChildrenCall(parentDepartmentId string, parentId *int, createBy *int, tx *gorm.DB) error {
	var err error

	// 拉取通讯录根部门信息
	items, err := e.Su.Client.GetDepartmentChildrens(parentDepartmentId)
	if err != nil {
		return err
	}
	for _, v := range items {
		var data = models.SysDept{}
		err = tx.Find(&data, "open_department_id = ?", v.OpenDepartmentId).Error
		if err != nil {
			return err
		}
		data.FeishuGenerate(v)
		data.ParentId = *parentId
		data.UpdateBy = *createBy
		if data.DeptId == 0 {
			data.CreateBy = *createBy
			err = tx.Create(&data).Error
			if err != nil {
				return err
			}
			err = buildPathLevel(&data, tx)
			err = tx.Model(&data).Updates(map[string]interface{}{"dept_path": data.DeptPath, "dept_path_name": data.DeptPathName, "level": data.Level}).Error
		} else {
			err = buildPathLevel(&data, tx)
			err = tx.Omit("dept_code", "user_number", "dept_type").Save(&data).Error
		}
		if err != nil {
			return err
		}
		// 拉取部门用户信息
		err = e.Su.PullDepartmentUsers(*v.OpenDepartmentId, &data.DeptId, createBy)
		if err != nil {
			return err
		}

		err = e.DepartmentChildrenCall(*v.OpenDepartmentId, &data.DeptId, createBy, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

// buildPathLevel
func buildPathLevel(data *models.SysDept, tx *gorm.DB) error {
	var err error
	deptPath := pkg.IntToString(data.DeptId) + "/"
	deptPathName := " > " + data.DeptName
	if data.ParentId != 0 {
		var deptP models.SysDept
		err = tx.First(&deptP, data.ParentId).Error
		deptPath = deptP.DeptPath + deptPath
		deptPathName = deptP.DeptPathName + deptPathName
		data.Level = deptP.Level + 1
	} else {
		deptPath = "/0/" + deptPath
		deptPathName = "深圳波赛冬网络科技有限公司" + deptPathName
		data.Level = 1
	}
	data.DeptPath = deptPath
	data.DeptPathName = deptPathName
	return err
}

// PullDepartmentBatchs 根据部门ID拉取通讯录部门信息
func (e *SysDept) PullDepartmentBatchs(d *dto.DepartmentBatch) error {
	var err error
	tx := e.Orm.Debug().Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	client, err := feishuUtils.NewFeishuClient()
	if err != nil {
		return err
	}

	// 根据部门ID拉取通讯录部门信息
	e.Su.Client = client
	items, err := client.GetDepartmentBatchs(d.OpenDepartmentIds)
	if err != nil {
		return err
	}
	for _, v := range items {
		var data models.SysDept
		e.Orm.Delete(&models.SysUserDept{}, "open_department_id = ?", v.OpenDepartmentId)
		err = e.Orm.Find(&data, "open_department_id = ?", v.OpenDepartmentId).Error
		if err != nil {
			e.Log.Errorf("获取部门信息 error:%s", err)
			return err
		}
		data.FeishuGenerate(v)
		data.ParentId = d.ParentId
		data.UpdateBy = d.CreateBy
		if data.DeptId == 0 {
			data.CreateBy = d.CreateBy
			err = tx.Create(&data).Error
			if err != nil {
				e.Log.Errorf("新增部门信息 error:%s", err)
				return err
			}
			err = buildPathLevel(&data, tx)
			err = tx.Model(&data).Updates(map[string]interface{}{"dept_path": data.DeptPath, "dept_path_name": data.DeptPathName, "level": data.Level}).Error
		} else {
			err = buildPathLevel(&data, tx)
			err = tx.Omit("dept_code", "user_number", "dept_type").Save(&data).Error
		}
		if err != nil {
			e.Log.Errorf("修改部门信息 error:%s", err)
			return err
		}
		// 拉取部门用户信息
		err = e.Su.PullDepartmentUsers(*v.OpenDepartmentId, &data.DeptId, &d.CreateBy)
		if err != nil {
			return err
		}

		// 根据修改部门leader信息
		if *v.LeaderUserId != "" {
			var user models.SysUser
			e.Su.Orm.Find(&user, "open_id = ?", v.LeaderUserId)
			err = tx.Model(&data).
				Updates(map[string]interface{}{
					"leader_uid": user.UserId,
					"leader":     user.CnName,
					"phone":      user.Phone,
					"email":      user.Email,
				}).Error
			if err != nil {
				e.Log.Errorf("修改部门leader信息 error:%s", err)
				return err
			}
		}
	}
	return nil
}

// PullDepartments 同步飞书部门Department
func (e *SysDept) PullDepartments() error {
	var err error
	client, err := feishuUtils.NewFeishuClient()
	if err != nil {
		return err
	}

	// 获取飞书token
	list, err := client.GetDepartmentList()
	if err != nil {
		return err
	}
	fmt.Println(list)
	return nil
}
