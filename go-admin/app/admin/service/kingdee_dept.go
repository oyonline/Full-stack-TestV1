package service

import (
	"encoding/json"
	"go-admin/app/admin/models"
	"go-admin/common/utils/kingdeeUtils"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type KingdeeDept struct {
	service.Service
}

// PullKingdeeDepts 拉取KingdeeDept对象
func (e *KingdeeDept) PullKingdeeDepts() error {
	var err error
	err = e.Orm.Table("kingdee_dept").Exec("DELETE FROM kingdee_dept").Error
	if err != nil {
		e.Log.Errorf("Delete error:%s", err)
		return err
	}

	// 拉取金蝶部门信息
	postData := map[string]string{
		"FormId":       "BD_Department",
		"FieldKeys":    "FDeptId,FParentId,FName,FNumber,FDocumentStatus,FForbidStatus,FUseOrgId",
		"FilterString": "",
		"StartRow":     "0",
		"Limit":        "0",
	}
	respStr := kingdeeUtils.BillQuery(postData)
	list := make([]models.KingdeeDept, 0)
	err = json.Unmarshal([]byte(respStr), &list)
	if err != nil {
		e.Log.Errorf("JsonUnmarshal error:%s", err)
		return err
	}

	err = e.Orm.CreateInBatches(&list, len(list)).Error
	if err != nil {
		e.Log.Errorf("CreateInBatches error:%s", err)
		return err
	}
	return nil
}
