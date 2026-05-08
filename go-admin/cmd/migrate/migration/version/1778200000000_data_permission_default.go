package version

import (
	"runtime"

	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1778200000000DataPermissionDefault)
}

// _1778200000000DataPermissionDefault 兜底 sys_role.data_scope，为 phase2 打开
// settings.application.enabledp 做前置数据清洗。
//
// 背景（C7-1，bead my-tpz）：
//
//	UI '数据范围' 下拉的目标语义是 1=全部 / 2=本部门 / 3=本部门及下级 / 4=仅本人 / 5=自定义。
//	历史 db.sql 把 admin (role_id=1) 的 data_scope 写成 ''（空字符串），其它在线
//	环境也存在 NULL/'' 的存量行。后续 enabledp=true 一旦打开，service 会按
//	data_scope 拼 WHERE，空值会被解读为 '不属于任何已知范围' → 直接掐断查询。
//
// 因此在打开开关之前，先把所有 NULL 或空 data_scope 的 sys_role 兜底为 '1'（全部）：
//   - admin（role_id=1）显式设为 '1'，与 RBAC '系统管理员看到全部数据' 的语义一致；
//   - 其它历史角色保持原状（已经填了 '2' / '3' / '4' / '5' 的不动），仅修空值；
//   - 不区分 NULL 与 ''，UPDATE WHERE data_scope IS NULL OR data_scope = '' 一并覆盖。
//
// 重跑安全：第二次运行命中 0 行。sys_migration 的版本号写入由
// migration.Migrate 框架在 SetVersion + 上层调度处保证幂等。
func _1778200000000DataPermissionDefault(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"UPDATE sys_role SET data_scope = ? WHERE data_scope IS NULL OR data_scope = ''",
			"1",
		).Error; err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
