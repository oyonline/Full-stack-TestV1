package service

import (
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/utils"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"github.com/google/uuid"
	"gorm.io/gorm"

	platformModels "go-admin/app/platform/models"
	"go-admin/app/platform/service/dto"
	"go-admin/common/authctx"
)

const attachmentBasePath = "static/uploadfile/attachment"

type Attachment struct {
	service.Service
}

func (e *Attachment) GetPage(c *dto.AttachmentGetPageReq, list *[]platformModels.AttachmentFile, count *int64) error {
	c.Normalize()
	db := e.Orm.Model(&platformModels.AttachmentFile{})
	if c.ModuleKey != "" {
		db = db.Where("module_key = ?", c.ModuleKey)
	}
	if c.BusinessType != "" {
		db = db.Where("business_type = ?", c.BusinessType)
	}
	if c.BusinessId != "" {
		db = db.Where("business_id = ?", c.BusinessId)
	}

	return db.Order("created_at DESC, attachment_id DESC").
		Offset((c.GetPageIndex() - 1) * c.GetPageSize()).
		Limit(c.GetPageSize()).
		Find(list).Limit(-1).Offset(-1).Count(count).Error
}

func (e *Attachment) Get(id int) (*platformModels.AttachmentFile, error) {
	var item platformModels.AttachmentFile
	if err := e.Orm.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (e *Attachment) Upload(c *gin.Context, req *dto.AttachmentUploadReq, fileHeader *multipart.FileHeader) (*platformModels.AttachmentFile, error) {
	if fileHeader == nil {
		return nil, errors.New("附件不能为空")
	}
	req.Normalize()

	var module platformModels.ModuleRegistry
	if err := e.Orm.Where("module_key = ? AND status = ?", req.ModuleKey, "2").First(&module).Error; err != nil {
		return nil, errors.New("模块未注册或未启用")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	dir := filepath.Join(attachmentBasePath, time.Now().Format("200601"))
	if err := utils.IsNotExistMkDir(dir); err != nil {
		return nil, err
	}
	storageName := uuid.NewString() + ext
	storagePath := filepath.Join(dir, storageName)
	if err := c.SaveUploadedFile(fileHeader, storagePath); err != nil {
		return nil, err
	}

	item := &platformModels.AttachmentFile{
		ModuleKey:    req.ModuleKey,
		BusinessType: req.BusinessType,
		BusinessId:   req.BusinessId,
		BusinessNo:   req.BusinessNo,
		FileName:     fileHeader.Filename,
		FileExt:      ext,
		FileSize:     fileHeader.Size,
		ContentType:  fileHeader.Header.Get("Content-Type"),
		StorageType:  dto.AttachmentStorageLocal,
		StoragePath:  storagePath,
		UploaderId:   user.GetUserId(c),
		UploaderName: user.GetUserName(c),
	}
	item.CreateBy = item.UploaderId
	item.UpdateBy = item.UploaderId

	if err := e.Orm.Create(item).Error; err != nil {
		_ = os.Remove(storagePath)
		return nil, err
	}
	return item, nil
}

func (e *Attachment) Delete(c *gin.Context, id int) error {
	item, err := e.Get(id)
	if err != nil {
		return err
	}

	currentUserID := user.GetUserId(c)
	roleKeys := authctx.GetRoleKeys(c)
	if item.UploaderId != currentUserID && !slices.Contains(roleKeys, "admin") {
		return errors.New("无权删除该附件")
	}

	return e.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&platformModels.AttachmentFile{}, id).Error; err != nil {
			return err
		}
		if item.StoragePath != "" {
			if err := os.Remove(item.StoragePath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	})
}
