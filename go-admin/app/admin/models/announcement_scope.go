package models

type AnnouncementScope struct {
	AnnouncementId int64 `gorm:"primaryKey;column:announcement_id" json:"announcementId"`
	DeptId         int   `gorm:"primaryKey;column:dept_id" json:"deptId"`
}

func (*AnnouncementScope) TableName() string {
	return "announcement_scope"
}
