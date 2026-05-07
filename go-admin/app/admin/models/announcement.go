package models

import (
	"time"

	"go-admin/common/models"
)

const (
	AnnouncementStatusDraft     = 1
	AnnouncementStatusPublished = 2
	AnnouncementStatusOffline   = 3
)

type Announcement struct {
	AnnouncementId int64      `gorm:"primaryKey;autoIncrement;column:announcement_id" json:"announcementId"`
	Title          string     `gorm:"size:200;not null" json:"title"`
	Content        string     `gorm:"type:mediumtext" json:"content"`
	CoverImageUrl  string     `gorm:"size:512" json:"coverImageUrl"`
	Status         int        `gorm:"size:4;not null;default:1" json:"status"`
	IsTop          int        `gorm:"size:4;not null;default:0" json:"isTop"`
	TopSort        int        `gorm:"size:11;not null;default:0" json:"topSort"`
	PublishAt      *time.Time `gorm:"" json:"publishAt"`
	ExpireAt       *time.Time `gorm:"" json:"expireAt"`
	CreatorId      int        `gorm:"index" json:"creatorId"`
	Remark         string     `gorm:"size:255" json:"remark"`
	models.ControlBy
	models.ModelTime
}

func (*Announcement) TableName() string {
	return "announcement"
}

func (e *Announcement) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *Announcement) GetId() interface{} {
	return e.AnnouncementId
}
