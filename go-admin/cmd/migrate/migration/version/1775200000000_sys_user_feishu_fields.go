package version

import (
	"runtime"

	adminModels "go-admin/app/admin/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1775200000000SysUserFeishuFields)
}

// _1775200000000SysUserFeishuFields 把飞书相关 5 个字段加到 sys_user 表。
//
// 运行时 model（app/admin/models/sys_user.go）已经声明这 5 列，但历史 migration
// 缺失，导致 fresh DB 创建用户时报 Unknown column 'open_id'。
//
// 加列：
//   - open_id              varchar(55)：飞书用户应用ID
//   - job_title            varchar(55)：飞书用户职务
//   - open_department_id   varchar(55)：飞书系统部门ID
//   - open_department_ids  varchar(255)：飞书系统多部门ID
//   - cn_name              varchar(25)：飞书中文名
func _1775200000000SysUserFeishuFields(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		columns := []struct {
			column string
			field  string
		}{
			{"open_id", "OpenId"},
			{"job_title", "JobTitle"},
			{"open_department_id", "OpenDepartmentId"},
			{"open_department_ids", "OpenDepartmentIds"},
			{"cn_name", "CnName"},
		}
		for _, c := range columns {
			if !tx.Migrator().HasColumn(&adminModels.SysUser{}, c.column) {
				if err := tx.Migrator().AddColumn(&adminModels.SysUser{}, c.field); err != nil {
					return err
				}
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
