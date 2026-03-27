package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// json字符串转map
func JsonStringToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return map[string]interface{}{}, nil
	}
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// json字符串转map
func JsonStringToArray(jsonStr string) ([]interface{}, error) {
	if jsonStr == "" {
		return []interface{}{}, nil
	}
	var result []interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func InTypeArray[T comparable](value T, slice []T) bool {
	if len(slice) == 0 {
		return false
	}
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// 判断是否是数组
func IsArray(v interface{}) bool {
	if v == nil {
		return false
	}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false // 避免对 nil 指针调用 Elem()
		}
		val = val.Elem()
	}
	kind := val.Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

// 判断是否是map
func IsMap(v interface{}) bool {
	if v == nil {
		return false
	}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false // 避免对 nil 指针调用 Elem()
		}
		val = val.Elem()
	}
	return val.Kind() == reflect.Map
}

// 检查 headField 是否包含 requireField 中的所有元素，使用 map 提高查找效率
func ContainsAllRequired(headField, requireField []string) (isPass bool, missField []string) {
	isPass = true
	headMap := make(map[string]bool)
	for _, head := range headField {
		headMap[head] = true
	}

	for _, req := range requireField {
		if !headMap[req] {
			isPass = false
			missField = append(missField, req)
		}
	}
	return
}

func MergeMaps(mergeTo *map[string]interface{}, data map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}, []interface{}:
			continue
		default:
			(*mergeTo)[key] = v
		}
	}
}

// CamelToSnake 将驼峰命名法转换为下划线命名法
func CamelToSnake(name string) string {
	var b strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(rune(name[i-1])) || (i+1 < len(name) && unicode.IsLower(rune(name[i+1])))) {
				b.WriteRune('_')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// 驼峰转下划线
func ConvertKeysToUnderscore(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	re := regexp.MustCompile(`(?m)([a-z])([A-Z])`)

	for key, value := range m {
		newKey := re.ReplaceAllStringFunc(key, func(match string) string {
			return match[:1] + "_" + strings.ToLower(match[1:])
		})
		newKey = strings.ToLower(newKey)

		if v, ok := value.(map[string]interface{}); ok {
			result[newKey] = ConvertKeysToUnderscore(v)
		} else {
			result[newKey] = value
		}
	}

	return result
}

// MapColumnValues 获取 map 中指定字段的值数组
func MapColumnValues[T string | int | int64 | float64](data []map[string]T, field string) []T {
	var result []T
	for _, row := range data {
		result = append(result, row[field])
	}
	return result
}

// ArraySum 获取 map 中指定字段的值数组
func ArraySum[T int | int64 | float64](data []map[string]T, field string) T {
	var sum T = 0
	for _, row := range data {
		sum += row[field]
	}
	return sum
}

func ToFloat64(i interface{}) float64 {
	if i == nil {
		return 0
	}
	switch v := i.(type) {
	case nil:
		return 0
	case float64:
		return v
	case float32:
		return float64(v)
	case int, int8, int16, int32, int64:
		return float64(v.(int))
	default:
		f, _ := strconv.ParseFloat(v.(string), 64)
		return f
	}
}

func ToInt(value interface{}) int {
	if value == nil {
		return 0
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(v.Int())
	case reflect.Float64, reflect.Float32:
		return int(v.Float())
	case reflect.String:
		parsedInt, err := strconv.Atoi(v.String())
		if err != nil {
			return 0
		}
		return parsedInt
	default:
		return 0
	}
}

func ToString(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10), nil
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10), nil
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("cannot convert %T to string", value)
	}
}

// splitTag 将 gorm 标签字符串拆分为多个部分
func splitTag(tag string) []string {
	return strings.FieldsFunc(tag, func(r rune) bool {
		return r == ';'
	})
}

// splitKeyValue 将键值对字符串拆分为键和值
func splitKeyValue(kv string) []string {
	parts := strings.SplitN(kv, ":", 2)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func ArrayUnique[T comparable](slice []T) []T {
	keys := make(map[T]struct{})
	j := 0
	for _, v := range slice {
		if _, exists := keys[v]; !exists {
			keys[v] = struct{}{}
			slice[j] = v
			j++
		}
	}
	return slice[:j]
}

// ArrayDiff 返回 src 中存在但不在其他 slices 中出现的元素
func ArrayDiff[T comparable](src []T, slices ...[]T) []T {
	if len(src) == 0 {
		return []T{}
	}

	// 构建一个 map 来记录所有其他 slice 中出现的元素
	excluded := make(map[T]bool)
	for _, slice := range slices {
		for _, item := range slice {
			excluded[item] = true
		}
	}

	// 遍历 src，保留不在 excluded 中的元素
	var result []T
	for _, item := range src {
		if !excluded[item] {
			result = append(result, item)
		}
	}

	return result
}

// PageSlice 对 slice 进行分页
func PageSlice[T any](slice []T, pageNumber, pageSize int, totalPages *int) []T {
	returnData := make([]T, 0)
	if len(slice) == 0 || pageNumber < 1 || pageSize < 1 {
		return returnData
	}

	// 计算总页数
	if *totalPages == 0 {
		*totalPages = int(math.Ceil(float64(len(slice)) / float64(pageSize)))
	}

	// 检查请求的页码是否超出范围
	if pageNumber > *totalPages {
		return returnData
	}

	// 计算起始和结束索引
	startIndex := (pageNumber - 1) * pageSize
	endIndex := startIndex + pageSize

	// 确保 endIndex 不超过 slice 的长度
	if endIndex > len(slice) {
		endIndex = len(slice)
	}

	// 返回分页后的子切片
	return slice[startIndex:endIndex]
}

// CompareStructFields 比对两个结构体相同字段的值
// commentKey true 返回值以s2的comment为Key false 返回值以字段名为key
func CompareStructFields(s1, s2 interface{}, commentKey bool) map[string][]interface{} {
	valMap := make(map[string][]interface{})
	v1 := reflect.ValueOf(s1)
	t1 := v1.Type()
	v2 := reflect.ValueOf(s2)
	t2 := v2.Type()

	if t1.Kind() != reflect.Struct || t2.Kind() != reflect.Struct {
		fmt.Println("One of the inputs is not a struct")
		return valMap
	}

	for i := 0; i < v2.NumField(); i++ {
		fieldName := t1.Field(i).Name
		field1 := v1.Field(i)
		comment := t2.Field(i).Tag.Get("comment")
		field2Value := v2.FieldByName(fieldName)
		var mapKey string
		if commentKey {
			mapKey = fieldName
		} else {
			mapKey = comment
		}
		if field2Value.IsValid() {
			if (field1.Kind() == field2Value.Kind()) && !reflect.DeepEqual(field1.Interface(), field2Value.Interface()) {
				valMap[mapKey] = []interface{}{field1.Interface(), field2Value.Interface()}
			}
		}
	}
	return valMap
}

func CompareFields(a, b interface{}) []string {
	var result []string
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if va.Kind() != reflect.Struct || vb.Kind() != reflect.Struct {
		return result // 确保传入的是结构体类型
	}

	// 获取a的结构体类型，用于获取字段名和值
	ta := va.Type()
	for i := 0; i < ta.NumField(); i++ {
		fieldA := va.Field(i)
		fieldB := vb.FieldByName(ta.Field(i).Name)                                                               // 从b中获取相同名字的字段
		if !fieldB.IsValid() || slices.Contains([]string{"Model", "ModelTime", "ControlBy"}, ta.Field(i).Name) { // 如果b中没有这个字段，跳过比较
			continue
		}
		if !reflect.DeepEqual(fieldA.Interface(), fieldB.Interface()) { // 比较值是否相等
			result = append(result, fmt.Sprintf("变更了【%v】由【%v】更新为【%v】", extractComment(ta.Field(i).Tag.Get("gorm")), fieldA.Interface(), fieldB.Interface()))
		}
	}

	return result
}

// extractComment 从 gorm tag 中提取 comment
func extractComment(gormTag string) string {
	if gormTag == "" {
		return ""
	}

	// 分割 tag，处理多个属性
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "comment:") {
			// 移除 "comment:" 前缀
			comment := strings.TrimPrefix(part, "comment:")
			return strings.TrimSpace(comment)
		}
	}

	return ""
}

// StructToJsonString 结构体转成JsonString
func StructToJsonString(a interface{}) string {
	jsonString, err := json.Marshal(a)
	if err != nil {
		return ""
	}
	return string(jsonString)
}

// RemoveValueFromSlice 从切片中删除指定元素
func RemoveValueFromSlice[T string | int | int64 | float32 | float64](slice []T, value T) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}

func ByteSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	n := math.Floor(math.Log(float64(bytes)) / math.Log(1024))
	val := float64(bytes) / math.Pow(1024, n)
	return fmt.Sprintf("%.2f %s", val, units[int(n)])
}
