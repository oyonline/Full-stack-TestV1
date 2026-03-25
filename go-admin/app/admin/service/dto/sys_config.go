package dto

import (
	"go-admin/app/admin/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

// SysConfigGetPageReq 列表或者搜索使用结构体
type SysConfigGetPageReq struct {
	dto.Pagination `search:"-"`
	ConfigName     string `form:"configName" search:"type:contains;column:config_name;table:sys_config"`
	ConfigKey      string `form:"configKey" search:"type:contains;column:config_key;table:sys_config"`
	ConfigType     string `form:"configType" search:"type:exact;column:config_type;table:sys_config"`
	IsFrontend     string `form:"isFrontend" search:"type:exact;column:is_frontend;table:sys_config"`
	SysConfigOrder
}

type SysConfigOrder struct {
	IdOrder         string `search:"type:order;column:id;table:sys_config" form:"idOrder"`
	ConfigNameOrder string `search:"type:order;column:config_name;table:sys_config" form:"configNameOrder"`
	ConfigKeyOrder  string `search:"type:order;column:config_key;table:sys_config" form:"configKeyOrder"`
	ConfigTypeOrder string `search:"type:order;column:config_type;table:sys_config" form:"configTypeOrder"`
	CreatedAtOrder  string `search:"type:order;column:created_at;table:sys_config" form:"createdAtOrder"`
}

func (m *SysConfigGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type SysConfigGetToSysAppReq struct {
	IsFrontend string `form:"isFrontend" search:"type:exact;column:is_frontend;table:sys_config"`
}

func (m *SysConfigGetToSysAppReq) GetNeedSearch() interface{} {
	return *m
}

// SysConfigControl 增、改使用的结构体
type SysConfigControl struct {
	Id          int    `uri:"Id" comment:"编码"` // 编码
	ConfigName  string `json:"configName" comment:""`
	ConfigKey   string `uri:"configKey" json:"configKey" comment:""`
	ConfigValue string `json:"configValue" comment:""`
	ConfigType  string `json:"configType" comment:""`
	IsFrontend  string `json:"isFrontend"`
	Remark      string `json:"remark" comment:""`
	common.ControlBy
}

// Generate 结构体数据转化 从 SysConfigControl 至 system.SysConfig 对应的模型
func (s *SysConfigControl) Generate(model *models.SysConfig) {
	if s.Id == 0 {
		model.Model = common.Model{Id: s.Id}
	}
	model.ConfigName = s.ConfigName
	model.ConfigKey = s.ConfigKey
	model.ConfigValue = s.ConfigValue
	model.ConfigType = s.ConfigType
	model.IsFrontend = s.IsFrontend
	model.Remark = s.Remark

}

// GetId 获取数据对应的ID
func (s *SysConfigControl) GetId() interface{} {
	return s.Id
}

// GetSetSysConfigReq 增、改使用的结构体
type GetSetSysConfigReq struct {
	ConfigKey   string `json:"configKey" comment:""`
	ConfigValue string `json:"configValue" comment:""`
}

// Generate 结构体数据转化 从 SysConfigControl 至 system.SysConfig 对应的模型
func (s *GetSetSysConfigReq) Generate(model *models.SysConfig) {
	model.ConfigValue = s.ConfigValue
}

type UpdateSetSysConfigReq map[string]string

// SysConfigByKeyReq 根据Key获取配置
type SysConfigByKeyReq struct {
	ConfigKey string `uri:"configKey" search:"type:contains;column:config_key;table:sys_config"`
}

func (m *SysConfigByKeyReq) GetNeedSearch() interface{} {
	return *m
}

type GetSysConfigByKEYForServiceResp struct {
	ConfigKey   string `json:"configKey" comment:""`
	ConfigValue string `json:"configValue" comment:""`
}

type SysConfigGetReq struct {
	Id int `uri:"id"`
}

func (s *SysConfigGetReq) GetId() interface{} {
	return s.Id
}

type SysConfigDeleteReq struct {
	Ids []int `json:"ids"`
	common.ControlBy
}

func (s *SysConfigDeleteReq) GetId() interface{} {
	return s.Ids
}

type SystemBrandingSettings struct {
	AppLogo string `json:"appLogo"`
	AppName string `json:"appName"`
}

type SystemUiAppSettings struct {
	ColorGrayMode      bool   `json:"colorGrayMode"`
	ColorWeakMode      bool   `json:"colorWeakMode"`
	ContentCompact     string `json:"contentCompact"`
	DynamicTitle       bool   `json:"dynamicTitle"`
	EnableCheckUpdates bool   `json:"enableCheckUpdates"`
	Layout             string `json:"layout"`
	Locale             string `json:"locale"`
	Watermark          bool   `json:"watermark"`
	WatermarkContent   string `json:"watermarkContent"`
}

type SystemUiBreadcrumbSettings struct {
	Enable      bool   `json:"enable"`
	HideOnlyOne bool   `json:"hideOnlyOne"`
	ShowHome    bool   `json:"showHome"`
	ShowIcon    bool   `json:"showIcon"`
	StyleType   string `json:"styleType"`
}

type SystemUiCopyrightSettings struct {
	CompanyName     string `json:"companyName"`
	CompanySiteLink string `json:"companySiteLink"`
	Date            string `json:"date"`
	Enable          bool   `json:"enable"`
	Icp             string `json:"icp"`
	IcpLink         string `json:"icpLink"`
}

type SystemUiFooterSettings struct {
	Enable bool `json:"enable"`
	Fixed  bool `json:"fixed"`
}

type SystemUiHeaderSettings struct {
	Enable    bool   `json:"enable"`
	Hidden    bool   `json:"hidden"`
	MenuAlign string `json:"menuAlign"`
	Mode      string `json:"mode"`
}

type SystemUiNavigationSettings struct {
	Accordion bool   `json:"accordion"`
	Split     bool   `json:"split"`
	StyleType string `json:"styleType"`
}

type SystemUiShortcutKeySettings struct {
	Enable           bool `json:"enable"`
	GlobalLockScreen bool `json:"globalLockScreen"`
	GlobalLogout     bool `json:"globalLogout"`
	GlobalSearch     bool `json:"globalSearch"`
}

type SystemUiSidebarSettings struct {
	AutoActivateChild  bool `json:"autoActivateChild"`
	Collapsed          bool `json:"collapsed"`
	CollapsedButton    bool `json:"collapsedButton"`
	CollapsedShowTitle bool `json:"collapsedShowTitle"`
	Draggable          bool `json:"draggable"`
	Enable             bool `json:"enable"`
	ExpandOnHover      bool `json:"expandOnHover"`
	FixedButton        bool `json:"fixedButton"`
	Width              int  `json:"width"`
}

type SystemUiTabbarSettings struct {
	Draggable          bool   `json:"draggable"`
	Enable             bool   `json:"enable"`
	MaxCount           int    `json:"maxCount"`
	MiddleClickToClose bool   `json:"middleClickToClose"`
	Persist            bool   `json:"persist"`
	ShowIcon           bool   `json:"showIcon"`
	ShowMaximize       bool   `json:"showMaximize"`
	ShowMore           bool   `json:"showMore"`
	StyleType          string `json:"styleType"`
	VisitHistory       bool   `json:"visitHistory"`
	Wheelable          bool   `json:"wheelable"`
}

type SystemUiThemeSettings struct {
	BuiltinType        string `json:"builtinType"`
	ColorPrimary       string `json:"colorPrimary"`
	FontSize           int    `json:"fontSize"`
	Mode               string `json:"mode"`
	Radius             string `json:"radius"`
	SemiDarkHeader     bool   `json:"semiDarkHeader"`
	SemiDarkSidebar    bool   `json:"semiDarkSidebar"`
	SemiDarkSidebarSub bool   `json:"semiDarkSidebarSub"`
}

type SystemUiTransitionSettings struct {
	Enable   bool   `json:"enable"`
	Loading  bool   `json:"loading"`
	Name     string `json:"name"`
	Progress bool   `json:"progress"`
}

type SystemUiWidgetSettings struct {
	Fullscreen     bool `json:"fullscreen"`
	GlobalSearch   bool `json:"globalSearch"`
	LanguageToggle bool `json:"languageToggle"`
	LockScreen     bool `json:"lockScreen"`
	Notification   bool `json:"notification"`
	Refresh        bool `json:"refresh"`
	SidebarToggle  bool `json:"sidebarToggle"`
	ThemeToggle    bool `json:"themeToggle"`
	Timezone       bool `json:"timezone"`
}

type SystemUiPreferences struct {
	App          SystemUiAppSettings         `json:"app"`
	Breadcrumb   SystemUiBreadcrumbSettings  `json:"breadcrumb"`
	Copyright    SystemUiCopyrightSettings   `json:"copyright"`
	Footer       SystemUiFooterSettings      `json:"footer"`
	Header       SystemUiHeaderSettings      `json:"header"`
	Navigation   SystemUiNavigationSettings  `json:"navigation"`
	ShortcutKeys SystemUiShortcutKeySettings `json:"shortcutKeys"`
	Sidebar      SystemUiSidebarSettings     `json:"sidebar"`
	Tabbar       SystemUiTabbarSettings      `json:"tabbar"`
	Theme        SystemUiThemeSettings       `json:"theme"`
	Transition   SystemUiTransitionSettings  `json:"transition"`
	Widget       SystemUiWidgetSettings      `json:"widget"`
}

type SystemSettingsPayload struct {
	Branding      SystemBrandingSettings `json:"branding"`
	UIPreferences SystemUiPreferences    `json:"uiPreferences"`
}
