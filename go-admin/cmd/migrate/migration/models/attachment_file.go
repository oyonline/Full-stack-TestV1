package models

import common "go-admin/common/models"

type AttachmentFile struct {
	AttachmentId int    `json:"attachmentId" gorm:"primaryKey;autoIncrement;comment:附件ID"`
	ModuleKey    string `json:"moduleKey" gorm:"size:64;index;comment:模块编码"`
	BusinessType string `json:"businessType" gorm:"size:64;index;comment:业务类型"`
	BusinessId   string `json:"businessId" gorm:"size:64;index;comment:业务ID"`
	BusinessNo   string `json:"businessNo" gorm:"size:128;comment:业务单号"`
	FileName     string `json:"fileName" gorm:"size:255;comment:原始文件名"`
	FileExt      string `json:"fileExt" gorm:"size:32;comment:文件扩展名"`
	FileSize     int64  `json:"fileSize" gorm:"comment:文件大小"`
	ContentType  string `json:"contentType" gorm:"size:128;comment:内容类型"`
	StorageType  string `json:"storageType" gorm:"size:32;default:local;comment:存储类型"`
	StoragePath  string `json:"storagePath" gorm:"size:512;comment:存储路径"`
	UploaderId   int    `json:"uploaderId" gorm:"index;comment:上传人ID"`
	UploaderName string `json:"uploaderName" gorm:"size:128;comment:上传人名称"`
	common.ControlBy
	common.ModelTime
}

func (*AttachmentFile) TableName() string {
	return "att_file"
}
