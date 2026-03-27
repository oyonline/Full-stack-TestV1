package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-admin/common/utils/structsUtils"
	"reflect"
	"runtime"
	"slices"
	"strings"

	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB

func GetDb() *gorm.DB {
	if dbInstance == nil {
		dbs := sdk.Runtime.GetDb()
		for _, db := range dbs {
			if db != nil {
				dbInstance = db
				break
			}
		}
	}
	return dbInstance
}

type TableNamer interface {
	TableName() string
}

func insertBatchProcesser(v reflect.Value, tableName string, list *[]interface{}) string {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			err := r.(error)
			log.Error(err.Error())
			fmt.Printf("Stack trace: %s\n%s\n", err.Error(), string(buf[:n]))
			return
		}
	}()
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	var fields, values, updates, autoDateFields, jsonField, numberField []string
	systemField := []string{"Model", "ModelTime", "ControlBy", "`Model`", "`ModelTime`", "`ControlBy`"}
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		//if field.Tag.Get("json") == "-" {
		//	continue
		//}
		var name string
		tag := field.Tag.Get("gorm")
		if tag != "-" {
			isAutoDate := false
			for _, tv := range parseGormTag(tag) {
				if key, ok := tv["column"]; ok {
					name = "`" + key + "`"
				}
				if key, ok := tv["type"]; ok && key == "json" {
					jsonField = append(jsonField, name)
				}
				if key, ok := tv["type"]; ok && (strings.Contains(key, "int") || strings.Contains(key, "decimal")) {
					numberField = append(numberField, name)
				}
				if def, ok := tv["default"]; ok && strings.ToUpper(def) == "CURRENT_TIMESTAMP" {
					isAutoDate = true
				}
			}
			if name == "" {
				name = "`" + CamelToSnake(field.Name) + "`"
			}
			if !slices.Contains(systemField, name) {
				if !isAutoDate {
					fields = append(fields, name)
				} else {
					autoDateFields = append(autoDateFields, name)
				}
				updates = append(updates, fmt.Sprintf("%s=VALUES(%s)", name, name))
			}
		}
	}

	for _, item := range *list {
		jsonBytes, _ := json.Marshal(item)
		modelData, _ := JsonStringToMap(string(jsonBytes))
		modelData = ConvertKeysToUnderscore(modelData)
		if id, isZero := modelData["id"]; isZero && fmt.Sprintf("%v", id) == "0" {
			modelData["id"] = "null"
		}
		var value []string
		for _, key := range fields {
			cleanKey := strings.ReplaceAll(key, "`", "")
			if mv, ok := modelData[cleanKey]; ok && !slices.Contains(autoDateFields, key) {
				var strValue string
				if IsArray(mv) || IsMap(mv) {
					bytesVal, _ := json.Marshal(mv)
					strValue = string(bytesVal)
				} else if slices.Contains(jsonField, key) && (mv == "" || mv == nil) {
					strValue = "[]"
				} else if slices.Contains(numberField, key) && (mv == "" || mv == nil) {
					strValue = "0"
				} else {
					strValue = fmt.Sprintf("%v", mv)
					strValue = strings.ReplaceAll(strValue, `'`, "`")
				}
				if strValue == "null" {
					value = append(value, strValue)
				} else {
					value = append(value, "'"+strValue+"'")
				}
			} else {
				fmt.Println(key)
			}
		}
		if len(value) > 0 {
			values = append(values, "("+strings.Join(value, ",")+")")
		}
	}
	if len(values) == 0 {
		return ""
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s",
		tableName,
		strings.Join(fields, ","),
		strings.Join(values, ","),
		strings.Join(updates, ","),
	)
	return query
}

func InsertBatchOnDuplicateByModel[T TableNamer](list *[]T) error {
	if len(*list) == 0 {
		return nil
	}
	interfaceList := make([]interface{}, 0)
	tableName := (*list)[0].TableName()
	v := reflect.ValueOf((*list)[0])
	for _, i := range *list {
		interfaceList = append(interfaceList, i)
	}
	query := insertBatchProcesser(v, tableName, &interfaceList)
	if query == "" {
		return nil
	}
	db := GetDb()
	err := db.Exec(query).Error
	return err
}

// InsertBatchOnDuplicate 批量插入
func InsertBatchOnDuplicate(list *[]interface{}, db *gorm.DB) (*gorm.DB, error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			err := r.(error)
			log.Error(err.Error())
			fmt.Printf("Stack trace: %s\n%s\n", err.Error(), string(buf[:n]))
			return
		}
	}()
	if list == nil || len(*list) == 0 {
		return db, nil
	}
	var tableName string
	if setter, ok := (*list)[0].(TableNamer); ok {
		tableName = setter.TableName()
		v := reflect.ValueOf(setter)

		if db == nil {
			db = GetDb()
		}
		query := insertBatchProcesser(v, tableName, list)
		if query == "" {
			return db, nil
		}
		err := db.Exec(query).Error
		return db, err
	}
	return db, fmt.Errorf("not interface TableNamer")
}

type MutiStruct interface {
	TableName() string
	SetStringID() interface{}
	GetStringID() string
}

type SubStruct interface {
	TableName() string
	SetRelationID(id string) interface{}
	DeleteRecord(relationIds []string, tx *gorm.DB)
}

// BatchSaveMutiStructOnDuplicate 批量保存嵌套结构体  relations 为子数据集的字段名称
func BatchSaveMutiStructOnDuplicate[T MutiStruct](records *[]T, relations *[]string, batchSize int) error {
	mainDatas := make([]interface{}, 0)
	subDatas := make(map[string][]interface{})
	for _, r := range *records {
		newR := r.SetStringID().(T)
		id := newR.GetStringID()
		elemType := reflect.TypeOf(newR)
		mainDatas = append(mainDatas, newR)
		for _, fieldName := range *relations {
			if field, ok := elemType.FieldByName(fieldName); ok {
				fieldIndex := field.Index[0]
				err := extractSubDatas(fieldName, id, &newR, fieldIndex, subDatas)
				if err != nil {
					return err
				}
			}
		}

		if len(mainDatas) >= batchSize {
			if err := saveBatchMutiData(&mainDatas, subDatas); err != nil {
				return err
			}
		}
	}

	if len(mainDatas) > 0 {
		if err := saveBatchMutiData(&mainDatas, subDatas); err != nil {
			return err
		}
	}
	return nil
}

// extractSubDatas 嵌套结构体 提取子数据集
func extractSubDatas[T MutiStruct](fieldName, id string, item *T, fieldIndex int, subDatas map[string][]interface{}) error {
	val := reflect.ValueOf(*item)
	fieldVal := val.Field(fieldIndex)
	var subTableName string
	if IsArray(fieldVal.Interface()) {
		for i := 0; i < fieldVal.Len(); i++ {
			elem := fieldVal.Index(i).Interface()
			if setter, ok := elem.(SubStruct); ok {
				setter = setter.SetRelationID(id).(SubStruct)
				subTableName = setter.TableName()
				fieldVal.Index(i).Set(reflect.ValueOf(setter))
				if _, exists := subDatas[subTableName]; !exists {
					subDatas[subTableName] = make([]interface{}, 0)
				}
				subDatas[subTableName] = append(subDatas[subTableName], setter)
			} else {
				return fmt.Errorf("%s element at index %d in slice is not SubStruct", fieldName, i)
			}
		}
	} else {
		if subEnt, ok := fieldVal.Interface().(SubStruct); ok {
			subEnt = subEnt.SetRelationID(id).(SubStruct)
			subTableName = subEnt.TableName()
			if _, exists := subDatas[subTableName]; !exists {
				subDatas[subTableName] = make([]interface{}, 0, 1)
			} else {
				return fmt.Errorf("Not SubStruct Interface")
			}
			subDatas[subTableName] = append(subDatas[subTableName], subEnt)
		} else {
			return fmt.Errorf("field %s is not a slice not SubStruct", val.Type().Field(fieldIndex))
		}
	}
	return nil
}

// saveBatchMutiData 保存嵌套结构体数据
func saveBatchMutiData(mainDatas *[]interface{}, subDatas map[string][]interface{}) error {
	IdArr := make([]string, 0)
	for _, md := range *mainDatas {
		if m, e := md.(MutiStruct); e {
			IdArr = append(IdArr, m.GetStringID())
		}
	}
	db := GetDb()
	tx := db.Begin()
	tx, err := InsertBatchOnDuplicate(mainDatas, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, subData := range subDatas {
		if len(subData) > 0 {
			first := subData[0]
			if fi, es := first.(SubStruct); es {
				fi.DeleteRecord(IdArr, tx)
			}
			tx, err = InsertBatchOnDuplicate(&subData, tx)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	tx.Commit()
	for relationKey, subData := range subDatas {
		if len(subData) > 0 {
			subDatas[relationKey] = subDatas[relationKey][:0]
		}
	}
	*mainDatas = make([]interface{}, 0, cap(*mainDatas))
	return nil
}

func QuerySql(sqlStr string, params []interface{}) (returnData []map[string]string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	var rows *sql.Rows
	rows, err = GetDb().Raw(sqlStr, params...).Rows()
	if err != nil {
		return
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {

		}
	}(rows)
	var columns []string
	columns, err = rows.Columns()
	if err != nil {
		return
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)
	for rows.Next() {
		values := make([]string, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err = rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		entry := make(map[string]string)
		for i, col := range columns {
			val := values[i]
			if val == "" {
				entry[col] = ""
				continue
			}
			entry[col] = val
		}

		returnData = append(returnData, entry)
	}
	return
}

// 获取与给定 JSON 标签相匹配的结构体字段名
func getJSONTagField(t reflect.Type, jsonTag string) string {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == jsonTag || tag == "-" {
			return field.Name
		}
		if parts := structsUtils.SplitJsonTag(tag); len(parts) > 0 && parts[0] == jsonTag {
			return field.Name
		}
	}
	return ""
}

// 解析 gorm 标签字符串为键值对
func parseGormTag(tag string) []map[string]string {
	var result []map[string]string
	parts := splitTag(tag)

	for _, part := range parts {
		if part == "" {
			continue
		}
		m := make(map[string]string)
		keyValue := splitKeyValue(part)
		if len(keyValue) == 2 {
			m[keyValue[0]] = keyValue[1]
		} else {
			m[keyValue[0]] = ""
		}
		result = append(result, m)
	}

	return result
}
