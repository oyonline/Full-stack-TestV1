package excelUtils

import (
	"bytes"
	"errors"
	"fmt"
	"go-admin/common/utils"
	"go-admin/common/utils/dateUtils"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type UploadError struct {
	Code int
	Err  error
	Msg  string
}

// ExportExcel 导出EXCEL
func ExportExcel[T any](list []T, sheetName string) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	if len(list) == 0 {
		return &buffer, nil
	}
	f, err := exportExcel(list, sheetName)
	// 将 Excel 文件写入内存中的缓冲区
	if err = f.Write(&buffer); err != nil {
		return nil, err
	}
	return &buffer, nil
}

func exportExcel[T any](list []T, sheetName string) (*excelize.File, error) {
	f := excelize.NewFile()
	defaultSheet := f.GetSheetName(0)
	err := f.SetSheetName(defaultSheet, sheetName)
	f.SetActiveSheet(0)
	if len(list) == 0 {
		return nil, nil
	}
	firstElem := reflect.ValueOf(list[0])
	elemType := firstElem.Type()
	row := 1
	col := 1
	headStyle := excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Font:      &excelize.Font{Bold: true},
	}
	headerStyle, _ := f.NewStyle(&headStyle)
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		cell := fmt.Sprintf("%s%d", colNumToExcelCol(col), row)
		head := field.Tag.Get("excel")
		if head != "" && head != "-" {
			err = f.SetCellValue(sheetName, cell, head)
			err = f.SetCellStyle(sheetName, cell, cell, headerStyle)
			if err != nil {
				return nil, err
			}
			col++
		}
	}
	// 写入数据行
	for i, item := range list {
		v := reflect.ValueOf(item)
		row = i + 2 // 数据从第二行开始
		col = 1
		for j := 0; j < v.NumField(); j++ {
			ex := elemType.Field(j).Tag.Get("excel")
			cellName := fmt.Sprintf("%s%d", colNumToExcelCol(col), row)
			if ex != "" && ex != "-" {
				value, _ := utils.ToString(v.Field(j).Interface())
				err = f.SetCellValue(sheetName, cellName, value)
				if err != nil {
					return nil, err
				}
				col++
			}
		}
	}
	return f, err
}

func colNumToExcelCol(colNum int) string {
	var letters []rune
	for colNum > 0 {
		colNum--
		letter := (colNum % 26) + 'A'
		letters = append([]rune{rune(letter)}, letters...)
		colNum = colNum / 26
	}
	return string(letters)
}

type modelTable interface {
	SheetName() string
}

// UploadExcelStream 获取文件流
func UploadExcelStream(c *gin.Context) (fileStream *excelize.File, upErr UploadError) {
	file, err := c.FormFile("file")
	if err != nil {
		upErr = UploadError{http.StatusBadRequest, err, "上传失败: " + err.Error()}
		return
	}

	tempFile, err := os.CreateTemp("temp/upload", "upload-*.xlsx")
	if err != nil {
		upErr = UploadError{http.StatusBadRequest, err, "创建文件失败“-"}
		return
	}
	// 将上传文件保存到临时文件
	if err = c.SaveUploadedFile(file, tempFile.Name()); err != nil {
		upErr = UploadError{http.StatusBadRequest, err, "保存文件失败"}
		return
	}

	fileStream, err = excelize.OpenFile(tempFile.Name())
	defer func(f *excelize.File) {
		_ = f.Close()
	}(fileStream)
	if err != nil {
		upErr = UploadError{http.StatusBadRequest, err, "解析 Excel 文件失败"}
		return
	}
	return
}

// UploadExcel 上传EXCEL
func UploadExcel[T modelTable](c *gin.Context, notNullField *[]string, sheetName string) (result []T, upErr UploadError) {
	file, err := c.FormFile("file")
	if err != nil {
		upErr = UploadError{http.StatusBadRequest, err, "上传失败: " + err.Error()}
		return
	}

	// 创建临时文件保存上传的文件
	tempFile, err := os.CreateTemp("temp/upload", "upload-*.xlsx")
	if err != nil {
		upErr = UploadError{http.StatusBadRequest, err, "创建文件失败“-"}
		return
	}
	// 将上传文件保存到临时文件
	if err := c.SaveUploadedFile(file, tempFile.Name()); err != nil {
		upErr = UploadError{http.StatusBadRequest, err, "保存文件失败"}
		return
	}

	// 打开临时文件以供读取
	f, err := excelize.OpenFile(tempFile.Name())
	defer func(f *excelize.File) {
		_ = f.Close()
	}(f)
	if err != nil {
		upErr = UploadError{http.StatusBadRequest, err, "解析 Excel 文件失败"}
		return
	}
	if sheetName == "" {
		sheetName = f.GetSheetList()[0]
	}
	upErr = LoadExcel(f, sheetName, notNullField, &result)
	return
}

func LoadExcel[T modelTable](file *excelize.File, sheetName string, notNullField *[]string, result *[]T) UploadError {

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return UploadError{http.StatusBadRequest, err, "读取行数据失败"}
	}
	headers := make(map[string]int)

	// 获取表头并建立索引映射
	for colIndex, header := range rows[0] {
		headers[header] = colIndex
	}

	// 反射获取 T 类型的结构体字段信息
	tType := reflect.TypeOf((*T)(nil)).Elem()
	v := reflect.New(tType).Elem()

	// 创建一个字段名到结构体字段的映射
	fieldMap := make(map[string]reflect.StructField)
	headFields := make([]string, 0)
	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		tag := field.Tag.Get("excel")
		if tag != "" && tag != "-" {
			fieldMap[tag] = field
			headFields = append(headFields, tag)
		}
	}
	if notNullField != nil && len(*notNullField) > 0 {
		isPass, missing := utils.ContainsAllRequired(headFields, *notNullField)
		if !isPass {
			msg := fmt.Sprintf("缺少必填字段 %s", strings.Join(missing, ","))
			err = errors.New(msg)
			return UploadError{http.StatusBadRequest, err, msg}
		}
	}

	// 遍历每一行数据并映射到结构体
	for rowIndex, row := range rows {
		if rowIndex == 0 { // 跳过表头
			continue
		}

		v = reflect.New(tType).Elem() // 每次循环都创建新的实例

		for excelTag, structField := range fieldMap {
			colIndex, ok := headers[excelTag]
			if !ok {
				continue // 如果 Excel 中没有对应的列，则跳过
			}
			cellValue := ""
			if colIndex <= len(row)-1 {
				cellValue = strings.TrimSpace(row[colIndex])
			}
			if notNullField != nil && len(*notNullField) > 0 && slices.Contains(*notNullField, excelTag) && cellValue == "" {
				msg := fmt.Sprintf("第%d行 %s 字段为空", rowIndex+1, excelTag)
				err = errors.New(msg)
				return UploadError{http.StatusBadRequest, err, msg}
			}
			if strings.Contains(excelTag, "时间") || strings.Contains(excelTag, "日期") {
				cellValue, err = excelDate(cellValue)
			}
			// 根据字段类型设置值
			switch structField.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if cellValue != "" {
					if cellValue == "是" {
						cellValue = "1"
					}
					if cellValue == "否" {
						cellValue = "0"
					}
					value, err := strconv.ParseInt(cellValue, 10, 64)
					if err != nil {
						return UploadError{http.StatusBadRequest, err, fmt.Sprintf("转换 %s 到 int 失败", excelTag)}
					}
					v.FieldByName(structField.Name).SetInt(value)
				} else {
					v.FieldByName(structField.Name).SetInt(0)
				}
			case reflect.Float32, reflect.Float64:
				if cellValue != "" {
					value, err := strconv.ParseFloat(cellValue, 64)
					if err != nil {
						return UploadError{http.StatusBadRequest, err, fmt.Sprintf("转换 %s 到 float 失败", excelTag)}
					}
					v.FieldByName(structField.Name).SetFloat(value)
				} else {
					v.FieldByName(structField.Name).SetFloat(0)
				}
			default:
				cellValue = fmt.Sprintf("%v", cellValue)
				v.FieldByName(structField.Name).SetString(cellValue)
			}
		}

		*result = append(*result, v.Interface().(T))
	}
	return UploadError{200, nil, "ok"}
}

func excelDate(dateStr string) (string, error) {
	format := time.DateOnly
	for _, layout := range dateUtils.PossibleLayouts {
		t, err := time.Parse(layout, dateStr)
		if err == nil {
			if len(layout) == 8 {
				currentYear := time.Now().Year()
				yearSuffix := t.Year() % 100
				if yearSuffix < 70 {
					t = time.Date(currentYear/100*100+yearSuffix, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
				} else {
					t = time.Date((currentYear/100-1)*100+yearSuffix, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
				}
			}
			if t.Hour()+t.Minute()+t.Second() > 0 {
				format = time.DateTime
			}
			return t.Format(format), nil
		}
	}

	return "", fmt.Errorf("无法解析日期: %s", dateStr)
}
