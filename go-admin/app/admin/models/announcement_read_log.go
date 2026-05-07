package models

import "time"

type AnnouncementReadLog struct {
	UserId         int       `gorm:"primaryKey;column:user_id" json:"userId"`
	AnnouncementId int64     `gorm:"primaryKey;column:announcement_id" json:"announcementId"`
	ReadAt         time.Time `gorm:"column:read_at" json:"readAt"`
}

func (*AnnouncementReadLog) TableName() string {
	return "announcement_read_log"
}
