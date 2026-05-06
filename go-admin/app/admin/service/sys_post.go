package service

import (
	"go-admin/app/admin/models"
	"go-admin/common/baseservice"
)

// SysPost 试点：使用 BaseService[models.SysPost] 提供默认 CRUD。
//
// 五件套（GetPage/Get/Insert/Update/Remove）由 BaseService 提供。如需差异行为，
// 在本类型上声明同名方法即可覆盖（外层方法优先于嵌入字段方法）。
type SysPost struct {
	baseservice.BaseService[models.SysPost]
}
