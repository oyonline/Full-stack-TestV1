package feishuUtils

import (
	"context"
	"encoding/json"
	"fmt"
	otherDto "go-admin/app/other/service/dto"
	"go-admin/config"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk"
	config2 "github.com/go-admin-team/go-admin-core/sdk/config"
	"github.com/go-admin-team/go-admin-core/storage"
	"github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
	"github.com/larksuite/oapi-sdk-go/v3/service/auth/v3"
	"github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// FeishuClient represents a client for interacting with the Feishu API.
type FeishuClient struct {
	feishuProperties *config.Feishu
	feishuClient     *lark.Client
	Cache            storage.AdapterCache
	CTX              context.Context
}

func (e *FeishuClient) Set(_ context.Context, key string, value string, expireTime time.Duration) error {
	return e.Cache.Set(key, value, int(expireTime))
}

func (e *FeishuClient) Get(_ context.Context, key string) (string, error) {
	token, err := e.Cache.Get(key)
	if token == "" {
		token, err = e.GetTenantAccessToken()
		err = e.Set(nil, key, token, 3600)
		if err != nil {
			return "", err
		}
	}
	return token, err
}

// AccessTokenRawResp AccessToken for interacting with the Feishu API.
type AccessTokenRawResp struct {
	// 基础响应字段（所有飞书接口通用）
	Code      int64  `json:"code"`       // 响应码，0 表示成功
	Msg       string `json:"msg"`        // 响应信息，"success" 表示成功
	RequestId string `json:"request_id"` // 请求 ID，用于排查问题

	// 核心业务字段（租户令牌相关）
	AppAccessToken    string `json:"app_access_token"`    // 应用级访问令牌
	TenantAccessToken string `json:"tenant_access_token"` // 租户级访问令牌
	Expire            int64  `json:"expire"`              // 过期时间（秒，默认 7200 秒）
}

// NewFeishuClient creates and initializes a new FeishuClient.
func NewFeishuClient() (*FeishuClient, error) {
	props := config.ExtConfig.Feishu
	if props.Appid == "" {
		props.Appid = "cli_a2692d11147bd00d"
	}
	if props.Appsecret == "" {
		props.Appsecret = "efJL696q0ldbF0OJJlYtXfxffSyAqkSO"
	}
	if props.Timeout == 0 {
		props.Timeout = 30
	}
	props.Timeout = 30 * time.Second
	var logLevel larkcore.LogLevel
	if config2.ApplicationConfig.Mode == "dev" {
		logLevel = larkcore.LogLevelError
	} else {
		logLevel = larkcore.LogLevelDebug
	}
	redisClient := sdk.Runtime.GetCacheAdapter()
	client := lark.NewClient(props.Appid, props.Appsecret,
		lark.WithLogLevel(logLevel),
		lark.WithReqTimeout(props.Timeout),
		lark.WithEnableTokenCache(true))
	return &FeishuClient{
		feishuProperties: &props,
		feishuClient:     client,
		Cache:            redisClient,
		CTX:              context.Background(),
	}, nil
}

// GetAppAccessToken retrieves an app access token.
func (c *FeishuClient) GetAppAccessToken() (string, error) {
	resp, err := c.feishuClient.Auth.V3.AppAccessToken.Internal(c.CTX,
		larkauth.NewInternalAppAccessTokenReqBuilder().
			Body(larkauth.NewInternalAppAccessTokenReqBodyBuilder().
				AppId(c.feishuProperties.Appid).
				AppSecret(c.feishuProperties.Appsecret).
				Build()).
			Build())
	if err != nil {
		return "", fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	var appResp AccessTokenRawResp
	err = json.Unmarshal(resp.RawBody, &appResp)
	if err != nil {
		return "token异常", err
	}
	return appResp.AppAccessToken, nil
}

// GetTenantAccessToken retrieves a tenant access token.
func (c *FeishuClient) GetTenantAccessToken() (string, error) {
	resp, err := c.feishuClient.Auth.V3.TenantAccessToken.Internal(c.CTX,
		larkauth.NewInternalTenantAccessTokenReqBuilder().
			Body(larkauth.NewInternalTenantAccessTokenReqBodyBuilder().
				AppId(c.feishuProperties.Appid).
				AppSecret(c.feishuProperties.Appsecret).
				Build()).
			Build())
	if err != nil {
		return "", fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	var tenantResp AccessTokenRawResp
	err = json.Unmarshal(resp.RawBody, &tenantResp)
	if err != nil {
		return "token异常", err
	}
	return tenantResp.TenantAccessToken, nil
}

// GetUserInfoByMobileORMail retrieves user information by mobile or email.
func (c *FeishuClient) GetUserInfoByMobileORMail(args []string) ([]*larkcontact.UserContactInfo, error) {
	req := larkcontact.NewBatchGetIdUserReqBuilder().
		UserIdType(`union_id`).
		Body(larkcontact.NewBatchGetIdUserReqBodyBuilder().
			Mobiles(args).
			IncludeResigned(true).
			Build()).
		Build()

	resp, err := c.feishuClient.Contact.User.BatchGetId(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("无数据")
	}
	return resp.Data.UserList, nil
}

// GetUserByDepartments 根据部门ID获取用户信息
func (c *FeishuClient) GetUserByDepartments(openDepartmentId string) ([]*larkcontact.User, error) {
	req := larkcontact.NewFindByDepartmentUserReqBuilder().DepartmentId(openDepartmentId).PageSize(50).Build()
	resp, err := c.feishuClient.Contact.V3.User.FindByDepartment(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	return resp.Data.Items, nil
}

// GetUserBatchs 根据用户应用ID获取用户信息
func (c *FeishuClient) GetUserBatchs(openIds []string) ([]*larkcontact.User, error) {
	req := larkcontact.NewBatchUserReqBuilder().UserIds(openIds).Build()
	resp, err := c.feishuClient.Contact.V3.User.Batch(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	return resp.Data.Items, nil
}

// GetDepartmentChildrens 获取子部门列表
func (c *FeishuClient) GetDepartmentChildrens(parentDepartmentID string) ([]*larkcontact.Department, error) {
	req := larkcontact.NewChildrenDepartmentReqBuilder().DepartmentId(parentDepartmentID).PageSize(50).Build()
	resp, err := c.feishuClient.Contact.V3.Department.Children(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	return resp.Data.Items, nil
}

// GetDepartmentBatchs 批量获取部门信息
func (c *FeishuClient) GetDepartmentBatchs(openDepartmentIds []string) ([]*larkcontact.Department, error) {
	req := larkcontact.NewBatchDepartmentReqBuilder().DepartmentIds(openDepartmentIds).Build()
	resp, err := c.feishuClient.Contact.V3.Department.Batch(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	return resp.Data.Items, nil
}

func (c *FeishuClient) GetDepartmentList() (*larkcontact.ListDepartmentRespData, error) {
	req := larkcontact.NewListDepartmentReqBuilder().Build()
	resp, err := c.feishuClient.Contact.Department.List(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	return resp.Data, nil
}

// CreateMessage 发送消息
func (c *FeishuClient) CreateMessage(messageInfo map[string]string) (*larkim.CreateMessageRespData, error) {
	req := larkim.NewCreateMessageReqBuilder().ReceiveIdType(messageInfo["receive_id_type"]).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(messageInfo["receive_id"]).
			MsgType(messageInfo["msg_type"]).
			Content(messageInfo["content"]).
			Build()).
		Build()
	resp, err := c.feishuClient.Im.V1.Message.Create(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	return resp.Data, nil
}

// UpdateMessage 编辑消息
func (c *FeishuClient) UpdateMessage(messageInfo map[string]string) (*larkim.UpdateMessageResp, error) {
	req := larkim.NewUpdateMessageReqBuilder().MessageId(messageInfo["message_id"]).
		Body(larkim.NewUpdateMessageReqBodyBuilder().
			MsgType(messageInfo["msg_type"]).
			Content(messageInfo["content"]).
			Build()).
		Build()
	resp, err := c.feishuClient.Im.V1.Message.Update(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	return resp, nil
}

// DeleteMessage 撤回消息
func (c *FeishuClient) DeleteMessage(messageId string) (*larkim.DeleteMessageResp, error) {
	req := larkim.NewDeleteMessageReqBuilder().MessageId(messageId).Build()
	resp, err := c.feishuClient.Im.V1.Message.Delete(c.CTX, req)
	if err != nil {
		return nil, fmt.Errorf("code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
	}
	return resp, nil
}

// GetDetail 查询审批事件详情
func (c *FeishuClient) GetDetail(instanceId, userId string) (otherDto.FeishuApiResponse, error) {
	// 创建请求对象
	respEnt := otherDto.FeishuApiResponse{}
	req := larkapproval.NewGetInstanceReqBuilder().
		InstanceId(instanceId).
		Locale(`zh-CN`).
		UserId(userId).
		UserIdType(`user_id`).
		Build()
	resp, err := c.feishuClient.Approval.V4.Instance.Get(context.Background(), req)
	// 处理错误
	if err != nil {
		return respEnt, err
	}

	// 服务端错误处理
	if !resp.Success() {
		return respEnt, fmt.Errorf("logId: %s, error response: \n%s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
	}

	// 业务处理
	jsonResp, err := json.Marshal(resp)
	jsonStr := string(jsonResp)
	err = json.Unmarshal([]byte(jsonStr), &respEnt)
	if err == nil && resp.Data.Form != nil {
		var formWidgets []otherDto.FormWidget
		if err = json.Unmarshal([]byte(*resp.Data.Form), &formWidgets); err == nil {
			respEnt.Data.Form = formWidgets
		}
	}
	return respEnt, err
}

// SendMessage 发送飞书信息
func (c *FeishuClient) SendMessage(userOpenid, msgType, jsonStr string) (*larkim.CreateMessageRespData, error) {
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(`open_id`).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(userOpenid).
			MsgType(msgType).
			Content(jsonStr).
			Build()).
		Build()
	resp, err := c.feishuClient.Im.V1.Message.Create(context.Background(), req)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("feishu Send Message Code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
		} else {
			return nil, fmt.Errorf("feishu Send Message error, err:%s", err.Error())
		}
	}
	// 服务端错误处理
	if !resp.Success() {
		return nil, fmt.Errorf("feishu logId: %s, error response: %s", resp.RequestId(), resp.CodeError.ErrorResp())
	}

	return resp.Data, nil
}

// Subscript 订阅应用
func (c *FeishuClient) Subscript(appCode string) (*larkapproval.SubscribeApprovalResp, error) {
	req := larkapproval.NewSubscribeApprovalReqBuilder().
		ApprovalCode(appCode).
		Build()
	// 发起请求
	resp, err := c.feishuClient.Approval.V4.Approval.Subscribe(context.Background(), req)

	// 处理错误
	if err != nil {
		return nil, err
	}

	// 服务端错误处理
	if !resp.Success() {
		err = fmt.Errorf("logId: %s, error response: \n%s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		return nil, err
	}
	return resp, nil
}

func (c *FeishuClient) TimeLineNodes(instanceCode, feishuUid, taskId string) error {
	reqData := larkapproval.PreviewInstanceReqBody{InstanceCode: &instanceCode, UserId: &feishuUid, TaskId: &taskId}
	req := larkapproval.NewPreviewInstanceReqBuilder().UserIdType("open_id").Body(&reqData).Build()
	resp, err := c.feishuClient.Approval.V4.Instance.Preview(context.Background(), req)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("feishu Approval Nodes Code:%d, msg:%s, err:%v", resp.Code, resp.Msg, resp.Error)
		} else {
			return fmt.Errorf("feishu Approval Nodes error, err:%s", err.Error())
		}
	}
	// 服务端错误处理
	if !resp.Success() {
		return fmt.Errorf("feishu Approval Nodes logId: %s, error response: %s", resp.RequestId(), resp.CodeError.ErrorResp())
	}
	fmt.Println(larkcore.Prettify(resp))
	return nil
}
