package apis

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"

	"go-admin/app/platform/models"
	"go-admin/app/platform/service"
	"go-admin/app/platform/service/dto"
	"go-admin/common/middleware"
)

type Attachment struct {
	api.Api
}

func (e Attachment) GetPage(c *gin.Context) {
	s := service.Attachment{}
	req := dto.AttachmentGetPageReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.Form).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}

	list := make([]models.AttachmentFile, 0)
	var count int64
	if err = s.GetPage(&req, &list, &count); err != nil {
		e.Error(500, err, "查询失败")
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

func (e Attachment) Upload(c *gin.Context) {
	s := service.Attachment{}
	req := dto.AttachmentUploadReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, binding.FormMultipart).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}

	fileHeader, fileErr := c.FormFile("file")
	if fileErr != nil {
		e.Error(500, fileErr, "附件不能为空")
		return
	}

	data, err := s.Upload(c, &req, fileHeader)
	if err != nil {
		e.Error(500, err, "上传失败,"+err.Error())
		return
	}

	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "平台附件",
		BusinessType:  middleware.AuditActionCreate,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.attachment.upload",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("模块编码", req.ModuleKey),
			middleware.AuditKV("业务类型", req.BusinessType),
			middleware.AuditKV("业务ID", req.BusinessId),
			middleware.AuditKV("文件名", fileHeader.Filename),
		),
	})
	e.OK(data, "上传成功")
}

func (e Attachment) Download(c *gin.Context) {
	s := service.Attachment{}
	req := dto.AttachmentGetReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}

	item, err := s.Get(req.Id)
	if err != nil {
		e.Error(500, err, "附件不存在")
		return
	}
	if _, statErr := os.Stat(item.StoragePath); statErr != nil {
		e.Error(500, statErr, "附件文件不存在")
		return
	}
	c.FileAttachment(item.StoragePath, item.FileName)
}

func (e Attachment) Delete(c *gin.Context) {
	s := service.Attachment{}
	req := dto.AttachmentGetReq{}
	err := e.MakeContext(c).
		MakeOrm().
		Bind(&req, nil).
		MakeService(&s.Service).
		Errors
	if err != nil {
		e.Error(500, err, err.Error())
		return
	}

	item, getErr := s.Get(req.Id)
	if getErr != nil {
		e.Error(500, getErr, "附件不存在")
		return
	}
	if err = s.Delete(c, req.Id); err != nil {
		e.Error(500, err, "删除失败,"+err.Error())
		return
	}

	middleware.SetAuditMeta(c, middleware.AuditMeta{
		Title:         "平台附件",
		BusinessType:  middleware.AuditActionDelete,
		BusinessTypes: middleware.AuditCategoryWorkflow,
		Method:        "platform.attachment.delete",
		OperatorType:  middleware.AuditOperatorManage,
		Remark: middleware.AuditSummary(
			middleware.AuditKV("附件ID", req.Id),
			middleware.AuditKV("模块编码", item.ModuleKey),
			middleware.AuditKV("业务类型", item.BusinessType),
			middleware.AuditKV("业务ID", item.BusinessId),
			middleware.AuditKV("文件名", item.FileName),
		),
	})
	e.OK(nil, "删除成功")
}
