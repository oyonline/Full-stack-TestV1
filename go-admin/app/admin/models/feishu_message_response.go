package models

import larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

type FeishuMessageTemplate struct {
	ID           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	TemplateCode string `json:"templateCode" gorm:"column:template_code;type:varchar(20)" comment:"模板编码"`
	TemplateName string `json:"templateName" gorm:"column:template_name;type:varchar(20)" comment:"模板名称"`
	MsgType      string `json:"msgType" gorm:"column:msg_type;type:varchar(20)" comment:"text：文本 post：富文本 image：图片 file：文件 audio：语音 media：视频 sticker：表情包 nteractive：卡片 system：系统消息。该类型仅支持在机器人单聊内推送系统消息"`
	Template     string `json:"template" gorm:"column:template;type:longtext" comment:"JSON模板"`
	Params       string `json:"params" gorm:"column:params;type:text"`
}

func (FeishuMessageTemplate) TableName() string {
	return "feishu_message_template"
}

// FeishuMessageResponse 飞书消息发送记录 json是下划线要对应飞书返回的json格式 如果要做消息查询 另做一个结构体不要改这个
type FeishuMessageResponse struct {
	ID             int64              `json:"id" gorm:"primaryKey;autoIncrement"`
	TemplateId     int64              `json:"template_id" gorm:"column:template_id;type:int"`
	MessageId      string             `json:"message_id" gorm:"column:message_id;type:varchar(64)"`             // 消息id open_message_id
	RootId         string             `json:"root_id" gorm:"column:root_id;type:varchar(64)"`                   // 根消息id open_message_id
	ParentId       string             `json:"parent_id" gorm:"column:parent_id;type:varchar(64)"`               // 父消息的id open_message_id
	ThreadId       string             `json:"thread_id" gorm:"column:thread_id;type:varchar(64)"`               // 消息所属的话题 ID
	MsgType        string             `json:"msg_type" gorm:"column:msg_type;type:varchar(64)"`                 // 消息类型 text post card image等等
	CreateTime     string             `json:"create_time" gorm:"column:create_time;type:varchar(30)"`           // 消息生成的时间戳(毫秒)
	UpdateTime     string             `json:"update_time" gorm:"column:update_time;type:varchar(30)"`           // 消息更新的时间戳
	Deleted        bool               `json:"deleted" gorm:"column:deleted;type:tinyint"`                       // 消息是否被撤回
	Updated        bool               `json:"updated" gorm:"column:updated;type:tinyint"`                       // 消息是否被更新
	ChatId         string             `json:"chat_id" gorm:"column:chat_id;type:varchar(64)"`                   // 所属的群
	Body           larkim.MessageBody `json:"body" gorm:"-"`                                                    // 消息内容,json结构
	BodyStr        string             `json:"-" gorm:"column:body;type:longtext"`                               // 消息内容,json结构
	UpperMessageId string             `json:"upper_message_id" gorm:"column:upper_message_id;type:varchar(64)"` // 合并消息的上一层级消息id open_message_id
}

func (FeishuMessageResponse) TableName() string {
	return "feishu_message_response"
}
