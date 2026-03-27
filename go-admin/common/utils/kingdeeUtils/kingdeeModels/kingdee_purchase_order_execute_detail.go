package kingdeeModels

import (
	"strconv"
	"strings"
	"time"
)

type KingdeePurchaseOrderExecuteDetail struct {
	ID            int64  `json:"id" gorm:"column:id;type:bigint;primaryKey;comment:唯一标识"`
	SupplierName  string `json:"FSUPPLIERNAME" gorm:"column:supplier_name;type:varchar(100);" comment:"供应商"`
	Sku           string `json:"FMATERIALID" gorm:"column:sku;type:varchar(100);" comment:"SKU"`
	SkuName       string `json:"FMATERIALNAME" gorm:"column:sku_name;type:varchar(200);" comment:"SKU名称"`
	Purchaser     string `json:"FPURCHASERID" gorm:"column:purchaser;type:varchar(100);" comment:"采购员"`
	DeliveryDate  string `json:"FDELIVERYDATE" gorm:"column:delivery_date;type:varchar(25);" comment:"预计交付时间"`
	UnitId        string `json:"FUnitId" gorm:"column:unit_id;type:varchar(10);" comment:"单位"`
	BillNo        string `json:"FBillNo" gorm:"column:bill_no;type:varchar(50);" comment:"采购单号"`
	FSrcBillNo    string `json:"FSrcBillNo" gorm:"column:src_bill_no;type:varchar(100)" comment:"源单单号"`
	FDate         string `json:"FDate" gorm:"column:order_date;type:varchar(25);" comment:"下单时间"`
	OrderQty      int    `json:"FOrderQty" gorm:"column:order_qty;type:int;" comment:"下单量"`
	ReceiveNumber string `json:"FReceiveNumber" gorm:"column:receive_number;type:varchar(100);" comment:"收料单号"`
	ReceiveDate   string `json:"FReceiveDate" gorm:"column:receive_date;type:varchar(25);" comment:"收料日期"`
	ReceiveQty    int    `json:"FReceiveQty" gorm:"column:receive_qty;type:int;" comment:"收料数量"`
	EnterNumber   string `json:"FEnterNumber" gorm:"column:enter_number;type:varchar(100);" comment:"入库单号"`
	EnterDate     string `json:"FEnterDate" gorm:"column:enter_date;type:varchar(25);" comment:"入库日期"`
	ImportQty     int    `json:"FImportQty" gorm:"column:enter_qty;type:int;" comment:"入库数量"`
	ReturnNumber  string `json:"FReturnNumber" gorm:"column:return_number;type:varchar(100);" comment:"退料单号"`
	ReturnDate    string `json:"FReturnDate" gorm:"column:return_date;type:varchar(25);" comment:"退料日期"`
	ReturnQty     int    `json:"FReturnQty" gorm:"column:return_qty;type:int;" comment:"退料数量"`
	MrpClose      string `json:"FMRPCLOSESTATUS" gorm:"column:mrp_close;type:varchar(20);" comment:"行业无关闭"`
	OrderClose    string `json:"FCLOSESTATUS" gorm:"column:order_close;type:varchar(20);" comment:"整单关闭"`
	MrpTerminate  string `json:"FMRPTERMINATESTATUS" gorm:"column:mrp_terminate;type:varchar(20);" comment:"行业务终止"`
}

func (k *KingdeePurchaseOrderExecuteDetail) MapKey() string {
	return k.BillNo + ":" + k.Sku
}

func (KingdeePurchaseOrderExecuteDetail) TableName() string {
	return "kingdee_purchase_order_execute_detail"
}
func (KingdeePurchaseOrderExecuteDetail) TableDesc() string { return "金蝶采购订单执行明细" }
func (KingdeePurchaseOrderExecuteDetail) DbName() string    { return "*" }

var KingdeePurchaseOrderField = "FBillNo,FDate,FSUPPLIERNAME,FPURCHASERID,FMATERIALID,FMATERIALNAME,FDELIVERYDATE,FUnitId,FOrderQty,FReceiveNumber,FReceiveDate,FReceiveQty,FEnterNumber,FEnterDate,FImportQty,FReturnNumber,FReturnQty,FReturnDate,FMRPCLOSESTATUS,FCLOSESTATUS,FMRPTERMINATESTATUS,FSrcBillNo"

var kingdeePurchaseOrderFieldKeys = strings.Split(KingdeePurchaseOrderField, ",")

// 字符串转整数工具函数（忽略空值）
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", "", -1)
	if s == "" || s == " " {
		return 0
	}
	i, _ := strconv.Atoi(s)
	return i
}

// 时间格式化工具函数
func formatDateString(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" || dateStr == " " {
		return ""
	}
	t, err := time.Parse("2006/1/2", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02")
}

// MapToKingDeePurchaseOrderExecuteDetail 核心映射函数
func (item *KingdeePurchaseOrderExecuteDetail) MapToKingDeePurchaseOrderExecuteDetail(row []string) {
	for i, key := range kingdeePurchaseOrderFieldKeys {
		val := strings.TrimSpace(row[i])
		switch key {
		case "FBillNo":
			item.BillNo = val
		case "FDate":
			item.FDate = formatDateString(val)
		case "FSUPPLIERNAME":
			item.SupplierName = val
		case "FPURCHASERID":
			item.Purchaser = val
		case "FMATERIALID":
			item.Sku = val
		case "FMATERIALNAME":
			item.SkuName = val
		case "FDELIVERYDATE":
			item.DeliveryDate = formatDateString(val)
		case "FUnitId":
			item.UnitId = val
		case "FOrderQty":
			item.OrderQty = parseInt(val)
		case "FReceiveNumber":
			item.ReceiveNumber = val
		case "FReceiveDate":
			item.ReceiveDate = formatDateString(val)
		case "FReceiveQty":
			item.ReceiveQty = parseInt(val)
		case "FEnterNumber":
			item.EnterNumber = val
		case "FEnterDate":
			item.EnterDate = formatDateString(val)
		case "FImportQty":
			item.ImportQty = parseInt(val)
		case "FReturnNumber":
			item.ReturnNumber = val
		case "FReturnQty":
			item.ReturnQty = parseInt(val)
		case "FReturnDate":
			item.ReturnDate = formatDateString(val)
		case "FMRPCLOSESTATUS":
			item.MrpClose = val
		case "FCLOSESTATUS":
			item.OrderClose = val
		case "FMRPTERMINATESTATUS":
			item.MrpTerminate = val

		}
	}
}
