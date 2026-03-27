package service

import (
	"encoding/json"
	"go-admin/app/admin/models"
	"go-admin/common/utils/kingdeeUtils"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type KingdeeOrganizeInfo struct {
	service.Service
}

// PullKingdeeOrganizeInfos 拉取KingdeeOrganizeInfo对象
func (e *KingdeeOrganizeInfo) PullKingdeeOrganizeInfos() error {
	var err error
	err = e.Orm.Table("kingdee_organize_info").Exec("DELETE FROM kingdee_organize_info").Error
	if err != nil {
		e.Log.Errorf("Delete error:%s", err)
		return err
	}

	// 拉取金蝶组织信息
	postData := map[string]string{
		"FormId":       "ORG_Organizations",
		"FieldKeys":    "FOrgID,FDocumentStatus,FForbidStatus,FName,FNumber,FDescription,FCreateDate,FModifyDate,FContact,FOrgFormID,FAddress,FTel,FAcctOrgType,FParentID,FIsBusinessOrg,FIsAccountOrg,F_PXZO_SCC,F_PXZO_EN,F_PXZO_ED",
		"FilterString": "",
		"StartRow":     "0",
		"Limit":        "0",
	}
	respStr := kingdeeUtils.BillQuery(postData)
	list := make([]models.KingdeeOrganizeInfo, 0)
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
