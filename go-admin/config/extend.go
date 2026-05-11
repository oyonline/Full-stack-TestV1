package config

import "time"

var ExtConfig Extend

// Extend 扩展配置
//
//	extend:
//	  demo:
//	    name: demo-name
//
// 使用方法： config.ExtConfig......即可！！

type Extend struct {
	Lingxing     Lingxing
	Feishu       Feishu
	Kingdee      Kingdee
	Announcement AnnouncementGCConfig
}

// AnnouncementGCConfig 公告附件 GC 配置
type AnnouncementGCConfig struct {
	AttachmentGC AttachmentGC `yaml:"attachment_gc"`
}

// AttachmentGC 附件垃圾回收开关
type AttachmentGC struct {
	DryRun bool `yaml:"dry_run"`
}

type Lingxing struct {
	Host   string
	Appid  string
	Secret string
	Scheme string
}

type Feishu struct {
	Appid        string
	Appsecret    string
	Timeout      time.Duration
	Token        string
	Approvalcode []string
}

type Kingdee struct {
	Hosturl  string
	Acctid   string
	Username string
	Password string
	Lcid     string
}
