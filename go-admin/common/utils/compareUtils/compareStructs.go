package compareUtils

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

const (
	// 忽略字段
	formatValue = "2006-01-02 15:04:05"
)

// ChangeRecord 记录单个字段的变更信息
type ChangeRecord struct {
	Field     string      `json:"field"` // 字段名
	FieldName string      `json:"fieldName"`
	OldValue  interface{} `json:"oldValue"` // 变更前值
	NewValue  interface{} `json:"newValue"` // 变更后值
}

// Compare 比较两个结构体对象，返回变更记录的 JSON 格式
func Compare(oldObj, newObj interface{}) (json.RawMessage, error) {
	oldVal := reflect.ValueOf(oldObj)
	newVal := reflect.ValueOf(newObj)
	//如果是指针，解引用
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}
	// 确保两个对象类型一致
	if oldVal.Type() != newVal.Type() {
		panic("objects must be of the same type")
	}
	var changes []ChangeRecord
	// 遍历结构体字段
	for i := 0; i < oldVal.NumField(); i++ {
		field := oldVal.Type().Field(i)
		oldFieldValue := oldVal.Field(i).Interface()
		newFieldValue := newVal.Field(i).Interface()

		// 获取 compare 标签
		compare := field.Tag.Get("compare")
		dict := field.Tag.Get("dict")

		// 处理字典映射
		if dict != "" {
			mapping := ParseDictMapping(dict)
			if desc, exists := mapping[oldFieldValue.(string)]; exists {
				oldFieldValue = desc
			}
			if desc, exists := mapping[newFieldValue.(string)]; exists {
				newFieldValue = desc
			}
		}

		// 特殊处理 time.Time 类型字段
		if fieldType := field.Type; fieldType == reflect.TypeOf(time.Time{}) {
			oldFieldValue = FormatTime(oldFieldValue.(time.Time))
			newFieldValue = FormatTime(newFieldValue.(time.Time))
		}
		// 比较字段值是否不同
		if !reflect.DeepEqual(oldFieldValue, newFieldValue) && compare != "" {
			changes = append(changes, ChangeRecord{
				Field:     field.Name,
				FieldName: compare,
				OldValue:  oldFieldValue,
				NewValue:  newFieldValue,
			})
		}
	}
	// 序列化为 JSON
	result, err := json.Marshal(changes)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ParseDictMapping 解析字段的字典映射规则
func ParseDictMapping(dict string) map[string]string {
	mapping := make(map[string]string)
	parts := strings.Split(dict, "|")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			mapping[kv[0]] = kv[1]
		}
	}
	return mapping
}

// FormatTime 格式化时间字段
func FormatTime(t time.Time) string {
	return t.Format(formatValue)
}
