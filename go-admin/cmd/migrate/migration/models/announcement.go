package models

import "time"

type Announcement struct {
	AnnouncementId int64      `gorm:"primaryKey;autoIncrement;column:announcement_id" json:"announcementId"`
	Title          string     `gorm:"size:200;not null" json:"title"`
	Content        string     `gorm:"type:mediumtext" json:"content"`
	CoverImageUrl  string     `gorm:"size:512" json:"coverImageUrl"`
	Status         int        `gorm:"size:4;not null;default:1" json:"status"`
	IsTop          int        `gorm:"size:4;not null;default:0" json:"isTop"`
	TopSort        int        `gorm:"size:11;not null;default:0" json:"topSort"`
	PublishAt      *time.Time `json:"publishAt"`
	ExpireAt       *time.Time `json:"expireAt"`
	CreatorId      int        `gorm:"index" json:"creatorId"`
	Remark         string     `gorm:"size:255" json:"remark"`
	ControlBy
	ModelTime
}

func (Announcement) TableName() string {
	return "announcement"
}

type AnnouncementScope struct {
	AnnouncementId int64 `gorm:"primaryKey;column:announcement_id"`
	DeptId         int   `gorm:"primaryKey;column:dept_id"`
}

func (AnnouncementScope) TableName() string {
	return "announcement_scope"
}

type AnnouncementReadLog struct {
	UserId         int       `gorm:"primaryKey;column:user_id"`
	AnnouncementId int64     `gorm:"primaryKey;column:announcement_id"`
	ReadAt         time.Time `gorm:"column:read_at"`
}

func (AnnouncementReadLog) TableName() string {
	return "announcement_read_log"
}
