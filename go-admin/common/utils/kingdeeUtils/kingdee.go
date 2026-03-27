package kingdeeUtils

import (
	"encoding/json"
	"fmt"
	"go-admin/common/utils/kingdeeUtils/kingdeeModels"
	"go-admin/common/utils/structsUtils"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/service"
)

type KingDee struct {
	service.Service
}

var host = "http://192.168.1.201/k3cloud/"
var loginApi = "Kingdee.BOS.WebApi.ServicesStub.AuthService.ValidateUser.common.kdsvc"

var purchaseOrderExecuteDetailApi = "Kingdee.BOS.WebApi.ServicesStub.DynamicFormService.GetSysReportData.common.kdsvc"

// 登录头
var header = make(map[string]string)
var expireAt = ""

type kingdeeRequest struct {
	FormId string                            `json:"formid"`
	Data   kingdeePurchaseOrderExecuteDetail `json:"data"`
}

type filterString struct {
	FieldName string `json:"FieldName"`
	Compare   string `json:"Compare"`
	Value     string `json:"Value"`
	Left      string `json:"Left"`
	Right     string `json:"Right"`
	Logic     string `json:"Logic"`
}
type kingdeePurchaseOrderExecuteDetail struct {
	FieldKeys             string         `json:"FieldKeys"`
	SchemeId              string         `json:"SchemeId"`
	StartRow              int            `json:"StartRow"`
	Limit                 int            `json:"Limit"`
	IsVerifyBaseDataField string         `json:"IsVerifyBaseDataField"`
	FilterString          []filterString `json:"FilterString"`
	Model                 executeModel   `json:"Model"`
}

type executeModel struct {
	FPurchaseOrgIdList string `json:"FPurchaseOrgIdList"`
	FOrderStartDate    string `json:"FOrderStartDate"`
	FOrderEndDate      string `json:"FOrderEndDate"`
	FBeginBillNumber   string `json:"FBeginBillNumber"`
	FEndBillNumber     string `json:"FEndBillNumber"`
	FBusinessType      string `json:"FBusinessType"`
	FDocumentStatus    string `json:"FDocumentStatus"`
	FLineStatus        string `json:"FLineStatus"`
}

type kingdeeResponse struct {
	Result kingdeeResult
}
type kingdeeResult struct {
	RowCount int        `json:"RowCount"`
	Rows     [][]string `json:"Rows"`
}

func newReq(billNos ...string) kingdeeRequest {
	d := kingdeePurchaseOrderExecuteDetail{}
	d.FieldKeys = kingdeeModels.KingdeePurchaseOrderField
	d.Limit = 2000
	d.IsVerifyBaseDataField = "true"
	dateFrom := time.Now().AddDate(0, -6, 0).Format(time.DateOnly)
	dateTo := time.Now().AddDate(0, 0, 1).Format(time.DateOnly)
	mod := executeModel{
		FPurchaseOrgIdList: "1,100072,100073,100074,100075,100076,100077,6206749,6453536,320791,6471574,100080,100081,100082,100084,100469",
		FOrderStartDate:    dateFrom,
		FOrderEndDate:      dateTo,
		FBusinessType:      "A",
		FDocumentStatus:    "C",
		FLineStatus:        "A",
	}
	if len(billNos) > 0 {
		if len(billNos) > 1 {
			billNoStrs := strings.Join(billNos, ",")
			filters := filterString{
				FieldName: "FBillNo",
				Compare:   "338",
				Value:     billNoStrs,
				Left:      "(",
				Right:     ")",
				Logic:     "0",
			}
			d.FilterString = append(d.FilterString, filters)
		} else {
			filters := filterString{
				FieldName: "FBillNo",
				Compare:   "67",
				Value:     billNos[0],
				Left:      "",
				Right:     "",
				Logic:     "0",
			}
			d.FilterString = append(d.FilterString, filters)
		}

	}
	d.Model = mod
	return kingdeeRequest{
		FormId: "PUR_PurchaseOrderDetailRpt",
		Data:   d,
	}
}

func (e *KingDee) PullPurchaseOrderExecuteDetail(billNos ...string) error {
	err := login()
	if err != nil {
		return err
	}
	url := host + purchaseOrderExecuteDetailApi
	req := newReq(billNos...)
	jsonByte, _ := json.Marshal(req)
	jsonStr := string(jsonByte)
	resp, err := httpPost(url, jsonStr)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(resp.Body)
	respBody, err := io.ReadAll(resp.Body)
	respStr := string(respBody)
	if strings.Index(respStr, "\"IsSuccess\":true") >= 0 {
		bodys := kingdeeResponse{}
		err := json.Unmarshal(respBody, &bodys)
		if err != nil {
			return err
		}
		if bodys.Result.RowCount > 0 {
			list := make([]kingdeeModels.KingdeePurchaseOrderExecuteDetail, 0)
			main := kingdeeModels.KingdeePurchaseOrderExecuteDetail{}
			for _, row := range bodys.Result.Rows {
				it := kingdeeModels.KingdeePurchaseOrderExecuteDetail{}
				it.MapToKingDeePurchaseOrderExecuteDetail(row)
				if it.BillNo != "" {
					structsUtils.CopyBeanProp(&main, &it)
				}
				it.BillNo = main.BillNo
				it.FDate = main.FDate
				it.SupplierName = main.SupplierName
				it.Purchaser = main.Purchaser
				list = append(list, it)
			}
			if len(list) > 0 {
				e.Orm.Model(&kingdeeModels.KingdeePurchaseOrderExecuteDetail{}).Delete(&kingdeeModels.KingdeePurchaseOrderExecuteDetail{}, "bill_no in ?", billNos)
				e.Orm.Model(&main).CreateInBatches(&list, len(list))
			}
		}
	}
	return err
}

type KingdeeLogin struct {
	AcctID   string `json:"acctId"`
	UserName string `json:"username"`
	Password string `json:"password"`
	Lcid     int    `json:"lcid"`
}

func kingdeeLoginParams() string {
	params := KingdeeLogin{
		AcctID:   "645767c7e35199",
		UserName: "自研对接",
		Password: "SjFpTuWGFpkuq3j",
		Lcid:     2052,
	}
	jsonStr, _ := json.Marshal(params)
	return string(jsonStr)
}

func login() error {
	if _, ok := header["Cookie"]; ok && expireAt != "" && expireAt < time.Now().Format(time.DateTime) {
		return nil
	}
	params := kingdeeLoginParams()
	url := host + loginApi
	resp, err := httpPost(url, params)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(resp.Body)

	var loginCookie string
	for key, values := range resp.Header {
		if strings.EqualFold(key, "Set-Cookie") {
			for _, cookie := range values {
				if strings.HasPrefix(cookie, "kdservice-sessionid") {
					loginCookie = cookie
					break
				}
			}
		}
	}

	if loginCookie == "" {
		return fmt.Errorf("kingdee获取Cookie失败")
	}

	header["Cookie"] = loginCookie
	expireAt = time.Now().Add(time.Second * 120).Format(time.DateTime)
	return err
}

// 模拟 HttpUtil.httpPost 的功能
func httpPost(url string, body string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.Index(url, "AuthService") < 0 {
		req.Header.Set("Cookie", header["Cookie"])
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
