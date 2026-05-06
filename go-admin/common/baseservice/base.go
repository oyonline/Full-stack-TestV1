// Package baseservice 提供基于泛型的 BaseService[T any]，封装 GetPage/Get/Insert/Update/Remove 五件套。
//
// 业务 service 通过嵌入 BaseService[YourModel] 即可获得默认 CRUD 实现；只需在差异点上重新声明
// 同名方法即可覆盖（Go 嵌入字段方法被外层同名方法 shadow，外部调用 s.GetPage(...) 会走外层）。
//
// 设计原则：
//   - 不引入额外执行语义，五件套与历史 service 的 SQL 行为一一对齐，便于现有 service 平滑迁移。
//   - 不试图覆盖所有边界（多表 join、自定义 where、软删除回滚等），这些场景由业务 service 通过覆盖
//     单个方法 + 直接调用 b.Orm 来处理，不在 BaseService 内增加配置型 API。
//   - 错误信息保留中文 + 原始 db 错误日志，避免影响前端既有提示。
package baseservice

import (
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	cDto "go-admin/common/dto"
)

// PageReq 是 GetPage 接受的请求 DTO 接口（与历史 *PageReq 兼容）。
type PageReq interface {
	GetNeedSearch() interface{}
	GetPageSize() int
	GetPageIndex() int
}

// IDReq 是 Get / Remove 接受的请求 DTO 接口；GetId 可返回 int / int64 / []int 等 GORM 主键查询支持的类型。
type IDReq interface {
	GetId() interface{}
}

// MutateReq[T] 是 Insert / Update 接受的请求 DTO 接口：将自身字段写入目标 model。
type MutateReq[T any] interface {
	Generate(*T)
}

// BaseService 是泛型基类，T 为业务 model 类型。
//
// 使用方式：
//
//	type SysPost struct {
//	    baseservice.BaseService[models.SysPost]
//	}
//
// 嵌入后会自动获得 GetPage/Get/Insert/Update/Remove 五件套；同时 service.Service 也被多层嵌入提升，
// 因此 api 层 MakeService(&s.Service) 仍然可以工作。
type BaseService[T any] struct {
	service.Service
}

// GetPage 通用分页查询。等价于历史 service 中以 cDto.MakeCondition + cDto.Paginate 组成的查询模板。
func (b *BaseService[T]) GetPage(c PageReq, list *[]T, count *int64) error {
	var data T
	err := b.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		b.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Get 通用按主键查询。命中 ErrRecordNotFound 时返回固定中文提示，与历史 service 行为一致。
func (b *BaseService[T]) Get(c IDReq, model *T) error {
	var data T
	db := b.Orm.Model(&data).First(model, c.GetId())
	err := db.Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		b.Log.Errorf("db error: %s", err)
		return err
	}
	if err != nil {
		b.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// Insert 通用新增。把请求 DTO 通过 Generate 写入新建的 T 实例后落库。
//
// 注意：历史 api 通过 req.GetId() 返回新建 ID，但 DTO 在 Insert 前 ID 通常为 0；
// BaseService 不修复该既有行为，迁移时不引入语义变化。如需返回真实 ID，应在调用方读取 model 主键。
func (b *BaseService[T]) Insert(c MutateReq[T]) error {
	var data T
	c.Generate(&data)
	if err := b.Orm.Create(&data).Error; err != nil {
		b.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

// InsertReturn 与 Insert 相同，但把新建后的 model 通过 out 返回，便于调用方读取自增主键 / 默认值。
// 旧 service 不依赖此方法；新模块或需要回填 ID 的场景使用此接口。
func (b *BaseService[T]) InsertReturn(c MutateReq[T], out *T) error {
	var data T
	c.Generate(&data)
	if err := b.Orm.Create(&data).Error; err != nil {
		b.Log.Errorf("db error: %s", err)
		return err
	}
	if out != nil {
		*out = data
	}
	return nil
}

// Update 通用更新。先按主键 First 加载，然后让 DTO Generate 覆盖差异字段，最后 Save 整行。
// RowsAffected==0 时返回"无权更新该数据"，行为与历史 service 一致。
func (b *BaseService[T]) Update(c interface {
	MutateReq[T]
	IDReq
}) error {
	var model T
	b.Orm.First(&model, c.GetId())
	c.Generate(&model)
	db := b.Orm.Save(&model)
	if err := db.Error; err != nil {
		b.Log.Errorf("db error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// Remove 通用删除。GORM 的 Delete 接受单值或切片作为主键集合，所以 GetId() 可以返回 int 或 []int。
// 命中软删除字段（gorm.DeletedAt）时由 GORM 自动改为更新 deleted_at，与硬删除调用方式相同。
func (b *BaseService[T]) Remove(c IDReq) error {
	var data T
	db := b.Orm.Model(&data).Delete(&data, c.GetId())
	if err := db.Error; err != nil {
		b.Log.Errorf("Delete error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}
