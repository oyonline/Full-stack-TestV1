package excelUtils

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 常量定义
const (
	ExcelTag          = "excel"
	DefaultSheetName  = "Sheet"
	DefaultDateFormat = "2006-01-02"
	DefaultScale      = 2
	DefaultWidth      = 15
	OperationImport   = "import"
	OperationExport   = "export"
	OperationBoth     = "both"
	MaxRowsSheet      = 65536 // Excel单Sheet最大行数限制（留出几行给表头）
)

// ExcelSingleStructureField Excel字段配置
type ExcelSingleStructureField struct {
	FieldName        string
	Title            string
	Sort             int
	Width            float64
	Type             string
	Required         bool
	Converter        map[string]string
	ReverseConverter map[string]string
	DictType         string
	Scale            int
	Operation        string
	DateFormat       string
}

// ExcelSingleStructureMapper Excel映射器
type ExcelSingleStructureMapper[T any] struct {
	fields       []ExcelSingleStructureField
	dictMappings map[string]map[string]string
}

// NewExcelSingleStructureMapper 创建Excel映射器
func NewExcelSingleStructureMapper[T any](dictMappings map[string]map[string]string) *ExcelSingleStructureMapper[T] {
	mapper := &ExcelSingleStructureMapper[T]{dictMappings: dictMappings}
	mapper.parseStructTags()
	return mapper
}

// 解析结构体标签
func (m *ExcelSingleStructureMapper[T]) parseStructTags() {
	var example T
	t := reflect.TypeOf(example)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fields := make([]ExcelSingleStructureField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		//解析excel标签
		excelTag := field.Tag.Get(ExcelTag)
		if excelTag == "" {
			continue
		}
		parts := strings.Split(excelTag, ",")
		title := strings.TrimSpace(parts[0])
		config := ExcelSingleStructureField{
			FieldName:        field.Name,
			Title:            title,
			Sort:             99,
			Width:            DefaultWidth,
			Type:             field.Type.Kind().String(),
			Required:         false,
			Converter:        make(map[string]string),
			ReverseConverter: make(map[string]string),
			DictType:         "",
			Scale:            DefaultScale,
			Operation:        OperationBoth,
			DateFormat:       DefaultDateFormat,
		}
		for _, part := range parts[1:] {
			m.parseTagPart(&config, part)
		}
		fields = append(fields, config)
	}
	// 按sort排序
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Sort < fields[j].Sort
	})
	m.fields = fields
}

// 解析标签部分
func (m *ExcelSingleStructureMapper[T]) parseTagPart(config *ExcelSingleStructureField, part string) {
	part = strings.TrimSpace(part)

	switch {
	case part == "required:true":
		config.Required = true
	case strings.HasPrefix(part, "sort:"):
		if val, err := strconv.Atoi(strings.TrimPrefix(part, "sort:")); err == nil {
			config.Sort = val
		}
	case strings.HasPrefix(part, "width:"):
		if val, err := strconv.ParseFloat(strings.TrimPrefix(part, "width:"), 64); err == nil {
			config.Width = val
		}
	case strings.HasPrefix(part, "converter:"):
		converterMap, reverseMap := parseConverter(strings.TrimPrefix(part, "converter:"))
		config.Converter = converterMap
		config.ReverseConverter = reverseMap
	case strings.HasPrefix(part, "dictType:"):
		config.DictType = strings.TrimPrefix(part, "dictType:")
	case strings.HasPrefix(part, "scale:"):
		if val, err := strconv.Atoi(strings.TrimPrefix(part, "scale:")); err == nil {
			config.Scale = val
		}
	case strings.HasPrefix(part, "operation:"):
		config.Operation = strings.TrimPrefix(part, "operation:")
	case strings.HasPrefix(part, "dateFormat:"):
		config.DateFormat = strings.TrimPrefix(part, "dateFormat:")
	}
}

// ImportFromExcel 从Excel导入
func (m *ExcelSingleStructureMapper[T]) ImportFromExcel(reader io.Reader, sheetName string) ([]T, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	actualSheetName := m.getSheetName(f, sheetName)
	rows, err := f.GetRows(actualSheetName)
	if err != nil || len(rows) <= 1 {
		return []T{}, nil
	}

	headerMap := m.createHeaderMap(rows[0])
	importFields := m.getFieldsByOperation(OperationImport)

	var results []T
	for i := 1; i < len(rows); i++ {
		item, err := m.rowToStruct(rows[i], headerMap, importFields)
		if err != nil {
			return nil, fmt.Errorf("第%d行解析失败: %v", i+1, err)
		}
		results = append(results, item)
	}

	return results, nil
}

// ExportToExcel 导出到Excel
func (m *ExcelSingleStructureMapper[T]) ExportToExcel(data []T, writer http.ResponseWriter, filename, sheetName string, operations string) error {
	writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	encodedFilename := url.QueryEscape(filename)
	writer.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s.xlsx\"; filename*=UTF-8''%s.xlsx", encodedFilename, encodedFilename))
	f := excelize.NewFile()
	sheetNameNew := DefaultSheetName
	if sheetName != "" {
		sheetNameNew = sheetName
	}
	exportFields := m.getFieldsByOperation(operations)
	totalCount := len(data)
	sheetCount := int(math.Max(1, math.Ceil(float64(totalCount)*1.0/float64(MaxRowsSheet))))
	for sheetIndex := 0; sheetIndex < sheetCount; sheetIndex++ {
		startIndex := sheetIndex * MaxRowsSheet
		endIndex := startIndex + MaxRowsSheet
		if endIndex > totalCount {
			endIndex = totalCount
		}
		currentSheetName := fmt.Sprintf("%s_%d", sheetNameNew, sheetIndex+1)
		batchData := data[startIndex:endIndex]
		// Sheet新建
		_, _ = f.NewSheet(currentSheetName)
		// 写入表头
		for i, field := range exportFields {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(currentSheetName, cell, field.Title)

			if field.Width > 0 {
				col, _ := excelize.ColumnNumberToName(i + 1)
				_ = f.SetColWidth(currentSheetName, col, col, field.Width)
			}

			// 导入模式下为字典和转换器字段设置下拉框
			if operations == OperationImport && (field.DictType != "" || len(field.Converter) > 0) {
				m.setDropdownValidation(f, currentSheetName, i+1, field)
			}
		}
		// 写入数据
		err := m.structToRow(batchData, exportFields, f, currentSheetName)
		if err != nil {
			return err
		}
	}
	//删除默认Sheet
	_ = f.DeleteSheet("Sheet1")
	return f.Write(writer)
}

// 获取Sheet名称
func (m *ExcelSingleStructureMapper[T]) getSheetName(f *excelize.File, sheetName string) string {
	if sheetName != "" {
		return sheetName
	}

	sheetNames := f.GetSheetList()
	if len(sheetNames) > 0 {
		return sheetNames[0]
	}

	return DefaultSheetName
}

// 根据操作类型获取字段
func (m *ExcelSingleStructureMapper[T]) getFieldsByOperation(operations string) []ExcelSingleStructureField {
	result := make([]ExcelSingleStructureField, 0, len(m.fields))
	for _, field := range m.fields {
		if field.Operation == operations || field.Operation == OperationBoth {
			result = append(result, field)
		}
	}
	return result
}

// 创建标题映射
func (m *ExcelSingleStructureMapper[T]) createHeaderMap(headers []string) map[string]int {
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}
	return headerMap
}

// 行转结构体
func (m *ExcelSingleStructureMapper[T]) rowToStruct(row []string, headerMap map[string]int, fields []ExcelSingleStructureField) (T, error) {
	var result T
	v := reflect.ValueOf(&result).Elem()

	for _, field := range fields {
		var colValue string
		colIdx, exists := headerMap[field.Title]
		if !exists {
			continue
		}
		if !(colIdx >= len(row)) {
			colValue = row[colIdx]
		}

		value, err := m.processValueForImport(colValue, field)
		if err != nil {
			return result, err
		}

		fieldValue := v.FieldByName(field.FieldName)
		if fieldValue.IsValid() && fieldValue.CanSet() {
			if err := setValueByType(fieldValue, value, field); err != nil {
				return result, fmt.Errorf("字段 %s 设置值失败: %v", field.FieldName, err)
			}
		}
	}

	return result, nil
}

// 处理导入时的值转换
func (m *ExcelSingleStructureMapper[T]) processValueForImport(value string, field ExcelSingleStructureField) (string, error) {
	// 优先使用converter
	if len(field.ReverseConverter) > 0 {
		if value1, exists := field.ReverseConverter[value]; exists {
			return value1, nil
		} else {
			return value, fmt.Errorf("字段 '%s' 字典转换失败", field.Title)
		}
	}
	// 使用预定义字典
	if field.DictType != "" {
		if dict, ok := m.dictMappings[field.DictType]; ok {
			for k, v := range dict {
				if v == value {
					return k, nil
				}
			}
			return value, fmt.Errorf("字段 '%s' 字典转换失败", field.Title)
		}
	}
	if value == "" && field.Required {
		return value, fmt.Errorf("字段 '%s' 是必填项", field.Title)
	}
	return value, nil
}

// 结构体转行
func (m *ExcelSingleStructureMapper[T]) structToRow(batchData []T, fields []ExcelSingleStructureField, f *excelize.File, currentSheetName string) error {

	for i, item := range batchData {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		for j, field := range fields {
			fieldValue := v.FieldByName(field.FieldName)
			if !fieldValue.IsValid() {
				continue
			}

			value := m.processValueForExport(fieldValue.Interface(), field)

			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			_ = f.SetCellValue(currentSheetName, cell, value)

		}
	}
	return nil
}

// 处理导出时的值转换
func (m *ExcelSingleStructureMapper[T]) processValueForExport(value interface{}, field ExcelSingleStructureField) interface{} {
	// strValue := fmt.Sprintf("%v", value)
	// 优先使用converter
	if display, exists := field.Converter[fmt.Sprintf("%v", value)]; exists {
		return display
	}
	if dict, ok := m.dictMappings[field.DictType]; ok {
		if display, exists := dict[fmt.Sprintf("%v", value)]; exists {
			return display
		}
	}
	if field.Type == "ptr" {
		vReflect := reflect.ValueOf(value)
		if vReflect.IsNil() {
			return nil
		}
		value = vReflect.Elem().Interface()
	}

	// 5. 根据字段类型特殊处理
	switch v := value.(type) {
	case time.Time:
		// 时间类型：格式化输出
		if !v.IsZero() {
			return v.Format(field.DateFormat)
		}
		return ""

	case float32, float64:
		// 根据 scale 设置保留小数位数
		return fmt.Sprintf("%.*f", field.Scale, value)
	case bool:
		// 布尔类型：转换为中文
		if v {
			return "是"
		}
		return "否"
	}
	return value
}

// 辅助函数：解析转换器
func parseConverter(converterStr string) (map[string]string, map[string]string) {
	forwardMap := make(map[string]string) // 导出用：实际值->显示值
	reverseMap := make(map[string]string) // 导入用：显示值->实际值
	pairs := strings.Split(converterStr, "|")

	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), "=")
		if len(kv) == 2 {
			//正向转换
			forwardMap[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			//反向转换
			reverseMap[strings.TrimSpace(kv[1])] = strings.TrimSpace(kv[0])
		}
	}

	return forwardMap, reverseMap
}

// 辅助函数：按类型设置值
func setValueByType(fieldValue reflect.Value, value string, field ExcelSingleStructureField) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			fieldValue.SetInt(i)
		} else {
			return err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(value, 10, 64); err == nil {
			fieldValue.SetUint(u)
		} else {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			fieldValue.SetFloat(f)
		} else {
			return err
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(value); err == nil {
			fieldValue.SetBool(b)
		} else {
			return err
		}
	case reflect.Ptr:
		if fieldValue.Type().Elem().PkgPath() == "time" && fieldValue.Type().Elem().Name() == "Time" {

			if parsedTime, err := parseSmartDate(value, field.DateFormat); err == nil {
				fieldValue.Set(reflect.ValueOf(parsedTime))
			}
		}
	default:
		panic("unhandled default case")
	}
	return nil
}

// setDropdownValidation 设置下拉验证
func (m *ExcelSingleStructureMapper[T]) setDropdownValidation(f *excelize.File, sheetName string, colIndex int, field ExcelSingleStructureField) {
	dropdownValues := make([]string, 0)
	// 优先使用 Converter 的 key（显示值）作为下拉选项
	if len(field.Converter) > 0 {
		for _, displayValue := range field.Converter {
			dropdownValues = append(dropdownValues, displayValue)
		}
	}
	// 使用预定义字典的 key 作为下拉选项
	if field.DictType != "" {
		if dict, ok := m.dictMappings[field.DictType]; ok {
			for _, v := range dict {
				dropdownValues = append(dropdownValues, v)
			}
		}
	}
	if len(dropdownValues) == 0 {
		return
	}

	colName, _ := excelize.ColumnNumberToName(colIndex)
	// Excel 下拉框公式，值之间用逗号分隔
	formula := "\"" + strings.Join(dropdownValues, ",") + "\""
	var errorTitle = "输入无效"
	var errorStr = "请从下拉列表中选择有效值"
	dd := &excelize.DataValidation{
		Type:             "list",
		Formula1:         formula,
		AllowBlank:       true,
		Sqref:            fmt.Sprintf("%s2:%s%d", colName, colName, MaxRowsSheet),
		ShowInputMessage: true,
		ShowErrorMessage: true,
		ErrorTitle:       &errorTitle,
		Error:            &errorStr,
	}

	if err := f.AddDataValidation(sheetName, dd); err != nil {
		fmt.Printf("设置下拉框失败：%v\n", err)
	}
}

// parseSmartDate 智能解析日期（自动适配所有常见格式）
func parseSmartDate(value string, dateFormat string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	if t, err := time.Parse(dateFormat, value); err == nil {
		return &t, nil
	} else {
		for _, format := range DateLayouts {
			if t, err = time.Parse(format, value); err == nil {
				return &t, nil
			}
		}
	}
	return nil, fmt.Errorf("无法解析日期：%s", value)
}

var DateLayouts = []string{
	// === 标准格式（双数月日）===
	"2006-01-02",
	"2006/01/02",
	"2006.01.02",

	// === 单数格式（重要补充）===
	"2006-1-2",
	"2006/1/2",
	"2006.1.2",

	// === 中文格式 ===
	"2006 年 01 月 02 日",
	"2006 年 1 月 2 日",
	"2006 年 01 月",
	"2006 年 1 月",

	// === 欧美格式（月/日/年）===
	"01/02/2006",
	"1/2/2006",
	"02/01/2006",
	"2/1/2006",
	"01-02-2006",
	"1-2-2006",
	"02-01-2006",
	"2-1-2006",

	// === 简写年份格式 ===
	"01-02-06",
	"1-2-06",
	"02-01-06",
	"2-1-06",
	"01/02/06",
	"1/2/06",
	"02/01/06",
	"2/1/06",
	"01.02.06",
	"1.2.06",
	"02.01.06",
	"2.1.06",

	// === 带时间的格式 ===
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
	"2006-1-2 15:04",
	"2006-1-2 15:04:05",
	"2006/01/02 15:04",
	"2006/01/02 15:04:05",
	"2006/1/2 15:04",
	"2006/1/2 15:04:05",
	"2006.01.02 15:04",
	"2006.01.02 15:04:05",
	"2006.1.2 15:04",
	"2006.1.2 15:04:05",

	// === 带时间的简写格式 ===
	"01/02/06 15:04",
	"1/2/06 15:04",
	"02/01/06 15:04",
	"2/1/06 15:04",
	"01-02-06 15:04",
	"1-2-06 15:04",
	"02-01-06 15:04",
	"2-1-06 15:04",
	"01/02/06 15:04:05",
	"1/2/06 15:04:05",
	"02/01/06 15:04:05",
	"2/1/06 15:04:05",

	// === ISO 格式 ===
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05+08:00",

	// === 纯时间格式 ===
	"2006-01-02 15:04:05.999",
	"2006/01/02 15:04:05.999",

	// === 紧凑格式 ===
	"20060102",
	"20060102150405",

	// === 英文月份格式 ===
	"Jan 2, 2006",
	"January 2, 2006",
	"2 Jan 2006",
	"2 January 2006",
	"02 Jan 2006",
	"02 January 2006",
	"Jan 02, 2006",
	"January 02, 2006",
}
